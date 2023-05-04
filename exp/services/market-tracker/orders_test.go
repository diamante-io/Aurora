package main

import (
	"testing"

	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/stretchr/testify/assert"
)

var badAmtOrders = []hProtocol.PriceLevel{hProtocol.PriceLevel{
	PriceR: hProtocol.Price{
		N: 4,
		D: 2,
	},
	Price:  "2.0",
	Amount: "amount",
}}

func TestConvertBids(t *testing.T) {
	usdXlmPrice := 0.10
	basePrice := 0.10
	bids, err := convertBids(badAmtOrders, usdXlmPrice, basePrice)
	assert.Error(t, err)
	assert.Equal(t, 0, len(bids))

	highBid := usdOrder{
		xlmAmount:  100.0,
		usdAmount:  10.0,
		baseAmount: 1.0,
		usdPrice:   0.25,
	}

	lowBid := usdOrder{
		xlmAmount:  50.0,
		usdAmount:  5.0,
		baseAmount: 0.5,
		usdPrice:   0.2,
	}

	bids, err = convertBids(hOrders, usdXlmPrice, basePrice)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, bids[0].usdPrice, bids[1].usdPrice)
	assert.Equal(t, highBid, bids[0])
	assert.Equal(t, lowBid, bids[1])
}

func TestConvertAsks(t *testing.T) {
	usdXlmPrice := 0.10
	basePrice := 0.10
	asks, err := convertAsks(badAmtOrders, usdXlmPrice, basePrice)
	assert.Error(t, err)
	assert.Equal(t, 0, len(asks))

	lowAsk := usdOrder{
		xlmAmount:  50,
		usdPrice:   0.2,
		usdAmount:  5,
		baseAmount: 0.5,
	}

	highAsk := usdOrder{
		xlmAmount:  100,
		usdPrice:   0.25,
		usdAmount:  10,
		baseAmount: 1,
	}

	orders := []hProtocol.PriceLevel{hHighOrder, hLowOrder}
	asks, err = convertAsks(orders, usdXlmPrice, basePrice)
	assert.NoError(t, err)
	assert.LessOrEqual(t, asks[0].usdPrice, asks[1].usdPrice)
	assert.Equal(t, lowAsk, asks[0])
	assert.Equal(t, highAsk, asks[1])
}
