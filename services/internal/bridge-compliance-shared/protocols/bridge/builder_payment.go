package bridge

import (
	"github.com/hcnet/go/amount"
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
	"github.com/hcnet/go/txnbuild"
)

// PaymentOperationBody represents payment operation
type PaymentOperationBody struct {
	Source      *string
	Destination string
	Amount      string
	Asset       protocols.Asset
}

// Build returns a txnbuild.Operation
func (op PaymentOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.Payment{
		Destination: op.Destination,
		Amount:      op.Amount,
		Asset:       op.Asset.ToBaseAsset(),
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op PaymentOperationBody) Validate() error {
	if !shared.IsValidAccountID(op.Destination) {
		return helpers.NewInvalidParameterError("destination", "Destination must be a public key (starting with `G`).")
	}

	_, err := amount.Parse(op.Amount)
	if err != nil {
		return helpers.NewInvalidParameterError("amount", "Invalid amount.")
	}

	err = op.Asset.Validate()
	if err != nil {
		return helpers.NewInvalidParameterError("asset", "Invalid asset.")
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
