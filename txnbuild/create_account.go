package txnbuild

import (
	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// CreateAccount represents the HcNet create account operation. See
// https://www.hcnet.org/developers/guides/concepts/list-of-operations.html
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
