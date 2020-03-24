package bridge

import (
	b "github.com/hcnet/go/build"
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/txnbuild"
)

// InflationOperationBody represents inflation operation
type InflationOperationBody struct {
	Source *string
}

// Build returns a txnbuild.Operation
func (op InflationOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.Inflation{}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// ToTransactionMutator returns go-hcnet-base TransactionMutator
func (op InflationOperationBody) ToTransactionMutator() b.TransactionMutator {
	var mutators []interface{}

	if op.Source != nil {
		mutators = append(mutators, b.SourceAccount{*op.Source})
	}

	return b.Inflation(mutators...)
}

// Validate validates if operation body is valid.
func (op InflationOperationBody) Validate() error {
	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
