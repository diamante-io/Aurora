package aurora

import (
	"encoding/json"
	"testing"

	"github.com/diamnet/go/protocols/aurora"
)

func TestOrderBookActions_Show(t *testing.T) {
	ht := StartHTTPTest(t, "order_books")
	defer ht.Finish()

	var result aurora.OrderBookSummary

	// with no query args
	w := ht.Get("/order_book")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// missing currency
	w = ht.Get("/order_book?selling_asset_type=native")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// invalid type
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=nothing")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	w = ht.Get("/order_book?selling_asset_type=nothing&buying_asset_type=native")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// missing code
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_issuer=123")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	w = ht.Get("/order_book?buying_asset_type=native&selling_asset_type=credit_alphanum4&selling_asset_issuer=123")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// missing issuer
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_code=USD")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	w = ht.Get("/order_book?buying_asset_type=native&selling_asset_type=credit_alphanum4&selling_asset_code=USD")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// incomplete currency
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_code=USD")
	if ht.Assert.Equal(400, w.Code) {
		ht.Assert.ProblemType(w.Body, "invalid_order_book")
	}

	// same currency
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=native")
	if ht.Assert.Equal(200, w.Code) {
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)

		ht.Assert.Len(result.Asks, 0)
		ht.Assert.Len(result.Bids, 0)
	}

	// happy path with default limit
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_code=USD&buying_asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	if ht.Assert.Equal(200, w.Code) {
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)

		ht.Assert.Equal("native", result.Selling.Type)
		ht.Assert.Equal("", result.Selling.Code)
		ht.Assert.Equal("", result.Selling.Issuer)
		ht.Assert.Equal("credit_alphanum4", result.Buying.Type)
		ht.Assert.Equal("USD", result.Buying.Code)
		ht.Assert.Equal("GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4", result.Buying.Issuer)

		ht.Require.Len(result.Asks, 3)
		ht.Require.Len(result.Bids, 3)

		ht.Assert.Equal("100.0000000", result.Asks[0].Amount)
		ht.Assert.Equal("900.0000000", result.Asks[1].Amount)
		ht.Assert.Equal("5000.0000000", result.Asks[2].Amount)
		ht.Assert.Equal("10.0000000", result.Bids[0].Amount)
		ht.Assert.Equal("100.0000000", result.Bids[1].Amount)
		ht.Assert.Equal("1000.0000000", result.Bids[2].Amount)
	}

	// happy path with smaller limit
	w = ht.Get("/order_book?selling_asset_type=native&buying_asset_type=credit_alphanum4&buying_asset_code=USD&buying_asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4&limit=1")
	if ht.Assert.Equal(200, w.Code) {
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)

		ht.Assert.Equal("native", result.Selling.Type)
		ht.Assert.Equal("", result.Selling.Code)
		ht.Assert.Equal("", result.Selling.Issuer)
		ht.Assert.Equal("credit_alphanum4", result.Buying.Type)
		ht.Assert.Equal("USD", result.Buying.Code)
		ht.Assert.Equal("GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4", result.Buying.Issuer)

		ht.Require.Len(result.Asks, 1)
		ht.Require.Len(result.Bids, 1)

		ht.Assert.Equal("100.0000000", result.Asks[0].Amount)
		ht.Assert.Equal("10.0000000", result.Bids[0].Amount)
	}
}
