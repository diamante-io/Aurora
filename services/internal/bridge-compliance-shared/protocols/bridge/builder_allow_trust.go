package bridge

import (
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/txnbuild"
)

// AllowTrustOperationBody represents allow_trust operation
type AllowTrustOperationBody struct {
	Source    *string
	AssetCode string `json:"asset_code"`
	Trustor   string
	Authorize bool
}

// Build returns a txnbuild.Operation
func (op AllowTrustOperationBody) Build() txnbuild.Operation {
	txnOp := txnbuild.AllowTrust{
		Trustor:   op.Trustor,
		Authorize: op.Authorize,
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
		txnOp.Type = txnbuild.CreditAsset{Code: op.AssetCode, Issuer: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op AllowTrustOperationBody) Validate() error {
	// Note (Peter 23-05-2019): We need source account to be set here because it is used for
	// creating an asset type which is needed by txnbuild.AllowTrust.
	// to do: Update documentation for bridge server to indicate that source account is required for
	// AllowTrust operation. Alternatively, update txnbuild to ignore issuer when building
	// AllowTrust operations.

	if op.Source == nil {
		return helpers.NewInvalidParameterError("source", "Source must be specified for AllowTrust operation.")
	}

	if !shared.IsValidAssetCode(op.AssetCode) {
		return helpers.NewInvalidParameterError("asset_code", "Asset code is invalid")
	}

	if !shared.IsValidAccountID(op.Trustor) {
		return helpers.NewInvalidParameterError("trustor", "Trustor must be a public key (starting with `G`).")
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
