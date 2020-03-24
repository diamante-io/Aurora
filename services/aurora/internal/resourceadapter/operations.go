package resourceadapter

import (
	"context"
	"fmt"

	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/protocols/aurora/operations"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/support/render/hal"
	"github.com/hcnet/go/xdr"
)

// NewOperation creates a new operation resource, finding the appropriate type to use
// based upon the row's type.
func NewOperation(
	ctx context.Context,
	operationRow history.Operation,
	transactionRow *history.Transaction,
	ledger history.Ledger,
) (result hal.Pageable, err error) {

	base := operations.Base{}
	PopulateBaseOperation(ctx, &base, operationRow, transactionRow, ledger)

	switch operationRow.Type {
	case xdr.OperationTypeBumpSequence:
		e := operations.BumpSequence{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeCreateAccount:
		e := operations.CreateAccount{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypePayment:
		e := operations.Payment{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypePathPayment:
		e := operations.PathPayment{}
		e.Payment.Base = base
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeManageBuyOffer:
		e := operations.ManageBuyOffer{}
		e.Offer.Base = base
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeManageSellOffer:
		e := operations.ManageSellOffer{}
		e.Offer.Base = base
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeCreatePassiveSellOffer:
		e := operations.CreatePassiveSellOffer{}
		e.Offer.Base = base
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeSetOptions:
		e := operations.SetOptions{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeChangeTrust:
		e := operations.ChangeTrust{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeAllowTrust:
		e := operations.AllowTrust{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeAccountMerge:
		e := operations.AccountMerge{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeInflation:
		e := operations.Inflation{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	case xdr.OperationTypeManageData:
		e := operations.ManageData{Base: base}
		err = operationRow.UnmarshalDetails(&e)
		result = e
	default:
		result = base
	}

	return
}

// Populate fills out this resource using `row` as the source.
func PopulateBaseOperation(
	ctx context.Context,
	dest *operations.Base,
	operationRow history.Operation,
	transactionRow *history.Transaction,
	ledger history.Ledger,
) {
	dest.ID = fmt.Sprintf("%d", operationRow.ID)
	dest.PT = operationRow.PagingToken()
	// Check db2/history.Transaction.Successful field comment for more information.
	if operationRow.TransactionSuccessful == nil {
		dest.TransactionSuccessful = true
	} else {
		dest.TransactionSuccessful = *operationRow.TransactionSuccessful
	}
	dest.SourceAccount = operationRow.SourceAccount
	populateOperationType(dest, operationRow)
	dest.LedgerCloseTime = ledger.ClosedAt
	dest.TransactionHash = operationRow.TransactionHash

	lb := hal.LinkBuilder{Base: httpx.BaseURL(ctx)}
	self := fmt.Sprintf("/operations/%d", operationRow.ID)
	dest.Links.Self = lb.Link(self)
	dest.Links.Succeeds = lb.Linkf("/effects?order=desc&cursor=%s", dest.PT)
	dest.Links.Precedes = lb.Linkf("/effects?order=asc&cursor=%s", dest.PT)
	dest.Links.Transaction = lb.Linkf("/transactions/%s", operationRow.TransactionHash)
	dest.Links.Effects = lb.Link(self, "effects")

	if transactionRow != nil {
		dest.Transaction = new(aurora.Transaction)
		PopulateTransaction(ctx, dest.Transaction, *transactionRow)
	}
}

func populateOperationType(dest *operations.Base, row history.Operation) {
	var ok bool
	dest.TypeI = int32(row.Type)
	dest.Type, ok = operations.TypeNames[row.Type]

	if !ok {
		dest.Type = "unknown"
	}
}
