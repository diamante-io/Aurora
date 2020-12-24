package aurora

import (
	"strconv"
	gTime "time"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/support/time"
	"github.com/diamnet/go/xdr"
)

// Interface verifications
var _ actions.JSONer = (*TradeIndexAction)(nil)
var _ actions.EventStreamer = (*TradeIndexAction)(nil)

type TradeIndexAction struct {
	Action
	BaseAssetFilter       xdr.Asset
	HasBaseAssetFilter    bool
	CounterAssetFilter    xdr.Asset
	HasCounterAssetFilter bool
	OfferFilter           int64
	AccountFilter         string
	PagingParams          db2.PageQuery
	Records               []history.Trade
	Page                  hal.Page
}

// JSON is a method for actions.JSON
func (action *TradeIndexAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

// SSE is a method for actions.SSE
func (action *TradeIndexAction) SSE(stream *sse.Stream) error {
	action.Setup(
		action.EnsureHistoryFreshness,
		action.loadParams,
	)
	action.Do(
		action.loadRecords,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			records := action.Records[stream.SentCount():]

			for _, record := range records {
				var res aurora.Trade
				resourceadapter.PopulateTrade(action.R.Context(), &res, record)
				stream.Send(sse.Event{
					ID:   res.PagingToken(),
					Data: res,
				})
			}
		},
	)

	return action.Err
}

// loadParams sets action.Query from the request params
func (action *TradeIndexAction) loadParams() {
	action.PagingParams = action.GetPageQuery()
	action.BaseAssetFilter, action.HasBaseAssetFilter = action.MaybeGetAsset("base_")
	action.CounterAssetFilter, action.HasCounterAssetFilter = action.MaybeGetAsset("counter_")
	action.OfferFilter = action.GetInt64("offer_id")
	action.AccountFilter = action.GetAddress("account_id")

	if (!action.HasBaseAssetFilter && action.HasCounterAssetFilter) ||
		(action.HasBaseAssetFilter && !action.HasCounterAssetFilter) {
		action.SetInvalidField("base_asset_type,counter_asset_type", errors.New("this endpoint supports asset pairs but only one asset supplied"))
	}
}

// loadRecords populates action.Records
func (action *TradeIndexAction) loadRecords() {
	trades := action.HistoryQ().Trades()

	if action.AccountFilter != "" {
		trades.ForAccount(action.AccountFilter)
	}

	if action.HasBaseAssetFilter {

		baseAssetId, err := action.HistoryQ().GetAssetID(action.BaseAssetFilter)
		if err != nil {
			action.Err = err
			return
		}

		if action.HasCounterAssetFilter {

			counterAssetId, err := action.HistoryQ().GetAssetID(action.CounterAssetFilter)
			if err != nil {
				action.Err = err
				return
			}
			trades = action.HistoryQ().TradesForAssetPair(baseAssetId, counterAssetId)
		} else {
			action.Err = errors.New("this endpoint supports asset pairs but only one asset supplied")
			return
		}
	}

	if action.OfferFilter != 0 {
		trades = trades.ForOffer(action.OfferFilter)
	}

	action.Err = trades.Page(action.PagingParams).Select(&action.Records)
}

// loadPage populates action.Page
func (action *TradeIndexAction) loadPage() {
	for _, record := range action.Records {
		var res aurora.Trade
		resourceadapter.PopulateTrade(action.R.Context(), &res, record)
		action.Page.Add(res)
	}

	action.Page.FullURL = action.FullURL()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}

// Interface verification
var _ actions.JSONer = (*TradeAggregateIndexAction)(nil)

type TradeAggregateIndexAction struct {
	Action
	BaseAssetFilter    xdr.Asset
	CounterAssetFilter xdr.Asset
	StartTimeFilter    time.Millis
	EndTimeFilter      time.Millis
	OffsetFilter       int64
	ResolutionFilter   int64
	PagingParams       db2.PageQuery
	Records            []history.TradeAggregation
	Page               hal.Page
}

// JSON is a method for actions.JSON
func (action *TradeAggregateIndexAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

func (action *TradeAggregateIndexAction) loadParams() {
	action.PagingParams = action.GetPageQuery()
	action.BaseAssetFilter = action.GetAsset("base_")
	action.CounterAssetFilter = action.GetAsset("counter_")
	action.OffsetFilter = action.GetInt64("offset")
	action.StartTimeFilter = action.GetTimeMillis("start_time")
	action.EndTimeFilter = action.GetTimeMillis("end_time")
	action.ResolutionFilter = action.GetInt64("resolution")

	//check if resolution is legal
	resolutionDuration := gTime.Duration(action.ResolutionFilter) * gTime.Millisecond
	if history.StrictResolutionFiltering {
		if _, ok := history.AllowedResolutions[resolutionDuration]; !ok {
			action.SetInvalidField("resolution", errors.New("illegal or missing resolution. "+
				"allowed resolutions are: 1 minute (60000), 5 minutes (300000), 15 minutes (900000), 1 hour (3600000), "+
				"1 day (86400000) and 1 week (604800000)"))
		}
	}
	// check if offset is legal
	offsetDuration := gTime.Duration(action.OffsetFilter) * gTime.Millisecond
	if offsetDuration%gTime.Hour != 0 || offsetDuration >= gTime.Hour*24 || offsetDuration > resolutionDuration {
		action.SetInvalidField("offset", errors.New("illegal or missing offset. offset must be a multiple of an"+
			" hour, less than or equal to the resolution, and less than 24 hours"))
	}
}

// loadRecords populates action.Records
func (action *TradeAggregateIndexAction) loadRecords() {
	historyQ := action.HistoryQ()

	//get asset ids
	baseAssetId, err := historyQ.GetCreateAssetID(action.BaseAssetFilter)
	if err != nil {
		action.Err = err
		return
	}
	counterAssetId, err := historyQ.GetCreateAssetID(action.CounterAssetFilter)
	if err != nil {
		action.Err = err
		return
	}

	//initialize the query builder with required params
	tradeAggregationsQ, err := historyQ.GetTradeAggregationsQ(
		baseAssetId, counterAssetId, action.ResolutionFilter, action.OffsetFilter, action.PagingParams)
	if err != nil {
		action.Err = err
		return
	}

	//set time range if supplied
	if !action.StartTimeFilter.IsNil() {
		tradeAggregationsQ, err = tradeAggregationsQ.WithStartTime(action.StartTimeFilter)
		if err != nil {
			action.SetInvalidField("start_time", errors.New("illegal start time. adjusted start time must "+
				"be less than the provided end time if the end time is greater than 0"))
			return
		}
	}
	if !action.EndTimeFilter.IsNil() {
		tradeAggregationsQ, err = tradeAggregationsQ.WithEndTime(action.EndTimeFilter)
		if err != nil {
			action.SetInvalidField("end_time", errors.New("illegal end time. adjusted end time "+
				"must be greater than the offset and greater than the provided start time"))
			return
		}
	}

	action.Err = historyQ.Select(&action.Records, tradeAggregationsQ.GetSql())
}

func (action *TradeAggregateIndexAction) loadPage() {
	action.Page.Init()
	for _, record := range action.Records {
		var res aurora.TradeAggregation

		action.Err = resourceadapter.PopulateTradeAggregation(action.R.Context(), &res, record)
		if action.Err != nil {
			return
		}

		action.Page.Add(res)
	}

	action.Page.Limit = action.PagingParams.Limit
	action.Page.Order = action.PagingParams.Order

	newUrl := action.FullURL() // preserve scheme and host for the new url links
	q := newUrl.Query()

	action.Page.Links.Self = hal.NewLink(newUrl.String())

	//adjust time range for next page
	if uint64(len(action.Records)) == 0 {
		action.Page.Links.Next = action.Page.Links.Self
	} else {
		if action.PagingParams.Order == "asc" {
			newStartTime := action.Records[len(action.Records)-1].Timestamp + action.ResolutionFilter
			if newStartTime >= action.EndTimeFilter.ToInt64() {
				newStartTime = action.EndTimeFilter.ToInt64()
			}
			q.Set("start_time", strconv.FormatInt(newStartTime, 10))
			newUrl.RawQuery = q.Encode()
			action.Page.Links.Next = hal.NewLink(newUrl.String())
		} else { //desc
			newEndTime := action.Records[len(action.Records)-1].Timestamp
			if newEndTime <= action.StartTimeFilter.ToInt64() {
				newEndTime = action.StartTimeFilter.ToInt64()
			}
			q.Set("end_time", strconv.FormatInt(newEndTime, 10))
			newUrl.RawQuery = q.Encode()
			action.Page.Links.Next = hal.NewLink(newUrl.String())
		}
	}
}
