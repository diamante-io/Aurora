package ledgerbackend

import (
	"context"
	"encoding/hex"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/diamnet/go/historyarchive"
	"github.com/diamnet/go/support/log"
	"github.com/diamnet/go/xdr"
)

// Ensure CaptiveDiamnetCore implements LedgerBackend
var _ LedgerBackend = (*CaptiveDiamnetCore)(nil)

func (c *CaptiveDiamnetCore) roundDownToFirstReplayAfterCheckpointStart(ledger uint32) uint32 {
	r := c.checkpointManager.GetCheckpointRange(ledger)
	if r.Low <= 1 {
		// Diamnet-Core doesn't stream ledger 1
		return 2
	}
	// All other checkpoints start at the next multiple of 64
	return r.Low
}

// CaptiveDiamnetCore is a ledger backend that starts internal Diamnet-Core
// subprocess responsible for streaming ledger data. It provides better decoupling
// than DatabaseBackend but requires some extra init time.
//
// It operates in two modes:
//   * When a BoundedRange is prepared it starts Diamnet-Core in catchup mode that
//     replays ledgers in memory. This is very fast but requires Diamnet-Core to
//     keep ledger state in RAM. It requires around 3GB of RAM as of August 2020.
//   * When a UnboundedRange is prepared it runs Diamnet-Core catchup mode to
//     sync with the first ledger and then runs it in a normal mode. This
//     requires the configAppendPath to be provided because a quorum set needs to
//     be selected.
//
// When running CaptiveDiamnetCore will create a temporary folder to store
// bucket files and other temporary files. The folder is removed when Close is
// called.
//
// The communication is performed via filesystem pipe which is created in a
// temporary folder.
//
// Currently BoundedRange requires a full-trust on history archive. This issue is
// being fixed in Diamnet-Core.
//
// While using BoundedRanges is straightforward there are a few gotchas connected
// to UnboundedRanges:
//   * PrepareRange takes more time because all ledger entries must be stored on
//     disk instead of RAM.
//   * If GetLedger is not called frequently (every 5 sec. on average) the
//     Diamnet-Core process can go out of sync with the network. This happens
//     because there is no buffering of communication pipe and CaptiveDiamnetCore
//     has a very small internal buffer and Diamnet-Core will not close the new
//     ledger if it's not read.
//
// Except for the Close function, CaptiveDiamnetCore is not thread-safe and should
// not be accessed by multiple go routines. Close is thread-safe and can be called
// from another go routine. Once Close is called it will interrupt and cancel any
// pending operations.
//
// Requires Diamnet-Core v13.2.0+.
type CaptiveDiamnetCore struct {
	archive           historyarchive.ArchiveInterface
	checkpointManager historyarchive.CheckpointManager
	ledgerHashStore   TrustedLedgerHashStore

	// cancel is the CancelFunc for context which controls the lifetime of a CaptiveDiamnetCore instance.
	// Once it is invoked CaptiveDiamnetCore will not be able to stream ledgers from Diamnet Core or
	// spawn new instances of Diamnet Core.
	cancel context.CancelFunc

	diamnetCoreRunner diamnetCoreRunnerInterface
	// diamnetCoreLock protects access to diamnetCoreRunner. When the read lock
	// is acquired diamnetCoreRunner can be accessed. When the write lock is acquired
	// diamnetCoreRunner can be updated.
	diamnetCoreLock sync.RWMutex

	// For testing
	diamnetCoreRunnerFactory func(mode diamnetCoreRunnerMode) (diamnetCoreRunnerInterface, error)

	// cachedMeta keeps that ledger data of the last fetched ledger. Updated in GetLedger().
	cachedMeta *xdr.LedgerCloseMeta

	prepared           *Range  // non-nil if any range is prepared
	nextLedger         uint32  // next ledger expected, error w/ restart if not seen
	lastLedger         *uint32 // end of current segment if offline, nil if online
	previousLedgerHash *string
}

// CaptiveCoreConfig contains all the parameters required to create a CaptiveDiamnetCore instance
type CaptiveCoreConfig struct {
	// BinaryPath is the file path to the Diamnet Core binary
	BinaryPath string
	// NetworkPassphrase is the Diamnet network passphrase used by captive core when connecting to the Diamnet network
	NetworkPassphrase string
	// HistoryArchiveURLs are a list of history archive urls
	HistoryArchiveURLs []string
	Toml               *CaptiveCoreToml

	// Optional fields

	// CheckpointFrequency is the number of ledgers between checkpoints
	// if unset, DefaultCheckpointFrequency will be used
	CheckpointFrequency uint32
	// LedgerHashStore is an optional store used to obtain hashes for ledger sequences from a trusted source
	LedgerHashStore TrustedLedgerHashStore
	// Log is an (optional) custom logger which will capture any output from the Diamnet Core process.
	// If Log is omitted then all output will be printed to stdout.
	Log *log.Entry
	// Context is the (optional) context which controls the lifetime of a CaptiveDiamnetCore instance. Once the context is done
	// the CaptiveDiamnetCore instance will not be able to stream ledgers from Diamnet Core or spawn new
	// instances of Diamnet Core. If Context is omitted CaptiveDiamnetCore will default to using context.Background.
	Context context.Context
	// StoragePath is the (optional) base path passed along to Core's
	// BUCKET_DIR_PATH which specifies where various bucket data should be
	// stored. We always append /captive-core to this directory, since we clean
	// it up entirely on shutdown.
	StoragePath string
}

// NewCaptive returns a new CaptiveDiamnetCore instance.
func NewCaptive(config CaptiveCoreConfig) (*CaptiveDiamnetCore, error) {
	// Here we set defaults in the config. Because config is not a pointer this code should
	// not mutate the original CaptiveCoreConfig instance which was passed into NewCaptive()

	// Log Captive Core straight to stdout by default
	if config.Log == nil {
		config.Log = log.New()
		config.Log.SetOutput(os.Stdout)
		config.Log.SetLevel(logrus.InfoLevel)
	}

	parentCtx := config.Context
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	var cancel context.CancelFunc
	config.Context, cancel = context.WithCancel(parentCtx)

	archivePool, err := historyarchive.NewArchivePool(
		config.HistoryArchiveURLs,
		historyarchive.ConnectOptions{
			NetworkPassphrase:   config.NetworkPassphrase,
			CheckpointFrequency: config.CheckpointFrequency,
			Context:             config.Context,
		},
	)

	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "Error connecting to ALL history archives.")
	}

	c := &CaptiveDiamnetCore{
		archive:           &archivePool,
		ledgerHashStore:   config.LedgerHashStore,
		cancel:            cancel,
		checkpointManager: historyarchive.NewCheckpointManager(config.CheckpointFrequency),
	}

	c.diamnetCoreRunnerFactory = func(mode diamnetCoreRunnerMode) (diamnetCoreRunnerInterface, error) {
		return newDiamnetCoreRunner(config, mode)
	}
	return c, nil
}

func (c *CaptiveDiamnetCore) getLatestCheckpointSequence() (uint32, error) {
	has, err := c.archive.GetRootHAS()
	if err != nil {
		return 0, errors.Wrap(err, "error getting root HAS")
	}

	return has.CurrentLedger, nil
}

func (c *CaptiveDiamnetCore) openOfflineReplaySubprocess(from, to uint32) error {
	latestCheckpointSequence, err := c.getLatestCheckpointSequence()
	if err != nil {
		return errors.Wrap(err, "error getting latest checkpoint sequence")
	}

	if from > latestCheckpointSequence {
		return errors.Errorf(
			"from sequence: %d is greater than max available in history archives: %d",
			from,
			latestCheckpointSequence,
		)
	}

	if to > latestCheckpointSequence {
		return errors.Errorf(
			"to sequence: %d is greater than max available in history archives: %d",
			to,
			latestCheckpointSequence,
		)
	}

	var runner diamnetCoreRunnerInterface
	if runner, err = c.diamnetCoreRunnerFactory(diamnetCoreRunnerModeOffline); err != nil {
		return errors.Wrap(err, "error creating diamnet-core runner")
	} else {
		// only assign c.diamnetCoreRunner if runner is not nil to avoid nil interface check
		// see https://golang.org/doc/faq#nil_error
		c.diamnetCoreRunner = runner
	}

	err = c.diamnetCoreRunner.catchup(from, to)
	if err != nil {
		return errors.Wrap(err, "error running diamnet-core")
	}

	// The next ledger should be the first ledger of the checkpoint containing
	// the requested ledger
	ran := BoundedRange(from, to)
	c.prepared = &ran
	c.nextLedger = c.roundDownToFirstReplayAfterCheckpointStart(from)
	c.lastLedger = &to
	c.previousLedgerHash = nil

	return nil
}

func (c *CaptiveDiamnetCore) openOnlineReplaySubprocess(ctx context.Context, from uint32) error {
	latestCheckpointSequence, err := c.getLatestCheckpointSequence()
	if err != nil {
		return errors.Wrap(err, "error getting latest checkpoint sequence")
	}

	// We don't allow starting the online mode starting with more than two
	// checkpoints from now. Such requests are likely buggy.
	// We should allow only one checkpoint here but sometimes there are up to a
	// minute delays when updating root HAS by diamnet-core.
	twoCheckPointsLength := (c.checkpointManager.GetCheckpoint(0) + 1) * 2
	maxLedger := latestCheckpointSequence + twoCheckPointsLength
	if from > maxLedger {
		return errors.Errorf(
			"trying to start online mode too far (latest checkpoint=%d), only two checkpoints in the future allowed",
			latestCheckpointSequence,
		)
	}

	var runner diamnetCoreRunnerInterface
	if runner, err = c.diamnetCoreRunnerFactory(diamnetCoreRunnerModeOnline); err != nil {
		return errors.Wrap(err, "error creating diamnet-core runner")
	} else {
		// only assign c.diamnetCoreRunner if runner is not nil to avoid nil interface check
		// see https://golang.org/doc/faq#nil_error
		c.diamnetCoreRunner = runner
	}

	runFrom, ledgerHash, err := c.runFromParams(ctx, from)
	if err != nil {
		return errors.Wrap(err, "error calculating ledger and hash for diamnet-core run")
	}

	err = c.diamnetCoreRunner.runFrom(runFrom, ledgerHash)
	if err != nil {
		return errors.Wrap(err, "error running diamnet-core")
	}

	// In the online mode we update nextLedger after streaming the first ledger.
	// This is to support versions before and after/including v17.1.0 that
	// introduced minimal persistent DB.
	c.nextLedger = 0
	ran := UnboundedRange(from)
	c.prepared = &ran
	c.lastLedger = nil
	c.previousLedgerHash = nil

	return nil
}

// runFromParams receives a ledger sequence and calculates the required values to call diamnet-core run with --start-ledger and --start-hash
func (c *CaptiveDiamnetCore) runFromParams(ctx context.Context, from uint32) (runFrom uint32, ledgerHash string, err error) {
	if from == 1 {
		// Trying to start-from 1 results in an error from Diamnet-Core:
		// Target ledger 1 is not newer than last closed ledger 1 - nothing to do
		// TODO maybe we can fix it by generating 1st ledger meta
		// like GenesisLedgerStateReader?
		err = errors.New("CaptiveCore is unable to start from ledger 1, start from ledger 2")
		return
	}

	if from <= 63 {
		// The line below is to support a special case for streaming ledger 2
		// that works for all other ledgers <= 63 (fast-forward).
		// We can't set from=2 because Diamnet-Core will not allow starting from 1.
		// To solve this we start from 3 and exploit the fast that Diamnet-Core
		// will stream data from 2 for the first checkpoint.
		from = 3
	}

	runFrom = from - 1
	if c.ledgerHashStore != nil {
		var exists bool
		ledgerHash, exists, err = c.ledgerHashStore.GetLedgerHash(ctx, runFrom)
		if err != nil {
			err = errors.Wrapf(err, "error trying to read ledger hash %d", runFrom)
			return
		}
		if exists {
			return
		}
	}

	ledgerHeader, err2 := c.archive.GetLedgerHeader(from)
	if err2 != nil {
		err = errors.Wrapf(err2, "error trying to read ledger header %d from HAS", from)
		return
	}
	ledgerHash = hex.EncodeToString(ledgerHeader.Header.PreviousLedgerHash[:])
	return
}

// nextExpectedSequence returns nextLedger (if currently set) or start of
// prepared range. Otherwise it returns 0.
// This is done because `nextLedger` is 0 between the moment Diamnet-Core is
// started and streaming the first ledger (in such case we return first ledger
// in requested range).
func (c *CaptiveDiamnetCore) nextExpectedSequence() uint32 {
	if c.nextLedger == 0 && c.prepared != nil {
		return c.prepared.from
	}
	return c.nextLedger
}

func (c *CaptiveDiamnetCore) startPreparingRange(ctx context.Context, ledgerRange Range) (bool, error) {
	c.diamnetCoreLock.Lock()
	defer c.diamnetCoreLock.Unlock()

	if c.isPrepared(ledgerRange) {
		return true, nil
	}

	if c.diamnetCoreRunner != nil {
		if err := c.diamnetCoreRunner.close(); err != nil {
			return false, errors.Wrap(err, "error closing existing session")
		}

		// Make sure Diamnet-Core is terminated before starting a new instance.
		processExited, _ := c.diamnetCoreRunner.getProcessExitError()
		if !processExited {
			return false, errors.New("the previous Diamnet-Core instance is still running")
		}
	}

	var err error
	if ledgerRange.bounded {
		err = c.openOfflineReplaySubprocess(ledgerRange.from, ledgerRange.to)
	} else {
		err = c.openOnlineReplaySubprocess(ctx, ledgerRange.from)
	}
	if err != nil {
		return false, errors.Wrap(err, "opening subprocess")
	}

	return false, nil
}

// PrepareRange prepares the given range (including from and to) to be loaded.
// Captive diamnet-core backend needs to initalize Diamnet-Core state to be
// able to stream ledgers.
// Diamnet-Core mode depends on the provided ledgerRange:
//   * For BoundedRange it will start Diamnet-Core in catchup mode.
//   * For UnboundedRange it will first catchup to starting ledger and then run
//     it normally (including connecting to the Diamnet network).
// Please note that using a BoundedRange, currently, requires a full-trust on
// history archive. This issue is being fixed in Diamnet-Core.
func (c *CaptiveDiamnetCore) PrepareRange(ctx context.Context, ledgerRange Range) error {
	if alreadyPrepared, err := c.startPreparingRange(ctx, ledgerRange); err != nil {
		return errors.Wrap(err, "error starting prepare range")
	} else if alreadyPrepared {
		return nil
	}

	_, err := c.GetLedger(ctx, ledgerRange.from)
	if err != nil {
		return errors.Wrapf(err, "Error fast-forwarding to %d", ledgerRange.from)
	}

	return nil
}

// IsPrepared returns true if a given ledgerRange is prepared.
func (c *CaptiveDiamnetCore) IsPrepared(ctx context.Context, ledgerRange Range) (bool, error) {
	c.diamnetCoreLock.RLock()
	defer c.diamnetCoreLock.RUnlock()

	return c.isPrepared(ledgerRange), nil
}

func (c *CaptiveDiamnetCore) isPrepared(ledgerRange Range) bool {
	if c.isClosed() {
		return false
	}

	lastLedger := uint32(0)
	if c.lastLedger != nil {
		lastLedger = *c.lastLedger
	}

	cachedLedger := uint32(0)
	if c.cachedMeta != nil {
		cachedLedger = c.cachedMeta.LedgerSequence()
	}

	if c.prepared == nil {
		return false
	}

	if lastLedger == 0 {
		return c.nextExpectedSequence() <= ledgerRange.from || cachedLedger == ledgerRange.from
	}

	// From now on: lastLedger != 0 so current range is bounded

	if ledgerRange.bounded {
		return (c.nextExpectedSequence() <= ledgerRange.from || cachedLedger == ledgerRange.from) &&
			lastLedger >= ledgerRange.to
	}

	// Requested range is unbounded but current one is bounded
	return false
}

// GetLedger will block until the ledger is available in the backend
// (even for UnboundedRange), then return it's LedgerCloseMeta.
//
// Call PrepareRange first to instruct the backend which ledgers to fetch.
// CaptiveDiamnetCore requires PrepareRange call first to initialize Diamnet-Core.
// Requesting a ledger on non-prepared backend will return an error.
//
// Please note that requesting a ledger sequence far after current
// ledger will block the execution for a long time.
//
// Because ledger data is streamed from Diamnet-Core sequentially, users should
// request sequences in a non-decreasing order. If the requested sequence number
// is less than the last requested sequence number, an error will be returned.
//
// This function behaves differently for bounded and unbounded ranges:
//   * BoundedRange: After getting the last ledger in a range this method will
//     also Close() the backend.
func (c *CaptiveDiamnetCore) GetLedger(ctx context.Context, sequence uint32) (xdr.LedgerCloseMeta, error) {
	c.diamnetCoreLock.RLock()
	defer c.diamnetCoreLock.RUnlock()

	if c.cachedMeta != nil && sequence == c.cachedMeta.LedgerSequence() {
		// GetLedger can be called multiple times using the same sequence, ex. to create
		// change and transaction readers. If we have this ledger buffered, let's return it.
		return *c.cachedMeta, nil
	}

	if c.isClosed() {
		return xdr.LedgerCloseMeta{}, errors.New("session is closed, call PrepareRange first")
	}

	if sequence < c.nextExpectedSequence() {
		return xdr.LedgerCloseMeta{}, errors.Errorf(
			"requested ledger %d is behind the captive core stream (expected=%d)",
			sequence,
			c.nextExpectedSequence(),
		)
	}

	if c.lastLedger != nil && sequence > *c.lastLedger {
		return xdr.LedgerCloseMeta{}, errors.Errorf(
			"reading past bounded range (requested sequence=%d, last ledger in range=%d)",
			sequence,
			*c.lastLedger,
		)
	}

	// Now loop along the range until we find the ledger we want.
	for {
		select {
		case <-ctx.Done():
			return xdr.LedgerCloseMeta{}, ctx.Err()
		case result, ok := <-c.diamnetCoreRunner.getMetaPipe():
			found, ledger, err := c.handleMetaPipeResult(sequence, result, ok)
			if found || err != nil {
				return ledger, err
			}
		}
	}
}

func (c *CaptiveDiamnetCore) handleMetaPipeResult(sequence uint32, result metaResult, ok bool) (bool, xdr.LedgerCloseMeta, error) {
	if err := c.checkMetaPipeResult(result, ok); err != nil {
		c.diamnetCoreRunner.close()
		return false, xdr.LedgerCloseMeta{}, err
	}

	seq := result.LedgerCloseMeta.LedgerSequence()
	// If we got something unexpected; close and reset
	if c.nextLedger != 0 && seq != c.nextLedger {
		c.diamnetCoreRunner.close()
		return false, xdr.LedgerCloseMeta{}, errors.Errorf(
			"unexpected ledger sequence (expected=%d actual=%d)",
			c.nextLedger,
			seq,
		)
	} else if c.nextLedger == 0 && seq > c.prepared.from {
		// First stream ledger is greater than prepared.from
		c.diamnetCoreRunner.close()
		return false, xdr.LedgerCloseMeta{}, errors.Errorf(
			"unexpected ledger sequence (expected=<=%d actual=%d)",
			c.prepared.from,
			seq,
		)
	}

	newPreviousLedgerHash := result.LedgerCloseMeta.PreviousLedgerHash().HexString()
	if c.previousLedgerHash != nil && *c.previousLedgerHash != newPreviousLedgerHash {
		// We got something unexpected; close and reset
		c.diamnetCoreRunner.close()
		return false, xdr.LedgerCloseMeta{}, errors.Errorf(
			"unexpected previous ledger hash for ledger %d (expected=%s actual=%s)",
			seq,
			*c.previousLedgerHash,
			newPreviousLedgerHash,
		)
	}

	c.nextLedger = result.LedgerSequence() + 1
	currentLedgerHash := result.LedgerCloseMeta.LedgerHash().HexString()
	c.previousLedgerHash = &currentLedgerHash

	// Update cache with the latest value because we incremented nextLedger.
	c.cachedMeta = result.LedgerCloseMeta

	if seq == sequence {
		// If we got the _last_ ledger in a segment, close before returning.
		if c.lastLedger != nil && *c.lastLedger == seq {
			if err := c.diamnetCoreRunner.close(); err != nil {
				return false, xdr.LedgerCloseMeta{}, errors.Wrap(err, "error closing session")
			}
		}
		return true, *c.cachedMeta, nil
	}

	return false, xdr.LedgerCloseMeta{}, nil
}

func (c *CaptiveDiamnetCore) checkMetaPipeResult(result metaResult, ok bool) error {
	// There are 3 types of errors we check for:
	// 1. User initiated shutdown by canceling the parent context or calling Close().
	// 2. The diamnet core process exited unexpectedly.
	// 3. Some error was encountered while consuming the ledgers emitted by captive core (e.g. parsing invalid xdr)
	if err := c.diamnetCoreRunner.context().Err(); err != nil {
		// Case 1 - User initiated shutdown by canceling the parent context or calling Close()
		return err
	}
	if !ok || result.err != nil {
		if result.err != nil {
			// Case 3 - Some error was encountered while consuming the ledger stream emitted by captive core.
			return result.err
		} else if exited, err := c.diamnetCoreRunner.getProcessExitError(); exited {
			// Case 2 - The diamnet core process exited unexpectedly
			if err == nil {
				return errors.Errorf("diamnet core exited unexpectedly")
			} else {
				return errors.Wrap(err, "diamnet core exited unexpectedly")
			}
		} else if !ok {
			// This case should never happen because the ledger buffer channel can only be closed
			// if and only if the process exits or the context is cancelled.
			// However, we add this check for the sake of completeness
			return errors.Errorf("meta pipe closed unexpectedly")
		}
	}
	return nil
}

// GetLatestLedgerSequence returns the sequence of the latest ledger available
// in the backend. This method returns an error if not in a session (start with
// PrepareRange).
//
// Note that for UnboundedRange the returned sequence number is not necessarily
// the latest sequence closed by the network. It's always the last value available
// in the backend.
func (c *CaptiveDiamnetCore) GetLatestLedgerSequence(ctx context.Context) (uint32, error) {
	c.diamnetCoreLock.RLock()
	defer c.diamnetCoreLock.RUnlock()

	if c.isClosed() {
		return 0, errors.New("diamnet-core must be opened to return latest available sequence")
	}

	if c.lastLedger == nil {
		return c.nextExpectedSequence() - 1 + uint32(len(c.diamnetCoreRunner.getMetaPipe())), nil
	}
	return *c.lastLedger, nil
}

func (c *CaptiveDiamnetCore) isClosed() bool {
	return c.prepared == nil || c.diamnetCoreRunner == nil || c.diamnetCoreRunner.context().Err() != nil
}

// Close closes existing Diamnet-Core process, streaming sessions and removes all
// temporary files. Note, once a CaptiveDiamnetCore instance is closed it can can no longer be used and
// all subsequent calls to PrepareRange(), GetLedger(), etc will fail.
// Close is thread-safe and can be called from another go routine.
func (c *CaptiveDiamnetCore) Close() error {
	c.diamnetCoreLock.RLock()
	defer c.diamnetCoreLock.RUnlock()

	// after the CaptiveDiamnetCore context is canceled all subsequent calls to PrepareRange() will fail
	c.cancel()

	// TODO: Sucks to ignore the error here, but no worse than it was before,
	// so...
	if c.ledgerHashStore != nil {
		c.ledgerHashStore.Close()
	}

	if c.diamnetCoreRunner != nil {
		return c.diamnetCoreRunner.close()
	}

	return nil
}
