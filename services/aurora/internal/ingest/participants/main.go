// Package participants contains functions to derive a set of "participant"
// addresses for various data structures in the HcNet network's ledger.
package participants

import (
	"fmt"

	"github.com/hcnet/go/xdr"
)

// ForOperation returns all the participating accounts from the
// provided operation.
func ForOperation(
	tx *xdr.Transaction,
	op *xdr.Operation,
) (result []xdr.AccountId, err error) {

	if op.SourceAccount != nil {
		result = append(result, *op.SourceAccount)
	} else {
		result = append(result, tx.SourceAccount)
	}

	switch op.Body.Type {
	case xdr.OperationTypeCreateAccount:
		result = append(result, op.Body.MustCreateAccountOp().Destination)
	case xdr.OperationTypePayment:
		result = append(result, op.Body.MustPaymentOp().Destination)
	case xdr.OperationTypePathPayment:
		result = append(result, op.Body.MustPathPaymentOp().Destination)
	case xdr.OperationTypeManageBuyOffer:
		// the only direct participant is the source_account
	case xdr.OperationTypeManageSellOffer:
		// the only direct participant is the source_account
	case xdr.OperationTypeCreatePassiveSellOffer:
		// the only direct participant is the source_account
	case xdr.OperationTypeSetOptions:
		// the only direct participant is the source_account
	case xdr.OperationTypeChangeTrust:
		// the only direct participant is the source_account
	case xdr.OperationTypeAllowTrust:
		result = append(result, op.Body.MustAllowTrustOp().Trustor)
	case xdr.OperationTypeAccountMerge:
		result = append(result, op.Body.MustDestination())
	case xdr.OperationTypeInflation:
		// the only direct participant is the source_account
	case xdr.OperationTypeManageData:
		// the only direct participant is the source_account
	case xdr.OperationTypeBumpSequence:
		// the only direct participant is the source_account
	default:
		err = fmt.Errorf("Unknown operation type: %s", op.Body.Type)
	}

	result = dedupe(result)
	return
}

// ForTransaction returns all the participating accounts from the provided
// transaction.
func ForTransaction(
	tx *xdr.Transaction,
	meta *xdr.TransactionMeta,
	feeMeta *xdr.LedgerEntryChanges,
) (result []xdr.AccountId, err error) {

	result = append(result, tx.SourceAccount)

	p, err := forMeta(meta)
	if err != nil {
		return
	}
	result = append(result, p...)

	p, err = forChanges(feeMeta)
	if err != nil {
		return
	}
	result = append(result, p...)

	for i := range tx.Operations {
		p, err = ForOperation(tx, &tx.Operations[i])
		if err != nil {
			return
		}
		result = append(result, p...)
	}

	result = dedupe(result)
	return
}

// dedupe remove any duplicate ids from `in`
func dedupe(in []xdr.AccountId) (out []xdr.AccountId) {
	set := map[string]xdr.AccountId{}
	for _, id := range in {
		set[id.Address()] = id
	}

	for _, id := range set {
		out = append(out, id)
	}
	return
}

func forChanges(
	changes *xdr.LedgerEntryChanges,
) (result []xdr.AccountId, err error) {

	for _, c := range *changes {
		var account *xdr.AccountId

		switch c.Type {
		case xdr.LedgerEntryChangeTypeLedgerEntryCreated:
			account = forLedgerEntry(c.MustCreated())
		case xdr.LedgerEntryChangeTypeLedgerEntryRemoved:
			account = forLedgerKey(c.MustRemoved())
		case xdr.LedgerEntryChangeTypeLedgerEntryUpdated:
			account = forLedgerEntry(c.MustUpdated())
		case xdr.LedgerEntryChangeTypeLedgerEntryState:
			account = forLedgerEntry(c.MustState())
		default:
			err = fmt.Errorf("Unknown change type: %s", c.Type)
			return
		}

		if account != nil {
			result = append(result, *account)
		}
	}

	return
}

func forLedgerEntry(le xdr.LedgerEntry) *xdr.AccountId {
	if le.Data.Type != xdr.LedgerEntryTypeAccount {
		return nil
	}
	aid := le.Data.MustAccount().AccountId
	return &aid
}

func forLedgerKey(lk xdr.LedgerKey) *xdr.AccountId {
	if lk.Type != xdr.LedgerEntryTypeAccount {
		return nil
	}
	aid := lk.MustAccount().AccountId
	return &aid
}

func forMeta(
	meta *xdr.TransactionMeta,
) (result []xdr.AccountId, err error) {

	if meta.Operations == nil {
		return
	}

	for _, op := range *meta.Operations {
		var acc []xdr.AccountId
		acc, err = forChanges(&op.Changes)
		if err != nil {
			return
		}

		result = append(result, acc...)
	}

	return
}
