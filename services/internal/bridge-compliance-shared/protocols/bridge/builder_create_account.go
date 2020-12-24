package bridge

import (
	"github.com/diamnet/go/amount"
	shared "github.com/diamnet/go/services/internal/bridge-compliance-shared"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/txnbuild"
)

// CreateAccountOperationBody represents create_account operation
type CreateAccountOperationBody struct {
	Source          *string
	Destination     string
	StartingBalance string `json:"starting_balance"`
}

// Build returns a txnbuild.Operation
func (op CreateAccountOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.CreateAccount{
		Destination: op.Destination,
		Amount:      op.StartingBalance,
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op CreateAccountOperationBody) Validate() error {
	if !shared.IsValidAccountID(op.Destination) {
		return helpers.NewInvalidParameterError("destination", "Destination must be a public key (starting with `G`)")
	}

	_, err := amount.Parse(op.StartingBalance)
	if err != nil {
		return helpers.NewInvalidParameterError("starting_balance", "Not a valid amount.")
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`)")
	}

	return nil
}
