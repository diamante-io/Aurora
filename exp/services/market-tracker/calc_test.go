package main

import (
	"math"
	"testing"

	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/stretchr/testify/assert"
)

var hLowOrder = hProtocol.PriceLevel{
	PriceR: hProtocol.Price{
		N: 4,
		D: 2,
	},
	Price:  "2.0",
	Amount: "50.0",
}

var hHighOrder = hProtocol.PriceLevel{
	PriceR: hProtocol.Price{
		N: 5,
		D: 2,
	},
	Price:  "2.5",
	Amount: "100.0",
}

var hOrders = []hProtocol.PriceLevel{hLowOrder, hHighOrder}

var lowUsdOrder = usdOrder{
	xlmAmount: 50.0,
	usdAmount: 5.0,
	usdPrice:  20.0,
}

var highUsdOrder = usdOrder{
	xlmAmount: 100.0,
	usdAmount: 10.0,
	usdPrice:  25.0,
}

func TestCalcSpreadPctForOrderBook(t *testing.T) {
	obStats := hProtocol.OrderBookSummary{
		Bids: hOrders,
		Asks: hOrders,
	}
	wantPct := 100.0 * (2.0 - 2.5) / 2.0
	gotPct := calcSpreadPctForOrderBook(obStats)
	assert.Equal(t, wantPct, gotPct)
}

func TestCalcSpread(t *testing.T) {
	// Test inputs that should result in 0 spread.
	wantSpread := 0.0
	highestBid := -1.0
	lowestAsk := 0.0
	gotSpread := calcSpread(highestBid, lowestAsk)
	assert.Equal(t, wantSpread, gotSpread)

	lowestAsk = -1.0
	highestBid = 0.0
	gotSpread = calcSpread(highestBid, lowestAsk)
	assert.Equal(t, wantSpread, gotSpread)

	highestBid = math.Inf(-1)
	gotSpread = calcSpread(highestBid, lowestAsk)
	assert.Equal(t, wantSpread, gotSpread)

	highestBid = -1.0
	lowestAsk = math.Inf(1)
	gotSpread = calcSpread(highestBid, lowestAsk)
	assert.Equal(t, wantSpread, gotSpread)

	// Test the spread calculation.
	lowestAsk = 10.0
	highestBid = 9.0
	wantSpread = (lowestAsk - highestBid) / lowestAsk
	gotSpread = calcSpread(highestBid, lowestAsk)
	assert.Equal(t, wantSpread, gotSpread)
}

func TestCalcSpreadPctAtDepth(t *testing.T) {
	bids := []usdOrder{
		usdOrder{
			xlmAmount: 100.0,
			usdAmount: 10.0,
			usdPrice:  10.0,
		},
	}
	asks := []usdOrder{lowUsdOrder, highUsdOrder}

	gotSpreadPct := calcSpreadPctAtDepth(bids, asks, 5.0)
	assert.Equal(t, 50.0, gotSpreadPct)

	gotSpreadPct = calcSpreadPctAtDepth(bids, asks, 10.0)
	assert.Equal(t, 60.0, gotSpreadPct)
}

func TestCalcHighestBid(t *testing.T) {
	// Test empty bids.
	gotBid := calcHighestBid([]hProtocol.PriceLevel{})
	assert.Equal(t, math.Inf(-1), gotBid)

	// Test non-empty bids.
	gotBid = calcHighestBid(hOrders)
	assert.Equal(t, 2.5, gotBid)
}

func TestCalcLowestAsk(t *testing.T) {
	// Test empty asks.
	asks := []hProtocol.PriceLevel{}
	wantAsk := math.Inf(1)
	gotAsk := calcLowestAsk(asks)
	assert.Equal(t, wantAsk, gotAsk)

	// Test non-empty asks.
	wantAsk = 2.0
	gotAsk = calcLowestAsk(hOrders)
	assert.Equal(t, wantAsk, gotAsk)
}

func TestCalcBestOrderAtDepth(t *testing.T) {
	bids := []usdOrder{highUsdOrder, lowUsdOrder}
	gotBid := calcBestOrderAtDepth(bids, 5.0)
	assert.Equal(t, 25.0, gotBid)

	gotBid = calcBestOrderAtDepth(bids, 15.0)
	assert.Equal(t, 20.0, gotBid)

	gotBid = calcBestOrderAtDepth(bids, 25.0)
	assert.Equal(t, 0.0, gotBid)
}

func TestCalcSlippageAtDepth(t *testing.T) {
	bids := []usdOrder{
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 30.,
			usdPrice:  30.,
		},
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 25.,
			usdPrice:  25.,
		},
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 50.,
			usdPrice:  20.,
		},
	}

	asks := []usdOrder{
		usdOrder{
			xlmAmount: 5.,
			usdAmount: 100.,
			usdPrice:  20.,
		},
		usdOrder{
			xlmAmount: 4.,
			usdAmount: 100.,
			usdPrice:  25.,
		},
		usdOrder{
			xlmAmount: 4.,
			usdAmount: 120.,
			usdPrice:  30.,
		},
	}

	bs := calcSlippageAtDepth(bids, asks, 10., true)
	assert.Equal(t, 20., bs)
	as := calcSlippageAtDepth(bids, asks, 10., false)
	assert.Equal(t, 20., as)

	bs = calcSlippageAtDepth(bids, asks, 50., true)
	assert.Equal(t, 10., bs)
	as = calcSlippageAtDepth(bids, asks, 50., false)
	assert.Equal(t, 20., as)

	bs = calcSlippageAtDepth(bids, asks, 100., true)
	assert.Equal(t, 0., bs)
	as = calcSlippageAtDepth(bids, asks, 100., false)
	assert.Equal(t, 20., as)

	bs = calcSlippageAtDepth(bids, asks, 1000., true)
	assert.Equal(t, 0., bs)
	as = calcSlippageAtDepth(bids, asks, 1000., false)
	assert.Equal(t, 0., as)
}

func TestCalcMidPrice(t *testing.T) {
	bids := []usdOrder{highUsdOrder, lowUsdOrder}
	asks := []usdOrder{lowUsdOrder, highUsdOrder}
	mp := calcMidPrice(bids, asks)
	assert.Equal(t, 22.5, mp)
}

func TestCalcAvgPriceAtDepth(t *testing.T) {
	orders := []usdOrder{highUsdOrder, lowUsdOrder}
	ap := calcAvgPriceAtDepth(orders, 10.)
	assert.Equal(t, 25., ap)
	ap = calcAvgPriceAtDepth(orders, 15.)
	assert.Equal(t, 22.5, ap)
	ap = calcAvgPriceAtDepth(orders, 20.)
	assert.Equal(t, 0., ap)
}

func TestCalcFairValuePct(t *testing.T) {
	bids := []usdOrder{
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 30.,
			usdPrice:  30.,
		},
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 25.,
			usdPrice:  25.,
		},
		usdOrder{
			xlmAmount: 1.,
			usdAmount: 50.,
			usdPrice:  20.,
		},
	}

	asks := []usdOrder{
		usdOrder{
			xlmAmount: 5.,
			usdAmount: 100.,
			usdPrice:  20.,
		},
		usdOrder{
			xlmAmount: 4.,
			usdAmount: 100.,
			usdPrice:  25.,
		},
		usdOrder{
			xlmAmount: 4.,
			usdAmount: 120.,
			usdPrice:  30.,
		},
	}

	trueAssetUsdPrice := 0.
	pct := calcFairValuePct(bids, asks, trueAssetUsdPrice)
	assert.Equal(t, 0., pct)

	trueAssetUsdPrice = 20.
	pct = calcFairValuePct(bids, asks, trueAssetUsdPrice)
	assert.Equal(t, 25., pct)

	trueAssetUsdPrice = 25.
	pct = calcFairValuePct(bids, asks, trueAssetUsdPrice)
	assert.Equal(t, 0., pct)
}
