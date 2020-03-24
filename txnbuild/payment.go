package txnbuild

import (
	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// Payment represents the HcNet payment operation. See
// https://www.hcnet.org/developers/guides/concepts/list-of-operations.html
type Payment struct {
	Destination   string
	Amount        string
	Asset         Asset
	SourceAccount Account
}

// BuildXDR for Payment returns a fully configured XDR Operation.
func (p *Payment) BuildXDR() (xdr.Operation, error) {
	var destAccountID xdr.AccountId

	err := destAccountID.SetAddress(p.Destination)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set destination address")
	}

	xdrAmount, err := amount.Parse(p.Amount)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse amount")
	}

	if p.Asset == nil {
		return xdr.Operation{}, errors.New("you must specify an asset for payment")
	}
	xdrAsset, err := p.Asset.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set asset type")
	}

	opType := xdr.OperationTypePayment
	xdrOp := xdr.PaymentOp{
		Destination: destAccountID,
		Amount:      xdrAmount,
		Asset:       xdrAsset,
	}
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR Operation")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, p.SourceAccount)
	return op, nil
}
