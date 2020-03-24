package core

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func TestGetOrderBookSummary(t *testing.T) {
	tt := test.Start(t).Scenario("order_books")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	selling, err := AssetFromDB(xdr.AssetTypeAssetTypeCreditAlphanum4, "USD", "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	tt.Require.NoError(err)
	buying, err := AssetFromDB(xdr.AssetTypeAssetTypeNative, "", "")
	tt.Require.NoError(err)

	var summary, inverted OrderBookSummary
	err = q.GetOrderBookSummary(&summary, selling, buying, 20)
	tt.Require.NoError(err)
	tt.Require.Len(summary, 6)
	err = q.GetOrderBookSummary(&inverted, buying, selling, 20)
	tt.Require.NoError(err)
	tt.Require.Len(inverted, 6)

	asks := summary.Asks()
	bids := summary.Bids()
	iasks := inverted.Asks()
	ibids := inverted.Bids()

	// Check that summary was loaded correct
	tt.Assert.Equal("100000000", asks[0].Amount)
	tt.Assert.Equal(int32(15), asks[0].Pricen)
	tt.Assert.Equal(int32(1), asks[0].Priced)

	tt.Assert.Equal("1000000000", asks[1].Amount)
	tt.Assert.Equal(int32(20), asks[1].Pricen)
	tt.Assert.Equal(int32(1), asks[1].Priced)

	tt.Assert.Equal("10000000000", asks[2].Amount)
	tt.Assert.Equal(int32(50), asks[2].Pricen)
	tt.Assert.Equal(int32(1), asks[2].Priced)

	tt.Assert.Equal("1000000000", bids[0].Amount)
	tt.Assert.Equal(int32(10), bids[0].Pricen)
	tt.Assert.Equal(int32(1), bids[0].Priced)

	tt.Assert.Equal("9000000000", bids[1].Amount)
	tt.Assert.Equal(int32(9), bids[1].Pricen)
	tt.Assert.Equal(int32(1), bids[1].Priced)

	tt.Assert.Equal("50000000000", bids[2].Amount)
	tt.Assert.Equal(int32(5), bids[2].Pricen)
	tt.Assert.Equal(int32(1), bids[2].Priced)

	// Check that the inversion was correct
	tt.Assert.Equal(asks[0].Pricef, ibids[0].InvertPricef())
	tt.Assert.Equal(asks[1].Pricef, ibids[1].InvertPricef())
	tt.Assert.Equal(asks[2].Pricef, ibids[2].InvertPricef())
	tt.Assert.Equal(bids[0].Pricef, iasks[0].InvertPricef())
	tt.Assert.Equal(bids[1].Pricef, iasks[1].InvertPricef())
	tt.Assert.Equal(bids[2].Pricef, iasks[2].InvertPricef())
}

// regression test for https://github.com/hcnet/go/services/aurora/internal/issues/310
func TestGetOrderBookSummary_Regress310(t *testing.T) {
	tt := test.Start(t).Scenario("order_books_310")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	selling, err := AssetFromDB(xdr.AssetTypeAssetTypeCreditAlphanum4, "USD", "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	tt.Require.NoError(err)
	buying, err := AssetFromDB(xdr.AssetTypeAssetTypeNative, "", "")
	tt.Require.NoError(err)

	var summary OrderBookSummary
	err = q.GetOrderBookSummary(&summary, selling, buying, 20)
	tt.Require.NoError(err)
	tt.Require.Len(summary, 20)

	// In the order_books_310 scenario, the orders were placed in such a way that
	// three orders at prices 10.2, 10.1, and 10.0 should appear first, when the
	// query is correct.  In a failing scenario the 10.2 transaction should not
	// appear, because it was inserted after the first 20 rows
	bids := summary.Bids()
	tt.Assert.Equal(10.2, bids[0].Pricef)
	tt.Assert.Equal(10.1, bids[1].Pricef)
	tt.Assert.Equal(10.0, bids[2].Pricef)

	// validate the inverse order book is correct as well
	err = q.GetOrderBookSummary(&summary, buying, selling, 20)
	tt.Require.NoError(err)
	tt.Require.Len(summary, 20)
	asks := summary.Asks()
	tt.Assert.Equal(1.0/10.2, asks[0].Pricef)
	tt.Assert.Equal(1.0/10.1, asks[1].Pricef)
	tt.Assert.Equal(1.0/10.0, asks[2].Pricef)
}

func TestGetOrderBookSummary_AmountsExceedInt64(t *testing.T) {
	tt := test.Start(t).Scenario("order_books")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	selling, err := AssetFromDB(xdr.AssetTypeAssetTypeCreditAlphanum4, "USD", "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	tt.Require.NoError(err)
	buying, err := AssetFromDB(xdr.AssetTypeAssetTypeNative, "", "")
	tt.Require.NoError(err)

	// Generate huge amount at a single price level.
	// Should be: 3 * 9131689504000000000 at both sides of the orderbook
	_, err = tt.CoreDB.Exec(
		`UPDATE offers SET amount = '9131689504000000000', pricen = 1, priced=1, price = 1`,
	)
	tt.Require.NoError(err)

	var summary OrderBookSummary
	err = q.GetOrderBookSummary(&summary, selling, buying, 20)
	tt.Require.NoError(err)
	tt.Assert.Len(summary, 2)
	asks := summary.Asks()
	bids := summary.Bids()
	tt.Assert.Equal("27395068512000000000", asks[0].Amount)
	tt.Assert.Equal("27395068512000000000", bids[0].Amount)
}
