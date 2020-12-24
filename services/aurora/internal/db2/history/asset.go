package history

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/xdr"
)

func (q *Q) GetAssetByID(dest interface{}, id int64) (err error) {
	sql := sq.Select("id", "asset_type", "asset_code", "asset_issuer").From("history_assets").Limit(1).Where(sq.Eq{"id": id})
	err = q.Get(dest, sql)
	return
}

// GetAssetIDs fetches the ids for many Assets at once
func (q *Q) GetAssetIDs(assets []xdr.Asset) ([]int64, error) {
	list := make([]string, 0, len(assets))
	for _, asset := range assets {
		list = append(list, asset.String())
	}

	sql := sq.Select("id").From("history_assets").Where(sq.Eq{
		"concat(asset_type, '/', asset_code, '/', asset_issuer)": list,
	})

	var ids []int64
	err := q.Select(&ids, sql)
	return ids, err
}

// GetAssetID fetches the id for an Asset. If fetching multiple values, look at GetAssetIDs
func (q *Q) GetAssetID(asset xdr.Asset) (id int64, err error) {
	var (
		assetType   string
		assetCode   string
		assetIssuer string
	)

	err = asset.Extract(&assetType, &assetCode, &assetIssuer)
	if err != nil {
		return
	}

	sql := sq.Select("id").From("history_assets").Limit(1).Where(sq.Eq{
		"asset_type":   assetType,
		"asset_code":   assetCode,
		"asset_issuer": assetIssuer})

	err = q.Get(&id, sql)
	return
}

// Get asset row id. If asset is first seen, it will be inserted and the new id returned.
func (q *Q) GetCreateAssetID(
	asset xdr.Asset,
) (result int64, err error) {

	result, err = q.GetAssetID(asset)

	//asset exists, return id
	if err == nil {
		return
	}

	//unexpected error
	if !q.NoRows(err) {
		return
	}

	//insert asset and return id
	var (
		assetType   string
		assetCode   string
		assetIssuer string
	)

	err = asset.Extract(&assetType, &assetCode, &assetIssuer)
	if err != nil {
		return
	}

	err = q.GetRaw(&result,
		`INSERT INTO history_assets (asset_type, asset_code, asset_issuer) VALUES (?,?,?) RETURNING id`,
		assetType, assetCode, assetIssuer)

	return
}
