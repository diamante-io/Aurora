package resourceadapter

import (
	"context"
	"fmt"

	"github.com/hcnet/go/amount"
	protocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/support/render/hal"
)

// Populate fills out the details of a trade using a row from the history_trades
// table.
func PopulateTrade(
	ctx context.Context,
	dest *protocol.Trade,
	row history.Trade,
) {
	dest.ID = row.PagingToken()
	dest.PT = row.PagingToken()
	dest.OfferID = fmt.Sprintf("%d", row.OfferID)
	dest.BaseOfferID = ""
	if row.BaseOfferID != nil {
		dest.BaseOfferID = fmt.Sprintf("%d", *row.BaseOfferID)
	}
	dest.BaseAccount = row.BaseAccount
	dest.BaseAssetType = row.BaseAssetType
	dest.BaseAssetCode = row.BaseAssetCode
	dest.BaseAssetIssuer = row.BaseAssetIssuer
	dest.BaseAmount = amount.String(row.BaseAmount)
	dest.CounterOfferID = ""
	if row.CounterOfferID != nil {
		dest.CounterOfferID = fmt.Sprintf("%d", *row.CounterOfferID)
	}
	dest.CounterAccount = row.CounterAccount
	dest.CounterAssetType = row.CounterAssetType
	dest.CounterAssetCode = row.CounterAssetCode
	dest.CounterAssetIssuer = row.CounterAssetIssuer
	dest.CounterAmount = amount.String(row.CounterAmount)
	dest.LedgerCloseTime = row.LedgerCloseTime
	dest.BaseIsSeller = row.BaseIsSeller

	if row.HasPrice() {
		dest.Price = &protocol.Price{
			N: int32(row.PriceN.Int64),
			D: int32(row.PriceD.Int64),
		}
	}

	populateTradeLinks(ctx, dest, row.HistoryOperationID)
}

func populateTradeLinks(
	ctx context.Context,
	dest *protocol.Trade,
	opid int64,
) {
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	dest.Links.Base = lb.Link("/accounts", dest.BaseAccount)
	dest.Links.Counter = lb.Link("/accounts", dest.CounterAccount)
	dest.Links.Operation = lb.Link(
		"/operations",
		fmt.Sprintf("%d", opid),
	)
}
