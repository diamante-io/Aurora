package expingest

import (
	"time"

	ilog "github.com/diamnet/go/support/log"
)

// LoggingStateReporter logs the progress of a session running its
// state pipelines
type LoggingStateReporter struct {
	Log      *ilog.Entry
	Interval int

	entryCount int
	sequence   uint32
	startTime  time.Time
}

// OnStartState logs that the session has started reading from the history archive snapshot
func (lr *LoggingStateReporter) OnStartState(sequence uint32) {
	lr.Log.WithField("sequence", sequence).Info("Reading from History Archive Snapshot")

	lr.entryCount = 0
	lr.sequence = sequence
	lr.startTime = time.Now()
}

// OnStateEntry logs that the session has processed an entry from the history archive snapshot
func (lr *LoggingStateReporter) OnStateEntry() {
	lr.entryCount++
	if lr.entryCount%lr.Interval == 0 {
		lr.Log.WithField("sequence", lr.sequence).
			WithField("numEntries", lr.entryCount).
			Info("Processing entries from History Archive Snapshot")
	}
}

// OnEndState logs that the session has finished processing the history archive snapshot
func (lr *LoggingStateReporter) OnEndState(err error, shutdown bool) {
	elapsedTime := time.Since(lr.startTime)

	l := lr.Log.WithField("sequence", lr.sequence).
		WithField("numEntries", lr.entryCount).
		WithField("shutdown", shutdown).
		WithField("duration", elapsedTime.Seconds())

	if err != nil {
		l.WithError(err).Error("Error processing History Archive Snapshot")
	} else {
		l.Info("Finished processing History Archive Snapshot")
	}
}

// LoggingLedgerReporter logs the progress of a session running its
// ledger pipelines
type LoggingLedgerReporter struct {
	Log *ilog.Entry

	transactionCount int
	sequence         uint32
	startTime        time.Time
}

// OnNewLedger logs that the session has started reading a new ledger
func (lr *LoggingLedgerReporter) OnNewLedger(sequence uint32) {
	lr.Log.WithField("sequence", sequence).Info("Reading new ledger")
	lr.transactionCount = 0
	lr.sequence = sequence
	lr.startTime = time.Now()
}

// OnLedgerTransaction records that the session has processed a transaction from the ledger
func (lr *LoggingLedgerReporter) OnLedgerTransaction() {
	lr.transactionCount++
}

// OnEndLedger logs that the session has finished processing the ledger
func (lr *LoggingLedgerReporter) OnEndLedger(err error, shutdown bool) {
	elapsedTime := time.Since(lr.startTime)

	l := lr.Log.WithField("sequence", lr.sequence).
		WithField("transactions", lr.transactionCount).
		WithField("shutdown", shutdown).
		WithField("duration", elapsedTime.Seconds())

	if err != nil {
		l.WithError(err).Error("Error processing ledger")
	} else {
		l.Info("Finished processing ledger")
	}
}
