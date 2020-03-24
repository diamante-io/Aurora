package ingest

import (
	"time"

	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	herr "github.com/hcnet/go/services/aurora/internal/errors"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/support/errors"
	ilog "github.com/hcnet/go/support/log"
)

// Backfill ingests history in reverse chronological order, from the current
// aurora elder query for `n` ledgers
func (i *System) Backfill(n uint) error {
	start := ledger.CurrentState().HistoryElder - 1
	end := start - int32(n) + 1
	is := NewSession(i)
	is.Cursor = NewCursor(start, end, i)

	log.WithField("start", start).
		WithField("end", end).
		WithField("err", is.Err).
		WithField("ingested", is.Ingested).
		Info("ingest: backfill start")

	is.Run()

	log.WithField("start", start).
		WithField("end", end).
		WithField("err", is.Err).
		WithField("ingested", is.Ingested).
		Info("ingest: backfill complete")

	return is.Err
}

// ClearAll removes all previously ingested historical data from the aurora
// database.
func (i *System) ClearAll() error {
	hdb := i.AuroraDB.Clone()
	ingestion := &Ingestion{DB: hdb}

	err := ingestion.Start()
	if err != nil {
		return errors.Wrap(err, "failed to begin ingestion")
	}

	err = ingestion.ClearAll()
	if err != nil {
		return errors.Wrap(err, "failed to clear history tables")
	}

	err = ingestion.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close ingestion")
	}

	log.Infof("cleared all history")

	return nil
}

// RebaseHistory re-establishes aurora's history database by clearing it, ingesting the latest ledger in hcnet-core then backfilling as many ledgers as possible
func (i *System) RebaseHistory() error {
	var latest int32
	var elder int32

	q := core.Q{Session: i.CoreDB}
	err := q.LatestLedger(&latest)
	if err != nil {
		return errors.Wrap(err, "load core latest ledger failed")
	}

	err = q.ElderLedger(&elder)
	if err != nil {
		return errors.Wrap(err, "load core elder ledger failed")
	}

	err = i.ClearAll()
	if err != nil {
		return errors.Wrap(err, "failed to  clear db")
	}

	log.Infof("rebasing history using ledgers %d-%d", elder, latest)

	_, err = i.ReingestRange(latest, elder)
	if err != nil {
		return errors.Wrap(err, "failed to ingest latest ledger segment")
	}

	return nil
}

// ReingestAll re-ingests all ledgers
func (i *System) ReingestAll() (int, error) {

	var elder int32
	var latest int32
	q := history.Q{Session: i.AuroraDB}

	err := q.ElderLedger(&elder)
	if err != nil {
		return 0, errors.Wrap(err, "load history elder ledger failed")
	}

	err = q.LatestLedger(&latest)
	if err != nil {
		return 0, errors.Wrap(err, "load history latest ledger failed")
	}

	log.
		WithField("start", latest).
		WithField("end", elder).
		Info("reingest: all")

	return i.ReingestRange(latest, elder)
}

// ReingestOutdated finds old ledgers and reimports them.
func (i *System) ReingestOutdated() (n int, err error) {
	q := history.Q{Session: i.AuroraDB}

	// NOTE: this loop will never terminate if some bug were cause a ledger
	// reingestion to silently fail.
	for {
		outdated := []int32{}
		err = q.OldestOutdatedLedgers(&outdated, CurrentVersion)
		if err != nil {
			return
		}

		if len(outdated) == 0 {
			return
		}

		log.
			WithField("lowest_sequence", outdated[0]).
			WithField("batch_size", len(outdated)).
			Info("reingest: outdated")

		var start, end int32
		flush := func() error {
			ingested, ferr := i.ReingestRange(start, end)

			if ferr != nil {
				return ferr
			}
			n += ingested
			return nil
		}

		for idx := range outdated {
			seq := outdated[idx]

			if start == 0 {
				start = seq
				end = seq
				continue
			}

			if seq == end+1 {
				end = seq
				continue
			}

			err = flush()
			if err != nil {
				return
			}

			start = seq
			end = seq
		}

		err = flush()
		if err != nil {
			return
		}
	}
}

// ReingestRange reingests a range of ledgers, from `start` to `end`, inclusive.
func (i *System) ReingestRange(start, end int32) (int, error) {
	is := NewSession(i)
	is.Cursor = NewCursor(start, end, i)
	is.ClearExisting = true

	is.Run()
	log.WithField("start", start).
		WithField("end", end).
		WithField("err", is.Err).
		WithField("ingested", is.Ingested).
		Info("ingest: range complete")
	return is.Ingested, is.Err
}

// ReingestSingle re-ingests a single ledger
func (i *System) ReingestSingle(sequence int32) error {
	_, err := i.ReingestRange(sequence, sequence)
	return err
}

// Tick triggers the ingestion system to ingest any new ledger data, provided
// that there currently is not an import session in progress.
func (i *System) Tick() *Session {
	i.lock.Lock()
	if i.current != nil {
		log.Info("ingest: already in progress")
		i.lock.Unlock()
		return nil
	}

	is := NewSession(i)
	i.current = is
	i.lock.Unlock()

	i.runOnce()
	return is
}

// run causes the importer to check hcnet-core to see if we can import new
// data.
func (i *System) runOnce() {
	defer func() {
		if rec := recover(); rec != nil {
			err := herr.FromPanic(rec)
			log.Errorf("import session panicked: %s", err)
			herr.ReportToSentry(err, nil)
		}
	}()

	// 1. stash a copy of the current ingestion session (assigned from the tick)
	// 2. decide what to import
	// 3. import until none available

	// 1.
	i.lock.Lock()
	is := i.current
	i.lock.Unlock()

	defer func() {
		i.lock.Lock()
		i.current = nil
		i.lock.Unlock()
	}()

	// Warning: do not check the current ledger state using ledger.CurrentState()! It is updated
	// in another go routine and can return the same data for two different ingestion sessions.
	var coreLatest, historyLatest int32

	coreQ := core.Q{Session: i.CoreDB}
	err := coreQ.LatestLedger(&coreLatest)
	if err != nil {
		log.WithFields(ilog.F{"err": err}).Error("Error getting core latest ledger")
		return
	}

	historyQ := history.Q{Session: i.AuroraDB}
	err = historyQ.LatestLedger(&historyLatest)
	if err != nil {
		log.WithFields(ilog.F{"err": err}).Error("Error getting history latest ledger")
		return
	}

	if is == nil {
		log.Warn("ingest: runOnce ran with a nil current session")
		return
	}

	if coreLatest == 1 {
		log.Warn("ingest: waiting for hcnet-core sync")
		return
	}

	if historyLatest == coreLatest {
		log.Debug("ingest: no new ledgers")
		return
	}

	// 2.
	if historyLatest == 0 {
		log.Infof(
			"history db is empty, establishing base at ledger %d",
			coreLatest,
		)
		is.Cursor = NewCursor(coreLatest, coreLatest, i)
	} else {
		is.Cursor = NewCursor(historyLatest+1, coreLatest, i)
	}

	// 3.
	logFields := ilog.F{
		"first_ledger": is.Cursor.FirstLedger,
		"last_ledger":  is.Cursor.LastLedger,
	}
	log.WithFields(logFields).Info("Ingesting ledgers...")
	ingestStart := time.Now()

	is.Run()

	if is.Err != nil {
		// We need to use `Error` method because `is.Err` is `withMessage` struct from
		// `github.com/pkg/errors` and encodes to `{}` in the logs.
		logFields["err"] = is.Err.Error()
		log.WithFields(logFields).Error("Error ingesting ledgers")
		return
	}

	logFields["duration"] = time.Since(ingestStart).Seconds()
	log.WithFields(logFields).Info("Finished ingesting ledgers")
}
