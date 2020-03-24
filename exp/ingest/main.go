package ingest

import (
	"sync"

	"github.com/hcnet/go/clients/hcnetcore"
	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/exp/ingest/ledgerbackend"
	"github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/support/historyarchive"
	"github.com/hcnet/go/xdr"
)

// standardSession contains common methods used by all sessions.
type standardSession struct {
	shutdown                          chan bool
	rwLock                            sync.RWMutex
	latestSuccessfullyProcessedLedger uint32

	runningMutex sync.Mutex
	running      bool
}

// LiveSession initializes the ledger state using `Archive` and `statePipeline`,
// then starts processing ledger data using `LedgerBackend` and `ledgerPipeline`.
type LiveSession struct {
	standardSession

	Archive           historyarchive.ArchiveInterface
	LedgerBackend     ledgerbackend.LedgerBackend
	HcNetCoreClient *hcnetcore.Client
	// HcNetCoreCursor defines cursor name used in `setcursor` command of
	// hcnet-core. If you run multiple sessions against a single hcnet-core
	// instance the cursor name needs to be different in each session.
	HcNetCoreCursor string
	StatePipeline     *pipeline.StatePipeline
	StateReporter     StateReporter
	LedgerPipeline    *pipeline.LedgerPipeline
	LedgerReporter    LedgerReporter
	// TempSet is a store used to hold temporary objects generated during
	// state processing. If nil, defaults to io.MemoryTempSet.
	TempSet io.TempSet
}

// SingleLedgerSession initializes the ledger state using `Archive` and `StatePipeline`
// and terminates. Useful for finding stats for a single ledger. Set `LedgerSequence`
// to `0` to process the latest checkpoint.
type SingleLedgerSession struct {
	standardSession

	Archive        *historyarchive.Archive
	LedgerSequence uint32
	StatePipeline  *pipeline.StatePipeline
	StateReporter  StateReporter
	// TempSet is a store used to hold temporary objects generated during
	// state processing. If nil, defaults to io.MemoryTempSet.
	TempSet io.TempSet
}

// Session is an implementation of a ingesting scenario. Some useful sessions
// can be found in this package.
type Session interface {
	// Run start the session and works as long as the session is active. If
	// you want to resume a session use Resume().
	Run() error
	// Resume resumes the session at ledger with a given sequence. It's up to
	// Session user to determine what was the last ledger processed by a
	// Session as it's stateless (or if Run() should be called first).
	Resume(ledgerSequence uint32) error
	// GetLatestSuccessfullyProcessedLedger returns the ledger sequence of the
	// latest successfully processed ledger in the session. Return 0 if no
	// ledgers were processed yet.
	// Please note that this value is not synchronized with pipeline hooks.
	GetLatestSuccessfullyProcessedLedger() uint32
	// Shutdown gracefully stops running session and stops all internal
	// objects.
	// Calling Shutdown() does not trigger post processing hooks.
	Shutdown()
	// QueryLock locks the session for sending a query. This ensures a consistent
	// view when data is saved to multiple stores (ex. postgres/redis/memory).
	// TODO this is fine for a demo but we need to check if it works for systems
	// that don't provide strong consistency. This may also slow down readers
	// if commiting data to stores take longer time. This probably doesn't work
	// in distributed apps.
	QueryLock()
	// QueryUnlock unlocks read lock of the session.
	QueryUnlock()
	// UpdateLock locks the session for updating data. This ensures a consistent
	// view when data is saved to multiple stores (ex. postgres/redis/memory).
	// TODO this is fine for a demo but we need to check if it works for systems
	// that don't provide strong consistency. This may also slow down readers
	// if commiting data to stores take longer time. This probably doesn't work
	// in distributed apps.
	UpdateLock()
	// UpdateUnlock unlocks write lock of the session.
	UpdateUnlock()
}

// StateReporter can be used by a session to log progress
// or update metrics as the session runs its state pipelines.
type StateReporter interface {
	// OnStartState is called when the session begins processing
	// a history archive snapshot at the given sequence number.
	OnStartState(sequence uint32)
	// OnStateEntry is called when the session processes an entry
	// from the io.StateReader
	OnStateEntry()
	// OnEndState is called when the session finishes processing
	// a history archive snapshot.
	// if err is not nil it means that the session stoped processing the
	// history archive snapshot because of an error.
	// if shutdown is true it means the session stopped processing the
	// history archive snapshot because it received a shutdown signal
	OnEndState(err error, shutdown bool)
}

// reporterLedgerReader instruments a io.StateReader with a StateReporter
// which reports each xdr.LedgerEntryChange which is read from the reader
type reporterStateReader struct {
	io.StateReader
	StateReporter
}

func (r reporterStateReader) Read() (xdr.LedgerEntryChange, error) {
	entry, err := r.StateReader.Read()
	if err == nil {
		r.OnStateEntry()
	}

	return entry, err
}

// LedgerReporter can be used by a session to log progress
// or update metrics as the session runs its ledger pipelines.
type LedgerReporter interface {
	// OnNewLedger is called when the session begins processing
	// a ledger at the given sequence number.
	OnNewLedger(sequence uint32)
	// OnLedgerEntry is called when the session processes a transaction
	// from the io.LedgerReader
	OnLedgerTransaction()
	// OnEndLedger is called when the session finishes processing
	// a ledger.
	// if err is not nil it means that the session stoped processing the
	// ledger because of an error.
	// if shutdown is true it means the session stopped processing the
	// ledger because it received a shutdown signal
	OnEndLedger(err error, shutdown bool)
}

// reporterLedgerReader instruments a io.LedgerReader with a LedgerReporter
// which reports each io.LedgerTransaction which is read from the reader
type reporterLedgerReader struct {
	io.LedgerReader
	LedgerReporter
}

func (r reporterLedgerReader) Read() (io.LedgerTransaction, error) {
	entry, err := r.LedgerReader.Read()
	if err == nil {
		r.OnLedgerTransaction()
	}

	return entry, err
}
