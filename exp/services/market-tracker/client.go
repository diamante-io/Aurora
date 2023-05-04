package main

import (
	"math"
	"time"

	hClient "github.com/diamnet/go/clients/auroraclient"
	hProtocol "github.com/diamnet/go/protocols/aurora"
)

type trackerClient struct {
	aurora *hClient.Client
}

func (tc *trackerClient) computeSpreadForTradePair(tp TradePair) (spread float64, err error) {
	obStats, err := tc.getOrderBookForTradePair(tp)
	if err != nil {
		return 0.0, err
	}

	highestBid := calcHighestBid(obStats.Bids)
	lowestAsk := calcLowestAsk(obStats.Asks)
	spread = calcSpread(highestBid, lowestAsk)
	return
}

func (tc *trackerClient) getOrderBookForTradePair(tp TradePair) (obStats hProtocol.OrderBookSummary, err error) {
	req := hClient.OrderBookRequest{
		SellingAssetType:   tp.SellingAsset.ProtocolAssetType,
		SellingAssetCode:   tp.SellingAsset.Code,
		SellingAssetIssuer: tp.SellingAsset.IssuerAddress,
		BuyingAssetType:    tp.BuyingAsset.ProtocolAssetType,
		BuyingAssetCode:    tp.BuyingAsset.Code,
		BuyingAssetIssuer:  tp.BuyingAsset.IssuerAddress,
		Limit:              200,
	}
	obStats, err = tc.aurora.OrderBook(req)
	return
}

func (tc *trackerClient) getAggTradesForTradePair(tp TradePair, start, end time.Time, res time.Duration) (taps []hProtocol.TradeAggregationsPage, err error) {
	const maxLimit = 200
	req := hClient.TradeAggregationRequest{
		StartTime:          start,
		EndTime:            end,
		Resolution:         res,
		Offset:             time.Duration(0),
		BaseAssetType:      tp.BuyingAsset.ProtocolAssetType,
		BaseAssetCode:      tp.BuyingAsset.Code,
		BaseAssetIssuer:    tp.BuyingAsset.IssuerAddress,
		CounterAssetType:   tp.SellingAsset.ProtocolAssetType,
		CounterAssetCode:   tp.SellingAsset.Code,
		CounterAssetIssuer: tp.SellingAsset.IssuerAddress,
		Limit:              maxLimit,
	}

	tap, err := tc.aurora.TradeAggregations(req)
	if err != nil {
		return
	}
	taps = append(taps, tap)

	// iterate through the remaining trade aggregations as needed, until error
	// TODO: Check error for Aurora rate limiting, before adding volume metrics.
	counter := 1
	numRes := float64(end.Sub(start) / res)
	numRequests := int(math.Ceil(numRes / maxLimit))
	for counter < numRequests {
		tap, err = tc.aurora.NextTradeAggregationsPage(tap)
		if err != nil {
			return
		}

		counter++
		taps = append(taps, tap)
	}
	return
}
