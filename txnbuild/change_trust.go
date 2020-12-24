package txnbuild

import (
	"math"

	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// ChangeTrust represents the DiamNet change trust operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html.
// If Limit is omitted, it defaults to txnbuild.MaxTrustlineLimit.
type ChangeTrust struct {
	Line          Asset
	Limit         string
	SourceAccount Account
}

// MaxTrustlineLimit represents the maximum value that can be set as a trustline limit.
var MaxTrustlineLimit = amount.StringFromInt64(math.MaxInt64)

// RemoveTrustlineOp returns a ChangeTrust operation to remove the trustline of the described asset,
// by setting the limit to "0".
func RemoveTrustlineOp(issuedAsset Asset) ChangeTrust {
	return ChangeTrust{
		Line:  issuedAsset,
		Limit: "0",
	}
}

// BuildXDR for ChangeTrust returns a fully configured XDR Operation.
func (ct *ChangeTrust) BuildXDR() (xdr.Operation, error) {
	if ct.Line.IsNative() {
		return xdr.Operation{}, errors.New("trustline cannot be extended to a native (XLM) asset")
	}
	xdrLine, err := ct.Line.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "can't convert trustline asset to XDR")
	}

	if ct.Limit == "" {
		ct.Limit = MaxTrustlineLimit
	}

	xdrLimit, err := amount.Parse(ct.Limit)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse limit amount")
	}

	opType := xdr.OperationTypeChangeTrust
	xdrOp := xdr.ChangeTrustOp{
		Line:  xdrLine,
		Limit: xdrLimit,
	}
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, ct.SourceAccount)
	return op, nil
}
