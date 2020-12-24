package txnbuild

import (
	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/price"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// CreatePassiveSellOffer represents the DiamNet create passive offer operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type CreatePassiveSellOffer struct {
	Selling       Asset
	Buying        Asset
	Amount        string
	Price         string
	SourceAccount Account
}

// BuildXDR for CreatePassiveSellOffer returns a fully configured XDR Operation.
func (cpo *CreatePassiveSellOffer) BuildXDR() (xdr.Operation, error) {
	xdrSelling, err := cpo.Selling.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Selling' field")
	}

	xdrBuying, err := cpo.Buying.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Buying' field")
	}

	xdrAmount, err := amount.Parse(cpo.Amount)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Amount'")
	}

	xdrPrice, err := price.Parse(cpo.Price)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Price'")
	}

	xdrOp := xdr.CreatePassiveSellOfferOp{
		Selling: xdrSelling,
		Buying:  xdrBuying,
		Amount:  xdrAmount,
		Price:   xdrPrice,
	}

	opType := xdr.OperationTypeCreatePassiveSellOffer
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, cpo.SourceAccount)
	return op, nil
}
