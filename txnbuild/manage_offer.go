package txnbuild

import (
	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/price"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

//CreateOfferOp returns a ManageSellOffer operation to create a new offer, by
// setting the OfferID to "0". The sourceAccount is optional, and if not provided,
// will be that of the surrounding transaction.
func CreateOfferOp(selling, buying Asset, amount, price string, sourceAccount ...Account) (ManageSellOffer, error) {
	if len(sourceAccount) > 1 {
		return ManageSellOffer{}, errors.New("offer can't have multiple source accounts")
	}
	offer := ManageSellOffer{
		Selling: selling,
		Buying:  buying,
		Amount:  amount,
		Price:   price,
		OfferID: 0,
	}
	if len(sourceAccount) == 1 {
		offer.SourceAccount = sourceAccount[0]
	}
	return offer, nil
}

// UpdateOfferOp returns a ManageSellOffer operation to update an offer.
// The sourceAccount is optional, and if not provided, will be that of
// the surrounding transaction.
func UpdateOfferOp(selling, buying Asset, amount, price string, offerID int64, sourceAccount ...Account) (ManageSellOffer, error) {
	if len(sourceAccount) > 1 {
		return ManageSellOffer{}, errors.New("offer can't have multiple source accounts")
	}
	offer := ManageSellOffer{
		Selling: selling,
		Buying:  buying,
		Amount:  amount,
		Price:   price,
		OfferID: offerID,
	}
	if len(sourceAccount) == 1 {
		offer.SourceAccount = sourceAccount[0]
	}
	return offer, nil
}

//DeleteOfferOp returns a ManageSellOffer operation to delete an offer, by
// setting the Amount to "0". The sourceAccount is optional, and if not provided,
// will be that of the surrounding transaction.
func DeleteOfferOp(offerID int64, sourceAccount ...Account) (ManageSellOffer, error) {
	// It turns out DiamNet core doesn't care about any of these fields except the amount.
	// However, Aurora will reject ManageSellOffer if it is missing fields.
	// Aurora will also reject if Buying == Selling.
	// Therefore unfortunately we have to make up some dummy values here.
	if len(sourceAccount) > 1 {
		return ManageSellOffer{}, errors.New("offer can't have multiple source accounts")
	}
	offer := ManageSellOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{Code: "FAKE", Issuer: "GBAQPADEYSKYMYXTMASBUIS5JI3LMOAWSTM2CHGDBJ3QDDPNCSO3DVAA"},
		Amount:  "0",
		Price:   "1",
		OfferID: offerID,
	}
	if len(sourceAccount) == 1 {
		offer.SourceAccount = sourceAccount[0]
	}
	return offer, nil
}

// ManageSellOffer represents the DiamNet manage offer operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type ManageSellOffer struct {
	Selling       Asset
	Buying        Asset
	Amount        string
	Price         string
	OfferID       int64
	SourceAccount Account
}

// BuildXDR for ManageSellOffer returns a fully configured XDR Operation.
func (mo *ManageSellOffer) BuildXDR() (xdr.Operation, error) {
	xdrSelling, err := mo.Selling.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Selling' field")
	}

	xdrBuying, err := mo.Buying.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Buying' field")
	}

	xdrAmount, err := amount.Parse(mo.Amount)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Amount'")
	}

	xdrPrice, err := price.Parse(mo.Price)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Price'")
	}

	opType := xdr.OperationTypeManageSellOffer
	xdrOp := xdr.ManageSellOfferOp{
		Selling: xdrSelling,
		Buying:  xdrBuying,
		Amount:  xdrAmount,
		Price:   xdrPrice,
		OfferId: xdr.Int64(mo.OfferID),
	}
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}

	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, mo.SourceAccount)
	return op, nil
}
