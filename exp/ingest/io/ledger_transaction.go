package io

import (
	"github.com/hcnet/go/xdr"
)

// Change is a developer friendly representation of LedgerEntryChanges.
// It also provides some helper functions to quickly check if a given
// change has occured in an entry.
//
// If an entry is created: Pre is nil and Post is not nil.
// If an entry is updated: Pre is not nil and Post is not nil.
// If an entry is removed: Pre is not nil and Post is nil.
type Change struct {
	Type xdr.LedgerEntryType
	Pre  *xdr.LedgerEntryData
	Post *xdr.LedgerEntryData
}

// AccountSignersChanged returns true if account signers have changed.
// Notice: this will return true on master key changes too!
func (c *Change) AccountSignersChanged() bool {
	if c.Type != xdr.LedgerEntryTypeAccount {
		panic("This should not be called on changes other than Account changes")
	}

	// New account so new master key (which is also a signer)
	if c.Pre == nil {
		return true
	}

	// Account merged, check if master key weight > 0, if so, it means
	// it got removed. Merged account can't have subentries so we don't
	// need to check other signers.
	// c.Pre != nil at this point.
	if c.Post == nil {
		preAccountEntry := c.Pre.MustAccount()
		return preAccountEntry.MasterKeyWeight() > 0
	}

	// c.Pre != nil && c.Post != nil at this point.
	preAccountEntry := c.Pre.MustAccount()
	postAccountEntry := c.Post.MustAccount()

	preSigners := preAccountEntry.SignerSummary()
	postSigners := postAccountEntry.SignerSummary()

	if len(preSigners) != len(postSigners) {
		return true
	}

	for postSigner, postWeight := range postSigners {
		preWeight, exist := preSigners[postSigner]
		if !exist {
			return true
		}

		if preWeight != postWeight {
			return true
		}
	}

	return false
}

// GetChanges returns a developer friendly representation of LedgerEntryChanges.
// It contains fee changes, transaction changes and operation changes in that
// order.
func (t *LedgerTransaction) GetChanges() []Change {
	// Fee meta
	changes := getChangesFromLedgerEntryChanges(t.FeeChanges)

	// Transaction meta
	v1Meta, ok := t.Meta.GetV1()
	if ok {
		txChanges := getChangesFromLedgerEntryChanges(v1Meta.TxChanges)
		changes = append(changes, txChanges...)
	}

	// Operation meta
	for _, operationMeta := range t.Meta.OperationsMeta() {
		ledgerEntryChanges := operationMeta.Changes
		opChanges := getChangesFromLedgerEntryChanges(ledgerEntryChanges)

		changes = append(changes, opChanges...)
	}

	return changes
}

// getChangesFromLedgerEntryChanges transforms LedgerEntryChanges to []Change.
// Each `update` and `removed` is preceded with `state` and `create` changes
// are alone, without `state`. The transformation we're doing is to move each
// change (state/update, state/removed or create) to an array of pre/post pairs.
// Then:
// - for create, pre is null and post is a new entry,
// - for update, pre is previous state and post is the current state,
// - for removed, pre is previous state and post is null.
//
// hcnet-core source:
// https://github.com/hcnet/hcnet-core/blob/e584b43/src/ledger/LedgerTxn.cpp#L582
func getChangesFromLedgerEntryChanges(ledgerEntryChanges xdr.LedgerEntryChanges) []Change {
	changes := []Change{}

	for i, entryChange := range ledgerEntryChanges {
		switch entryChange.Type {
		case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
			created := entryChange.MustCreated()
			changes = append(changes, Change{
				Type: created.Data.Type,
				Pre:  nil,
				Post: &created.Data,
			})
		case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
			state := ledgerEntryChanges[i-1].MustState()
			updated := entryChange.MustUpdated()
			changes = append(changes, Change{
				Type: state.Data.Type,
				Pre:  &state.Data,
				Post: &updated.Data,
			})
		case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
			state := ledgerEntryChanges[i-1].MustState()
			changes = append(changes, Change{
				Type: state.Data.Type,
				Pre:  &state.Data,
				Post: nil,
			})
		case xdr.LedgerEntryChangeTypeLedgerEntryState:
			continue
		default:
			panic("Invalid LedgerEntryChangeType")
		}
	}

	return changes
}
