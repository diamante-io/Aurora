package core

import (
	"fmt"
	"math/big"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// PagingToken returns a suitable paging token for the Offer
func (r Offer) PagingToken() string {
	return fmt.Sprintf("%d", r.OfferID)
}

// PriceAsString return the price fraction as a floating point approximate.
func (r Offer) PriceAsString() string {
	return big.NewRat(int64(r.Pricen), int64(r.Priced)).FloatString(7)
}

// ConnectedAssets loads xdr.Asset records for the purposes of path
// finding.  Given the input asset type, a list of xdr.Assets is returned that
// each have some available trades for the input asset.
func (q *Q) ConnectedAssets(dest interface{}, selling xdr.Asset) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	if schemaVersion < 9 {
		return q.connectedAssetsSchema8(dest, selling)
	} else {
		return q.connectedAssetsSchema9(dest, selling)
	}
}

func (q *Q) connectedAssetsSchema9(dest interface{}, selling xdr.Asset) error {
	assets, ok := dest.(*[]xdr.Asset)
	if !ok {
		return errors.New("dest is not *[]xdr.Asset")
	}

	sellingAssetXDRString, err := xdr.MarshalBase64(selling)
	if err != nil {
		return errors.Wrap(err, "Error marshaling selling")
	}

	sql := sq.Select("buyingasset").
		From("offers").
		Where(sq.Eq{"sellingasset": sellingAssetXDRString}).
		GroupBy("buyingasset")

	var rows []struct {
		Asset xdr.Asset `db:"buyingasset"`
	}

	err = q.Select(&rows, sql)

	if err != nil {
		return err
	}

	results := make([]xdr.Asset, len(rows))
	*assets = results

	for i, r := range rows {
		results[i] = r.Asset
	}

	return nil
}

// ConnectedAssets loads xdr.Asset records for the purposes of path
// finding.  Given the input asset type, a list of xdr.Assets is returned that
// each have some available trades for the input asset.
func (q *Q) connectedAssetsSchema8(dest interface{}, selling xdr.Asset) error {
	assets, ok := dest.(*[]xdr.Asset)
	if !ok {
		return errors.New("dest is not *[]xdr.Asset")
	}

	var (
		t xdr.AssetType
		c string
		i string
	)

	err := selling.Extract(&t, &c, &i)
	if err != nil {
		return err
	}

	sql := sq.Select(
		"buyingassettype AS type",
		"coalesce(buyingassetcode, '') AS code",
		"coalesce(buyingissuer, '') AS issuer").
		From("offers").
		Where(sq.Eq{"sellingassettype": t}).
		GroupBy("buyingassettype", "buyingassetcode", "buyingissuer")

	if t != xdr.AssetTypeAssetTypeNative {
		sql = sql.Where(sq.Eq{"sellingassetcode": c, "sellingissuer": i})
	}

	var rows []struct {
		Type   xdr.AssetType
		Code   string
		Issuer string
	}

	err = q.Select(&rows, sql)

	if err != nil {
		return err
	}

	results := make([]xdr.Asset, len(rows))
	*assets = results

	for i, r := range rows {
		results[i], err = AssetFromDB(r.Type, r.Code, r.Issuer)
		if err != nil {
			return err
		}
	}

	return nil
}

// OffersByAddress loads a page of active offers for the given
// address.
func (q *Q) OffersByAddress(dest interface{}, addy string, pq db2.PageQuery) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	offers := []internalOffer{}

	sql := sq.Select("co.*").
		From("offers co").
		Where("co.sellerid = ?", addy).
		Limit(uint64(pq.Limit))

	cursor, err := pq.CursorInt64()
	if err != nil {
		return err
	}

	switch pq.Order {
	case "asc":
		sql = sql.Where("co.offerid > ?", cursor).OrderBy("co.offerid asc")
	case "desc":
		sql = sql.Where("co.offerid < ?", cursor).OrderBy("co.offerid desc")
	}

	err = q.Select(&offers, sql)
	if err != nil {
		return err
	}

	newOffers := make([]Offer, len(offers))

	for i, offer := range offers {
		newOffers[i] = offer.get()
	}

	if schemaVersion >= 9 {
		*dest.(*[]Offer) = newOffers
		return nil
	}

	// Convert schema 8 results to xdr.Assets
	for i, offer := range offers {
		var sellingAsset, buyingAsset xdr.Asset

		if offer.SellingAssetType == xdr.AssetTypeAssetTypeNative {
			sellingAsset.SetNative()
		} else {
			var account xdr.AccountId
			err := account.SetAddress(offer.SellingIssuer.String)
			if err != nil {
				return errors.Wrap(err, "Error setting offer.SellingIssuer")
			}
			sellingAsset.SetCredit(offer.SellingAssetCode.String, account)
		}

		if offer.BuyingAssetType == xdr.AssetTypeAssetTypeNative {
			buyingAsset.SetNative()
		} else {
			var account xdr.AccountId
			err := account.SetAddress(offer.BuyingIssuer.String)
			if err != nil {
				return errors.Wrap(err, "Error setting offer.BuyingIssuer")
			}
			buyingAsset.SetCredit(offer.BuyingAssetCode.String, account)
		}

		newOffers[i].SellingAsset = sellingAsset
		newOffers[i].BuyingAsset = buyingAsset
	}

	*dest.(*[]Offer) = newOffers
	return nil
}
