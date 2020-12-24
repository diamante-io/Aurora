package txnbuild

import (
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// AccountMerge represents the DiamNet merge account operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type AccountMerge struct {
	Destination   string
	SourceAccount Account
}

// BuildXDR for AccountMerge returns a fully configured XDR Operation.
func (am *AccountMerge) BuildXDR() (xdr.Operation, error) {
	var xdrOp xdr.AccountId

	err := xdrOp.SetAddress(am.Destination)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set destination address")
	}

	opType := xdr.OperationTypeAccountMerge
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, am.SourceAccount)
	return op, nil
}
