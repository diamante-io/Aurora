package txnbuild

import (
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// AllowTrust represents the HcNet allow trust operation. See
// https://www.hcnet.org/developers/guides/concepts/list-of-operations.html
type AllowTrust struct {
	Trustor       string
	Type          Asset
	Authorize     bool
	SourceAccount Account
}

// BuildXDR for AllowTrust returns a fully configured XDR Operation.
func (at *AllowTrust) BuildXDR() (xdr.Operation, error) {
	var xdrOp xdr.AllowTrustOp

	// Set XDR address associated with the trustline
	err := xdrOp.Trustor.SetAddress(at.Trustor)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set trustor address")
	}

	// Validate this is an issued asset
	if at.Type.IsNative() {
		return xdr.Operation{}, errors.New("trustline doesn't exist for a native (XLM) asset")
	}

	// AllowTrust has a special asset type - map to it
	xdrAsset := xdr.Asset{}

	xdrOp.Asset, err = xdrAsset.ToAllowTrustOpAsset(at.Type.GetCode())
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "can't convert asset for trustline to allow trust asset type")
	}

	// Set XDR auth flag
	xdrOp.Authorize = at.Authorize

	opType := xdr.OperationTypeAllowTrust
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, at.SourceAccount)
	return op, nil
}
