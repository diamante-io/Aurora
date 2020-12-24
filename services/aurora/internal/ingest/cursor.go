package ingest

import (
	"time"

	"github.com/diamnet/go/meta"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/toid"
	"github.com/diamnet/go/xdr"
)

// BeforeAndAfter loads the ledger entry for `target` before the current
// operation was applied and after the operation was applied.
func (c *Cursor) BeforeAndAfter(target xdr.LedgerKey) (
	before *xdr.LedgerEntry,
	after *xdr.LedgerEntry,
	err error,
) {
	m := c.TransactionMetaBundle()
	before, err = m.StateBefore(target, c.op)
	if err != nil {
		return
	}

	after, err = m.StateAfter(target, c.op)
	if err != nil {
		return
	}

	return
}

// InLedger returns true if the cursor is on a ledger.
func (c *Cursor) InLedger() bool {
	return c.lg != 0
}

// InOperation returns true if the cursor is on a operation. Will return false
// after advancing to a new transaction but before advancing on to the
// transaciton's first operation.
func (c *Cursor) InOperation() bool {
	return c.InLedger() && c.op != -1
}

// InTransaction returns true if the cursor is pointing to a transaction.  This
// will return false after advancing to a new ledger but prior to advancing into
// the ledger's first transaction.
func (c *Cursor) InTransaction() bool {
	return c.InLedger() && c.tx != -1
}

// Ledger returns the current ledger
func (c *Cursor) Ledger() *core.LedgerHeader {
	return &c.data.Header
}

// LedgerID returns the current ledger's id, as used by the history system.
func (c *Cursor) LedgerID() int64 {
	return toid.New(c.lg, 0, 0).ToInt64()
}

// LedgerRange returns the beginning and end of id values that map to the
// current ledger.  Useful for clearing a ledgers worth of data.
func (c *Cursor) LedgerRange() (start int64, end int64) {
	if c.lg == 1 {
		start = 0
	} else {
		start = toid.New(c.lg, 0, 0).ToInt64()
	}

	return start, toid.New(c.lg+1, 0, 0).ToInt64()
}

// LedgerSequence returns the current ledger's sequence
func (c *Cursor) LedgerSequence() int32 {
	return c.data.Sequence
}

// NextLedger advances `c` to the next ledger in the iteration, loading a new
// LedgerBundle from the core database. Returns false if an error occurs or
// the iteration is complete.
func (c *Cursor) NextLedger() bool {
	if c.Err != nil {
		return false
	}

	if !c.incrementLg() {
		return false
	}

	c.data = &LedgerBundle{Sequence: c.lg}
	start := time.Now()
	c.Err = c.data.Load(c.CoreDB)
	if c.Err != nil {
		return false
	}

	if c.Metrics != nil {
		c.Metrics.LoadLedgerTimer.Update(time.Since(start))
	}

	c.tx = -1
	c.op = -1

	return true
}

func (c *Cursor) incrementLg() bool {
	isReverse := c.FirstLedger > c.LastLedger

	if c.lg == 0 {
		c.lg = c.FirstLedger
	} else {
		increment := int32(1)
		if isReverse {
			increment = int32(-1)
		}
		c.lg += increment
	}

	// If we're complete, reset the cursor and finish the iteration.  Complete in
	// this context means the current ledger (c.lg) has progressed beyond the
	// LastLedger: If iterating forward the current ledger is greater than
	// LastLedger, if reverse the current ledger is less than LastLedger.
	if (!isReverse && c.lg > c.LastLedger) || (isReverse && c.lg < c.LastLedger) {
		c.data = nil
		c.lg = 0
		return false
	}

	return true
}

// NextOp advances `c` to the next operation in the current transaction.  Returns
// false if the current transaction has nothing left to visit.
func (c *Cursor) NextOp() bool {
	if c.Err != nil {
		return false
	}
	c.op++
	return c.op < len(c.Operations())
}

// NextTx advances `c` to the next transaction in the current ledger.  Returns
// false if the current ledger has no transactions left to visit.
func (c *Cursor) NextTx() bool {
	if c.Err != nil {
		return false
	}
	c.tx++
	c.op = -1
	return c.tx < len(c.data.Transactions)
}

// Operation returns the current operation
func (c *Cursor) Operation() *xdr.Operation {
	return &c.data.Transactions[c.tx].Envelope.Tx.Operations[c.op]
}

// OperationChanges returns all of LedgerEntryChanges that occurred in the
// course of applying the current operation.
func (c *Cursor) OperationChanges() xdr.LedgerEntryChanges {
	return c.TransactionMetaBundle().OperationsMetas()[c.op].Changes
}

// OperationCount returns the count of operations in the current transaction
func (c *Cursor) OperationCount() int {
	return len(c.data.Transactions[c.tx].Envelope.Tx.Operations)
}

// OperationID returns the current operations id, as used by the history system.
func (c *Cursor) OperationID() int64 {
	return toid.New(c.lg, int32(c.tx+1), int32(c.op+1)).ToInt64()
}

// OperationOrder returns the order of the current operation amongst the
// current transaction's operations.
func (c *Cursor) OperationOrder() int32 {
	return int32(c.op + 1)
}

// OperationResult returns the current operation's result record
func (c *Cursor) OperationResult() *xdr.OperationResultTr {
	txr := &c.data.Transactions[c.tx].Result.Result
	tr := txr.Result.MustResults()[c.op].MustTr()
	return &tr
}

// OperationSourceAccount returns the current operation's effective source
// account (i.e. default's to the transaction's source account).
func (c *Cursor) OperationSourceAccount() xdr.AccountId {
	aid := c.Operation().SourceAccount
	if aid != nil {
		return *aid
	}

	return c.TransactionSourceAccount()
}

// OperationType returns the current operation type
func (c *Cursor) OperationType() xdr.OperationType {
	return c.Operation().Body.Type
}

// Operations returns the current transactions operations.
func (c *Cursor) Operations() []xdr.Operation {
	return c.data.Transactions[c.tx].Envelope.Tx.Operations
}

// Transaction returns the current transaction
func (c *Cursor) Transaction() *core.Transaction {
	return &c.data.Transactions[c.tx]
}

// TransactionFee returns the txfeehistory row for the current
// transaction.
func (c *Cursor) TransactionFee() *core.TransactionFee {
	return &c.data.TransactionFees[c.tx]
}

// TransactionMetaBundle provides easier access to the meta data regarding
// the application of the current transaction.
func (c *Cursor) TransactionMetaBundle() *meta.Bundle {
	return &meta.Bundle{
		FeeMeta:         c.TransactionFee().Changes,
		TransactionMeta: c.Transaction().ResultMeta,
	}
}

// SuccessfulLedgerOperationCount returns the count of operations in the current ledger
func (c *Cursor) SuccessfulLedgerOperationCount() (ret int) {
	for i := range c.data.Transactions {
		if !c.data.Transactions[i].IsSuccessful() {
			continue
		}
		ret += len(c.data.Transactions[i].Envelope.Tx.Operations)
	}
	return
}

// SuccessfulTransactionCount returns the count of transactions in the current
// ledger that succeeded.
func (c *Cursor) SuccessfulTransactionCount() (ret int) {
	for i := range c.data.Transactions {
		if c.data.Transactions[i].IsSuccessful() {
			ret++
		}
	}
	return
}

// FailedTransactionCount returns the count of transactions in the current
// ledger that failed.
func (c *Cursor) FailedTransactionCount() (ret int) {
	for i := range c.data.Transactions {
		if !c.data.Transactions[i].IsSuccessful() {
			ret++
		}
	}
	return
}

// TransactionID returns the current tranaction's id, as used by the history
// system.
func (c *Cursor) TransactionID() int64 {
	return toid.New(c.lg, int32(c.tx+1), 0).ToInt64()
}

// TransactionSourceAccount returns the current transaction's source account id
func (c *Cursor) TransactionSourceAccount() xdr.AccountId {
	return c.Transaction().Envelope.Tx.SourceAccount
}
