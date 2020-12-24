package bridge

import (
	shared "github.com/diamnet/go/services/internal/bridge-compliance-shared"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/txnbuild"
)

// SetOptionsOperationBody represents set_options operation
type SetOptionsOperationBody struct {
	Source          *string
	InflationDest   *string           `json:"inflation_dest"`
	SetFlags        *[]int            `json:"set_flags"`
	ClearFlags      *[]int            `json:"clear_flags"`
	MasterWeight    *uint32           `json:"master_weight"`
	LowThreshold    *uint32           `json:"low_threshold"`
	MediumThreshold *uint32           `json:"medium_threshold"`
	HighThreshold   *uint32           `json:"high_threshold"`
	HomeDomain      *string           `json:"home_domain"`
	Signer          *SetOptionsSigner `json:"signer"`
}

// SetOptionsSigner is a struct that representing signer in SetOptions operation body
type SetOptionsSigner struct {
	PublicKey string `json:"public_key"`
	Weight    uint32 `json:"weight"`
}

// Build returns a txnbuild.Operation
func (op SetOptionsOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.SetOptions{}

	if op.InflationDest != nil {
		txnOp.InflationDestination = op.InflationDest
	}

	if op.SetFlags != nil {
		for _, flag := range *op.SetFlags {
			txnOp.SetFlags = append(txnOp.SetFlags, txnbuild.AccountFlag(flag))
		}
	}

	if op.ClearFlags != nil {
		for _, flag := range *op.ClearFlags {
			txnOp.ClearFlags = append(txnOp.ClearFlags, txnbuild.AccountFlag(flag))
		}
	}

	if op.MasterWeight != nil {
		txnOp.MasterWeight = txnbuild.NewThreshold(txnbuild.Threshold(*op.MasterWeight))
	}

	if op.LowThreshold != nil {
		txnOp.LowThreshold = txnbuild.NewThreshold(txnbuild.Threshold(*op.LowThreshold))
	}

	if op.MediumThreshold != nil {
		txnOp.MediumThreshold = txnbuild.NewThreshold(txnbuild.Threshold(*op.MediumThreshold))
	}

	if op.HighThreshold != nil {
		txnOp.HighThreshold = txnbuild.NewThreshold(txnbuild.Threshold(*op.HighThreshold))
	}

	if op.HomeDomain != nil {
		txnOp.HomeDomain = op.HomeDomain
	}

	if op.Signer != nil {
		txnOp.Signer = &txnbuild.Signer{
			Address: op.Signer.PublicKey,
			Weight:  txnbuild.Threshold(op.Signer.Weight),
		}
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op SetOptionsOperationBody) Validate() error {
	if op.InflationDest != nil && !shared.IsValidAccountID(*op.InflationDest) {
		return helpers.NewInvalidParameterError("inflation_dest", "Inflation destination must be a public key (starting with `G`).")
	}

	if op.Signer != nil {
		if !shared.IsValidAccountID(op.Signer.PublicKey) {
			return helpers.NewInvalidParameterError("signer.public_key", "Public key invlaid, must start with `G`.")
		}
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
