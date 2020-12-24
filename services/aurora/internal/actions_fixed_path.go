package aurora

import (
	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/paths"
	"github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/services/aurora/internal/simplepath"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/xdr"
)

// Interface verification
var _ actions.JSONer = (*FixedPathIndexAction)(nil)

// FixedPathIndexAction provides path finding where the source asset and destination asset is fixed
type FixedPathIndexAction struct {
	Action

	sourceAccount    *xdr.AccountId
	sourceAsset      xdr.Asset
	amountToSpend    xdr.Int64
	destinationAsset xdr.Asset

	Records []paths.Path
	Page    hal.BasePage
}

// JSON implements actions.JSON
func (action *FixedPathIndexAction) JSON() error {
	action.Do(
		action.loadQuery,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

func (action *FixedPathIndexAction) loadQuery() {
	if action.Base.GetString("source_account") != "" {
		accountID := action.Base.GetAccountID("source_account")
		action.sourceAccount = &accountID
	} else {
		action.sourceAccount = nil
	}
	action.destinationAsset = action.GetAsset("destination_")
	action.sourceAsset = action.GetAsset("source_")
	action.amountToSpend = action.GetPositiveAmount("source_amount")
}

func (action *FixedPathIndexAction) loadRecords() {
	action.Records, action.Err = action.App.paths.FindFixedPaths(
		action.sourceAccount,
		action.sourceAsset,
		action.amountToSpend,
		action.destinationAsset,
		action.App.config.MaxPathLength,
	)
	if action.Err == simplepath.ErrEmptyInMemoryOrderBook {
		action.Err = problem.StillIngesting
	}
}

func (action *FixedPathIndexAction) loadPage() {
	action.Page.Init()
	for _, p := range action.Records {
		var res aurora.Path
		action.Err = resourceadapter.PopulatePath(action.R.Context(), &res, p)

		if action.Err != nil {
			return
		}
		action.Page.Add(res)
	}
}
