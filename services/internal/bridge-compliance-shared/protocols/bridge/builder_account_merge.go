package bridge

import (
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/txnbuild"
)

// AccountMergeOperationBody represents account_merge operation
type AccountMergeOperationBody struct {
	Source      *string
	Destination string
}

// Build returns a txnbuild.Operation
func (op AccountMergeOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.AccountMerge{
		Destination: op.Destination,
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op AccountMergeOperationBody) Validate() error {
	if !shared.IsValidAccountID(op.Destination) {
		return helpers.NewInvalidParameterError("destination", "Destination must start with `G`.")
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must start with `G`.")
	}

	return nil
}
