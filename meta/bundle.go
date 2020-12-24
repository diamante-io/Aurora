package meta

import (
	"errors"
	"fmt"
	"math"

	"github.com/diamnet/go/xdr"
)

// ErrMetaNotFound is returned when no meta that matches a provided filter can
// be found.
var ErrMetaNotFound = errors.New("meta: no changes found")

// InitialState returns the initial state of the LedgerEntry identified by `key`
// just prior to the application of the transaction the produced `b`.  Returns
// nil if the ledger entry did not exist prior to the bundle.
func (b *Bundle) InitialState(key xdr.LedgerKey) (*xdr.LedgerEntry, error) {
	all := b.Changes(key)

	if len(all) == 0 {
		return nil, ErrMetaNotFound
	}

	first := all[0]

	if first.Type != xdr.LedgerEntryChangeTypeLedgerEntryState {
		return nil, nil
	}

	result := first.MustState()

	return &result, nil
}

// Changes returns any changes within the bundle that apply to the entry
// identified by `key`.
func (b *Bundle) Changes(target xdr.LedgerKey) (ret []xdr.LedgerEntryChange) {
	return b.changes(target, math.MaxInt32)
}

// StateAfter returns the state of entry `key` after the application of the
// operation at `opidx`
func (b *Bundle) StateAfter(key xdr.LedgerKey, opidx int) (*xdr.LedgerEntry, error) {
	all := b.changes(key, opidx)

	if len(all) == 0 {
		return nil, ErrMetaNotFound
	}

	change := all[len(all)-1]

	switch change.Type {
	case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
		entry := change.MustCreated()
		return &entry, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
		return nil, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
		entry := change.MustUpdated()
		return &entry, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryState:
		// scott: diamnet-core should not emit a lone state entry, and we are
		// retrieving changes from the end of the collection.  If this situation
		// occurs, it means that I didn't understand something correctly or there is
		// a bug in diamnet-core.
		panic(fmt.Errorf("Unexpected 'state' entry"))
	default:
		panic(fmt.Errorf("Unknown change type: %v", change.Type))
	}
}

// StateBefore returns the state of entry `key` just prior to the application of
// the operation at `opidx`
func (b *Bundle) StateBefore(key xdr.LedgerKey, opidx int) (*xdr.LedgerEntry, error) {
	all := b.changes(key, opidx)

	if len(all) == 0 {
		return nil, ErrMetaNotFound
	}

	// If we only found one entry, that means it didn't exist prior to this
	// operation
	if len(all) == 1 {
		return nil, nil
	}

	change := all[len(all)-2]

	switch change.Type {
	case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
		entry := change.MustCreated()
		return &entry, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
		return nil, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
		entry := change.MustUpdated()
		return &entry, nil
	case xdr.LedgerEntryChangeTypeLedgerEntryState:
		entry := change.MustState()
		return &entry, nil
	default:
		panic(fmt.Errorf("Unknown change type: %v", change.Type))
	}
}

//OperationMetas retrieves all operation metas from a transaction bundle
func (b *Bundle) OperationsMetas() []xdr.OperationMeta {
	switch b.TransactionMeta.V {
	case 0:
		return b.TransactionMeta.MustOperations()
	default:
		return b.TransactionMeta.V1.Operations
	}
}

//filterChanges takes a LedgerEntryChange slice and filters out changes that don't match the given target ledger key
func filterChanges(changes []xdr.LedgerEntryChange, target xdr.LedgerKey) (filteredChanges []xdr.LedgerEntryChange) {
	for _, change := range changes {
		key := change.LedgerKey()

		if !key.Equals(target) {
			continue
		}

		filteredChanges = append(filteredChanges, change)
	}
	return filteredChanges
}

// changes returns any changes within the bundle that apply to the entry
// identified by `key` that occurred at or before `maxOp`.
func (b *Bundle) changes(target xdr.LedgerKey, maxOp int) []xdr.LedgerEntryChange {

	//allChanges accumulates all ledger changes
	allChanges := b.FeeMeta

	if b.TransactionMeta.V > 0 {
		allChanges = append(allChanges, b.TransactionMeta.V1.TxChanges...)
	}

	for i, op := range b.OperationsMetas() {
		if i > maxOp {
			break
		}
		allChanges = append(allChanges, op.Changes...)
	}

	return filterChanges(allChanges, target)
}
