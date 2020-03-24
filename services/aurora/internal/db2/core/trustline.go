package core

import (
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/xdr"
)

func (tl Trustline) IsAuthorized() bool {
	return (tl.Flags & int32(xdr.TrustLineFlagsAuthorizedFlag)) != 0
}

// AssetsForAddress returns a list of assets and balances for those assets held by
// a given address.
func (q *Q) AssetsForAddress(addy string) ([]xdr.Asset, []xdr.Int64, error) {
	var tls []Trustline
	var account Account

	if err := q.AccountByAddress(&account, addy); q.NoRows(err) {
		// if there is no account for the given address then
		// we return an empty list of assets and balances
		return []xdr.Asset{}, []xdr.Int64{}, nil
	} else if err != nil {
		return nil, nil, err
	}

	if err := q.TrustlinesByAddress(&tls, addy); err != nil {
		return nil, nil, err
	}

	assets := make([]xdr.Asset, len(tls)+1)
	balances := make([]xdr.Int64, len(tls)+1)

	var err error
	for i, tl := range tls {
		assets[i], err = AssetFromDB(tl.Assettype, tl.Assetcode, tl.Issuer)
		if err != nil {
			return nil, nil, err
		}
		balances[i] = tl.Balance
	}

	assets[len(assets)-1], err = xdr.NewAsset(xdr.AssetTypeAssetTypeNative, nil)
	balances[len(assets)-1] = account.Balance

	return assets, balances, err
}

// AllAssets loads all (unique) assets from core DB
func (q *Q) AllAssets(dest interface{}) error {
	var tls []Trustline

	sql := sq.Select(
		"tl.assettype",
		"tl.issuer",
		"tl.assetcode",
	).From("trustlines tl").GroupBy("(tl.assettype, tl.issuer, tl.assetcode)")
	err := q.Select(&tls, sql)
	if err != nil {
		return err
	}

	dtl, ok := dest.(*[]xdr.Asset)
	if !ok {
		return errors.New("Invalid destination")
	}

	result := make([]xdr.Asset, len(tls))
	*dtl = result

	for i, tl := range tls {
		result[i], err = AssetFromDB(tl.Assettype, tl.Assetcode, tl.Issuer)
		if err != nil {
			return err
		}
	}

	return nil
}

// TrustlinesByAddress loads all trustlines for `addy`
func (q *Q) TrustlinesByAddress(dest interface{}, addy string) error {
	sql := selectTrustline.Where("accountid = ?", addy)
	return q.Select(dest, sql)
}

// BalancesForAsset returns all the balances by asset type, code, issuer
func (q *Q) BalancesForAsset(
	assetType int32,
	assetCode string,
	assetIssuer string,
) (int32, string, error) {
	sql := selectBalances.Where(sq.Eq{
		"assettype": assetType,
		"assetcode": assetCode,
		"issuer":    assetIssuer,
		"flags":     1,
	})
	result := struct {
		Count int32  `db:"count"`
		Sum   string `db:"sum"`
	}{}
	err := q.Get(&result, sql)
	return result.Count, result.Sum, err
}

var selectTrustline = sq.Select(
	"tl.accountid",
	"tl.assettype",
	"tl.issuer",
	"tl.assetcode",
	"tl.tlimit",
	"tl.balance",
	"tl.flags",
	"tl.lastmodified",
	// Liabilities can be NULL so can error without `coalesce`:
	// `Invalid value for xdr.Int64`
	"coalesce(tl.buyingliabilities, 0) as buyingliabilities",
	"coalesce(tl.sellingliabilities, 0) as sellingliabilities",
).From("trustlines tl")

var selectBalances = sq.Select("COUNT(*)", "COALESCE(SUM(balance), 0) as sum").From("trustlines")
