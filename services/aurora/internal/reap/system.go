package reap

import (
	"time"

	"github.com/hcnet/go/services/aurora/internal/errors"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/toid"
	"github.com/hcnet/go/support/log"
)

// DeleteUnretainedHistory removes all data associated with unretained ledgers.
func (r *System) DeleteUnretainedHistory() error {
	// RetentionCount of 0 indicates "keep all history"
	if r.RetentionCount == 0 {
		return nil
	}

	var (
		latest      = ledger.CurrentState()
		targetElder = (latest.HistoryLatest - int32(r.RetentionCount)) + 1
	)

	if targetElder < latest.HistoryElder {
		return nil
	}

	err := r.clearBefore(targetElder)
	if err != nil {
		return err
	}

	log.
		WithField("new_elder", targetElder).
		Info("reaper succeeded")

	return nil
}

// Tick triggers the reaper system to update itself, deleted unretained history
// if it is the appropriate time.
func (r *System) Tick() {
	if time.Now().Before(r.nextRun) {
		return
	}

	r.runOnce()
	r.nextRun = time.Now().Add(1 * time.Hour)
}

func (r *System) runOnce() {
	defer func() {
		if rec := recover(); rec != nil {
			err := errors.FromPanic(rec)
			log.Errorf("reaper panicked: %s", err)
			errors.ReportToSentry(err, nil)
		}
	}()

	err := r.DeleteUnretainedHistory()
	if err != nil {
		log.Errorf("reaper failed: %s", err)
	}
}

func (r *System) clearBefore(seq int32) error {
	log.WithField("new_elder", seq).Info("reaper: clearing")

	clear := r.AuroraDB.DeleteRange
	end := toid.New(seq, 0, 0).ToInt64()

	err := clear(0, end, "history_effects", "history_operation_id")
	if err != nil {
		return err
	}
	err = clear(0, end, "history_operation_participants", "history_operation_id")
	if err != nil {
		return err
	}
	err = clear(0, end, "history_operations", "id")
	if err != nil {
		return err
	}
	err = clear(0, end, "history_transaction_participants", "history_transaction_id")
	if err != nil {
		return err
	}
	err = clear(0, end, "history_transactions", "id")
	if err != nil {
		return err
	}
	err = clear(0, end, "history_ledgers", "id")
	if err != nil {
		return err
	}

	return nil
}
