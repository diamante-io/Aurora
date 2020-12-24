package simplepath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/diamnet/go/xdr"
)

func TestOrderBook(t *testing.T) {
	tt := test.Start(t).Scenario("paths")
	defer tt.Finish()

	ob := orderBook{
		Selling: makeAsset(
			xdr.AssetTypeAssetTypeCreditAlphanum4,
			"EUR",
			"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"),
		Buying: makeAsset(
			xdr.AssetTypeAssetTypeCreditAlphanum4,
			"USD",
			"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"),
		Q: &core.Q{Session: tt.CoreSession()},
	}

	testCases := []struct {
		scenario    string
		eur         int64
		wantCostUSD int64
	}{
		{"first unit", 2, 1},                                // taking from the first offer, where the price is 0.25 (i.e. offer to sell 4 EUR for 1 USD)
		{"first full offer", 100000000, 50000000},           // taking all from first offer (p=0.25)
		{"first full offer + 1", 100000002, 50000001},       // taking all from first offer (p=0.25), and first full unit of second offer (p=0.5)
		{"first two full offers", 200000000, 100000000},     // taking all from first two offers (p=0.25, p=0.5)
		{"first two full offers + 1", 200000001, 100000001}, // taking all from first two offers (p=0.25, p=0.5), and first full unit of third offer (p=1.0)
		{"first three full offers", 300000000, 200000000},   // taking all from first three offers (p=0.25, p=0.5, p = 1.0)
	}

	for _, kase := range testCases {
		t.Run(kase.scenario, func(t *testing.T) {
			r, err := ob.CostToConsumeLiquidity(xdr.Int64(kase.eur))
			if tt.Assert.NoError(err) {
				tt.Assert.Equal(xdr.Int64(kase.wantCostUSD), r)
			}
		})
	}

	// taking 1 more than there is available
	t.Run("one more than available liquidity", func(t *testing.T) {
		_, err := ob.CostToConsumeLiquidity(xdr.Int64(300000001))
		tt.Assert.Error(err)
	})
}

func TestOrderBook_BadCost(t *testing.T) {
	tt := test.Start(t).Scenario("bad_cost")
	defer tt.Finish()

	ob := orderBook{
		Selling: makeAsset(
			xdr.AssetTypeAssetTypeCreditAlphanum4,
			"EUR",
			"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"),
		Buying: makeAsset(
			xdr.AssetTypeAssetTypeCreditAlphanum4,
			"USD",
			"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN"),
		Q: &core.Q{Session: tt.CoreSession()},
	}

	r, err := ob.CostToConsumeLiquidity(2000000000)
	if tt.Assert.NoError(err) {
		tt.Assert.Equal(xdr.Int64(10000000), r)
	}
}

func TestWillAddOverflow(t *testing.T) {
	testCases := []struct {
		a                int64
		b                int64
		wantWillOverflow bool
	}{
		{1, 2, false},
		{0, 1, false},
		{math.MaxInt64, 0, false},
		{math.MaxInt64 - 1, 1, false},
		{math.MaxInt64, 1, true},
		{math.MaxInt64 - 1, 2, true},
		{math.MaxInt64 - 1, math.MaxInt64, true},
		{math.MaxInt64, math.MaxInt64, true},
	}
	for _, kase := range testCases {
		t.Run(t.Name(), func(t *testing.T) {
			r := willAddOverflow(kase.a, kase.b)
			assert.Equal(t, kase.wantWillOverflow, r)
		})
	}
}
