package main

import (
	"math"

	hProtocol "github.com/diamnet/go/protocols/aurora"
)

func calcSpreadPctForOrderBook(obStats hProtocol.OrderBookSummary) float64 {
	highestBid := calcHighestBid(obStats.Bids)
	lowestAsk := calcLowestAsk(obStats.Asks)
	spread := calcSpread(highestBid, lowestAsk)
	spreadPct := 100.0 * spread
	return spreadPct
}

func calcSpread(highestBid float64, lowestAsk float64) float64 {
	if lowestAsk == 0 || highestBid == 0 || math.IsInf(highestBid, -1) || math.IsInf(lowestAsk, 1) {
		return 0
	}
	return (lowestAsk - highestBid) / lowestAsk
}

func calcSpreadPctAtDepth(bids, asks []usdOrder, depth float64) float64 {
	highestBid := calcBestOrderAtDepth(bids, depth)
	lowestAsk := calcBestOrderAtDepth(asks, depth)
	spread := calcSpread(highestBid, lowestAsk)
	spreadPct := 100.0 * spread
	return spreadPct
}

func calcHighestBid(bids []hProtocol.PriceLevel) float64 {
	highestBid := math.Inf(-1)
	for _, bid := range bids {
		currBid := float64(bid.PriceR.N) / float64(bid.PriceR.D)
		if currBid > highestBid {
			highestBid = currBid
		}
	}
	return highestBid
}

func calcLowestAsk(bids []hProtocol.PriceLevel) float64 {
	lowestAsk := math.Inf(1)
	for _, bid := range bids {
		currBid := float64(bid.PriceR.N) / float64(bid.PriceR.D)
		if currBid < lowestAsk {
			lowestAsk = currBid
		}
	}
	return lowestAsk
}

func calcBestOrderAtDepth(orders []usdOrder, depth float64) float64 {
	total := 0.0
	for _, order := range orders {
		total += order.usdAmount
		if total >= depth {
			return order.usdPrice
		}
	}

	// We return 0.0 in case of insufficient depth.
	return 0.0
}

func calcSlippageAtDepth(bids, asks []usdOrder, depth float64, isBid bool) float64 {
	mp := calcMidPrice(bids, asks)

	var ap float64
	if isBid {
		ap = calcAvgPriceAtDepth(bids, depth)
	} else {
		ap = calcAvgPriceAtDepth(asks, depth)
	}

	// if insufficient liquidity, return 0.
	if ap == 0 {
		return 0
	}

	return 100. * math.Abs(ap-mp) / mp
}

func calcMidPrice(bids, asks []usdOrder) float64 {
	// note that this assumes bids and asks have already been sorted,
	// which is done in the convertBids / convertAsks functions
	highestBid := calcBestOrderAtDepth(bids, 0.)
	lowestAsk := calcBestOrderAtDepth(asks, 0.)
	midPrice := (highestBid + lowestAsk) / 2
	return midPrice
}

func calcAvgPriceAtDepth(orders []usdOrder, depth float64) float64 {
	// note that this assumes bids and asks have already been sorted,
	// which is done in the convertBids / convertAsks functions
	totalPrice := 0.0
	totalAmount := 0.0
	for i, order := range orders {
		totalPrice += order.usdPrice
		totalAmount += order.usdAmount
		if totalAmount >= depth {
			return totalPrice / float64(i+1)
		}
	}
	return 0.0
}

func calcFairValuePct(sortedBids, sortedAsks []usdOrder, trueAssetUsdPrice float64) float64 {
	if trueAssetUsdPrice == 0 {
		return 0.
	}

	dexAssetUsdPrice := calcMidPrice(sortedBids, sortedAsks)
	fairValuePct := 100. * math.Abs(dexAssetUsdPrice-trueAssetUsdPrice) / trueAssetUsdPrice
	return fairValuePct
}
