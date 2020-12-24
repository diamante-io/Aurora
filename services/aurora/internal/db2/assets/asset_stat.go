package assets

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2"
)

// AssetStatsR is the result from the AssetStatsQ query
type AssetStatsR struct {
	SortKey     string `db:"sort_key"`
	Type        string `db:"asset_type"`
	Code        string `db:"asset_code"`
	Issuer      string `db:"asset_issuer"`
	Amount      string `db:"amount"`
	NumAccounts int32  `db:"num_accounts"`
	Flags       int8   `db:"flags"`
	Toml        string `db:"toml"`
}

// PagingToken implementation for hal.Pageable
//func (res AssetStat) PagingToken() string {
//	return res.PT
//}

// AssetStatsQ is the query to fetch all assets in the system
type AssetStatsQ struct {
	AssetCode   *string
	AssetIssuer *string
	PageQuery   *db2.PageQuery
}

// GetSQL allows this query to be executed by the caller
func (q AssetStatsQ) GetSQL() (sq.SelectBuilder, error) {
	sql := selectQuery
	if q.AssetCode != nil && *q.AssetCode != "" {
		sql = sql.Where("hist.asset_code = ?", *q.AssetCode)
	}
	if q.AssetIssuer != nil && *q.AssetIssuer != "" {
		sql = sql.Where("hist.asset_issuer = ?", *q.AssetIssuer)
	}

	var err error
	if q.PageQuery != nil {
		// cursor needs to work for descending case as well
		cursor := q.PageQuery.Cursor
		if q.PageQuery.Order == "desc" && cursor == "" {
			cursor = "zzzzzzzzzzzzz" // 12 + 1 "z"s so it will always be greater than the _ delimiter since code is max 12 chars
		}

		sql, err = q.PageQuery.ApplyToUsingCursor(sql, "concat(hist.asset_code, '_', hist.asset_issuer, '_', hist.asset_type)", cursor)
		if err != nil {
			return sql, err
		}
	} else {
		sql = sql.OrderBy("sort_key ASC")
	}
	return sql, nil
}

var selectQuery = sq.
	Select(
		"concat(hist.asset_code, '_', hist.asset_issuer, '_', hist.asset_type) as sort_key",
		"hist.asset_type",
		"hist.asset_code",
		"hist.asset_issuer",
		"stats.amount",
		"stats.num_accounts",
		"stats.flags",
		"stats.toml",
	).
	From("history_assets hist").
	Join("asset_stats stats ON hist.id = stats.id")
