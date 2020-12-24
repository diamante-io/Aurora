package txnbuild

import (
	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// CreateAccount represents the DiamNet create account operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type CreateAccount struct {
	Destination   string
	Amount        string
	SourceAccount Account
}

// BuildXDR for CreateAccount returns a fully configured XDR Operation.
func (ca *CreateAccount) BuildXDR() (xdr.Operation, error) {
	var xdrOp xdr.CreateAccountOp

	err := xdrOp.Destination.SetAddress(ca.Destination)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set destination address")
	}

	xdrOp.StartingBalance, err = amount.Parse(ca.Amount)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse amount")
	}

	opType := xdr.OperationTypeCreateAccount
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, ca.SourceAccount)
	return op, nil
}
