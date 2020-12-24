package aurora

import (
	"fmt"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/assets"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/render/hal"
)

// This file contains the actions:
//
// AssetsAction: pages of assets

// Interface verification
var _ actions.JSONer = (*AssetsAction)(nil)

// AssetsAction renders a page of Assets
type AssetsAction struct {
	Action
	AssetCode    string
	AssetIssuer  string
	PagingParams db2.PageQuery
	Records      []assets.AssetStatsR
	Page         hal.Page
}

const maxAssetCodeLength = 12

// JSON is a method for actions.JSON
func (action *AssetsAction) JSON() error {
	action.Do(
		action.loadParams,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

func (action *AssetsAction) loadParams() {
	action.AssetCode = action.GetString("asset_code")
	if len(action.AssetCode) > maxAssetCodeLength {
		action.SetInvalidField("asset_code", fmt.Errorf("max length is: %d", maxAssetCodeLength))
		return
	}

	if len(action.GetString("asset_issuer")) > 0 {
		issuerAccount := action.GetAccountID("asset_issuer")
		if action.Err != nil {
			return
		}
		action.AssetIssuer = issuerAccount.Address()
	}
	action.PagingParams = action.GetPageQuery(actions.DisableCursorValidation)
}

func (action *AssetsAction) loadRecords() {
	sql, err := assets.AssetStatsQ{
		AssetCode:   &action.AssetCode,
		AssetIssuer: &action.AssetIssuer,
		PageQuery:   &action.PagingParams,
	}.GetSQL()
	if err != nil {
		action.Err = err
		return
	}
	action.Err = action.HistoryQ().Select(&action.Records, sql)
}

func (action *AssetsAction) loadPage() {
	for _, record := range action.Records {
		var res aurora.AssetStat
		err := resourceadapter.PopulateAssetStat(action.R.Context(), &res, record)
		if err != nil {
			action.Err = err
			return
		}
		action.Page.Add(res)
	}

	action.Page.FullURL = action.FullURL()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}
