package history

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// GetOfferByID loads a row from the `offers` table, selected by offerid.
func (q *Q) GetOfferByID(id int64) (Offer, error) {
	var offer Offer
	sql := selectOffers.Where("offers.offerid = ?", id)
	err := q.Get(&offer, sql)
	return offer, err
}

// GetAllOffers loads a row from `history_accounts`, by address
func (q *Q) GetAllOffers() ([]Offer, error) {
	var offers []Offer
	err := q.Select(&offers, selectOffers)
	return offers, err
}

// UpsertOffer creates / updates a row in the offers table
func (q *Q) UpsertOffer(offer xdr.OfferEntry, lastModifiedLedger xdr.Uint32) error {
	var price float64
	if offer.Price.N > 0 {
		price = float64(offer.Price.N) / float64(offer.Price.D)
	} else if offer.Price.D == 0 {
		return errors.New("offer price denominator is zero")
	}
	buyingAsset, err := xdr.MarshalBase64(offer.Buying)
	if err != nil {
		return errors.Wrap(err, "cannot marshal buying asset in offer")
	}
	sellingAsset, err := xdr.MarshalBase64(offer.Selling)
	if err != nil {
		return errors.Wrap(err, "cannot marshal selling asset in offer")
	}
	sql := sq.Insert("offers").SetMap(
		map[string]interface{}{
			"sellerid":             offer.SellerId.Address(),
			"offerid":              offer.OfferId,
			"sellingasset":         sellingAsset,
			"buyingasset":          buyingAsset,
			"amount":               offer.Amount,
			"pricen":               offer.Price.N,
			"priced":               offer.Price.D,
			"price":                price,
			"flags":                offer.Flags,
			"last_modified_ledger": lastModifiedLedger,
		},
	).Suffix(`
			ON CONFLICT (offerid) DO UPDATE SET
				sellerid=EXCLUDED.sellerid,
				sellingasset=EXCLUDED.sellingasset,
				buyingasset=EXCLUDED.buyingasset,
				amount=EXCLUDED.amount,
				pricen=EXCLUDED.pricen,
				priced=EXCLUDED.priced,
				price=EXCLUDED.price,
				flags=EXCLUDED.flags,
				last_modified_ledger=EXCLUDED.last_modified_ledger
		`)

	_, err = q.Exec(sql)
	return err
}

// RemoveOffer deletes a row in the offers table
func (q *Q) RemoveOffer(offerID xdr.Int64) error {
	sql := sq.Delete("offers").Where(sq.Eq{
		"offerid": offerID,
	})

	_, err := q.Exec(sql)
	return err
}

var selectOffers = sq.Select(`
	sellerid,
	offerid,
	sellingasset,
	buyingasset,
	amount,
	pricen,
	priced,
	price,
	flags,
	last_modified_ledger
`).From("offers")
