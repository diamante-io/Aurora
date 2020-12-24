package history

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/support/errors"
	strtime "github.com/diamnet/go/support/time"
	"github.com/diamnet/go/xdr"
)

// AllowedResolutions is the set of trade aggregation time windows allowed to be used as the
// `resolution` parameter.
var AllowedResolutions = map[time.Duration]struct{}{
	time.Minute:        {}, //1 minute
	time.Minute * 5:    {}, //5 minutes
	time.Minute * 15:   {}, //15 minutes
	time.Hour:          {}, //1 hour
	time.Hour * 24:     {}, //day
	time.Hour * 24 * 7: {}, //week
}

// StrictResolutionFiltering represents a simple feature flag to determine whether only
// predetermined resolutions of trade aggregations are allowed.
var StrictResolutionFiltering = true

// TradeAggregation represents an aggregation of trades from the trades table
type TradeAggregation struct {
	Timestamp     int64     `db:"timestamp"`
	TradeCount    int64     `db:"count"`
	BaseVolume    string    `db:"base_volume"`
	CounterVolume string    `db:"counter_volume"`
	Average       float64   `db:"avg"`
	High          xdr.Price `db:"high"`
	Low           xdr.Price `db:"low"`
	Open          xdr.Price `db:"open"`
	Close         xdr.Price `db:"close"`
}

// TradeAggregationsQ is a helper struct to aid in configuring queries to
// bucket and aggregate trades
type TradeAggregationsQ struct {
	baseAssetID    int64
	counterAssetID int64
	resolution     int64
	offset         int64
	startTime      strtime.Millis
	endTime        strtime.Millis
	pagingParams   db2.PageQuery
}

// GetTradeAggregationsQ initializes a TradeAggregationsQ query builder based on the required parameters
func (q Q) GetTradeAggregationsQ(baseAssetID int64, counterAssetID int64, resolution int64,
	offset int64, pagingParams db2.PageQuery) (*TradeAggregationsQ, error) {

	//convert resolution to a duration struct
	resolutionDuration := time.Duration(resolution) * time.Millisecond
	offsetDuration := time.Duration(offset) * time.Millisecond

	//check if resolution allowed
	if StrictResolutionFiltering {
		if _, ok := AllowedResolutions[resolutionDuration]; !ok {
			return &TradeAggregationsQ{}, errors.New("resolution is not allowed")
		}
	}
	// check if offset is allowed. Offset must be 1) a multiple of an hour 2) less than the resolution and 3)
	// less than 24 hours
	if offsetDuration%time.Hour != 0 || offsetDuration >= time.Hour*24 || offsetDuration > resolutionDuration {
		return &TradeAggregationsQ{}, errors.New("offset is not allowed.")
	}

	return &TradeAggregationsQ{
		baseAssetID:    baseAssetID,
		counterAssetID: counterAssetID,
		resolution:     resolution,
		offset:         offset,
		pagingParams:   pagingParams,
	}, nil
}

// WithStartTime adds an optional lower time boundary filter to the trades being aggregated.
func (q *TradeAggregationsQ) WithStartTime(startTime strtime.Millis) (*TradeAggregationsQ, error) {
	offsetMillis := strtime.MillisFromInt64(q.offset)
	var adjustedStartTime strtime.Millis
	// Round up to offset if the provided start time is less than the offset.
	if startTime < offsetMillis {
		adjustedStartTime = offsetMillis
	} else {
		adjustedStartTime = (startTime - offsetMillis).RoundUp(q.resolution) + offsetMillis
	}
	if !q.endTime.IsNil() && adjustedStartTime > q.endTime {
		return &TradeAggregationsQ{}, errors.New("start time is not allowed")
	} else {
		q.startTime = adjustedStartTime
		return q, nil
	}
}

// WithEndTime adds an upper optional time boundary filter to the trades being aggregated.
func (q *TradeAggregationsQ) WithEndTime(endTime strtime.Millis) (*TradeAggregationsQ, error) {
	// Round upper boundary down, to not deliver partial bucket
	offsetMillis := strtime.MillisFromInt64(q.offset)
	var adjustedEndTime strtime.Millis
	// the end time isn't allowed to be less than the offset
	if endTime < offsetMillis {
		return &TradeAggregationsQ{}, errors.New("end time is not allowed")
	} else {
		adjustedEndTime = (endTime - offsetMillis).RoundDown(q.resolution) + offsetMillis
	}
	if adjustedEndTime < q.startTime {
		return &TradeAggregationsQ{}, errors.New("end time is not allowed")
	} else {
		q.endTime = adjustedEndTime
		return q, nil
	}
}

// GetSql generates a sql statement to aggregate Trades based on given parameters
func (q *TradeAggregationsQ) GetSql() sq.SelectBuilder {
	var orderPreserved bool
	orderPreserved, q.baseAssetID, q.counterAssetID = getCanonicalAssetOrder(q.baseAssetID, q.counterAssetID)

	var bucketSQL sq.SelectBuilder
	if orderPreserved {
		bucketSQL = bucketTrades(q.resolution, q.offset)
	} else {
		bucketSQL = reverseBucketTrades(q.resolution, q.offset)
	}

	bucketSQL = bucketSQL.From("history_trades").
		Where(sq.Eq{"base_asset_id": q.baseAssetID, "counter_asset_id": q.counterAssetID})

	//adjust time range and apply time filters
	bucketSQL = bucketSQL.Where(sq.GtOrEq{"ledger_closed_at": q.startTime.ToTime()})
	if !q.endTime.IsNil() {
		bucketSQL = bucketSQL.Where(sq.Lt{"ledger_closed_at": q.endTime.ToTime()})
	}

	//ensure open/close order for cases when multiple trades occur in the same ledger
	bucketSQL = bucketSQL.OrderBy("history_operation_id ", "\"order\"")

	return sq.Select(
		"timestamp",
		"count(*) as count",
		"sum(base_amount) as base_volume",
		"sum(counter_amount) as counter_volume",
		"sum(counter_amount)/sum(base_amount) as avg",
		"max_price(price) as high",
		"min_price(price) as low",
		"first(price)  as open",
		"last(price) as close",
	).
		FromSelect(bucketSQL, "htrd").
		GroupBy("timestamp").
		Limit(q.pagingParams.Limit).
		OrderBy("timestamp " + q.pagingParams.Order)
}

// formatBucketTimestampSelect formats a sql select clause for a bucketed timestamp, based on given resolution
// and the offset. Given a time t, it gives it a timestamp defined by
// f(t) = ((t - offset)/resolution)*resolution + offset.
func formatBucketTimestampSelect(resolution int64, offset int64) string {
	return fmt.Sprintf("div((cast((extract(epoch from ledger_closed_at) * 1000 ) as bigint) - %d), %d)*%d + %d as timestamp",
		offset, resolution, resolution, offset)
}

// bucketTrades generates a select statement to filter rows from the `history_trades` table in
// a compact form, with a timestamp rounded to resolution and reversed base/counter.
func bucketTrades(resolution int64, offset int64) sq.SelectBuilder {
	return sq.Select(
		formatBucketTimestampSelect(resolution, offset),
		"history_operation_id",
		"\"order\"",
		"base_asset_id",
		"base_amount",
		"counter_asset_id",
		"counter_amount",
		"ARRAY[price_n, price_d] as price",
	)
}

// reverseBucketTrades generates a select statement to filter rows from the `history_trades` table in
// a compact form, with a timestamp rounded to resolution and reversed base/counter.
func reverseBucketTrades(resolution int64, offset int64) sq.SelectBuilder {
	return sq.Select(
		formatBucketTimestampSelect(resolution, offset),
		"history_operation_id",
		"\"order\"",
		"counter_asset_id as base_asset_id",
		"counter_amount as base_amount",
		"base_asset_id as counter_asset_id",
		"base_amount as counter_amount",
		"ARRAY[price_d, price_n] as price",
	)
}
