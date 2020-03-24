package aurora

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	. "github.com/hcnet/go/services/aurora/internal/db2/history"
	. "github.com/hcnet/go/services/aurora/internal/test/trades"
	"github.com/hcnet/go/support/render/hal"
	hcnetTime "github.com/hcnet/go/support/time"
	"github.com/hcnet/go/xdr"
)

func TestTradeActions_Index(t *testing.T) {
	ht := StartHTTPTest(t, "trades")
	defer ht.Finish()
	var records []aurora.Trade
	var firstTrade aurora.Trade

	// All trades
	w := ht.Get("/trades")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)

		// 	ensure created_at is populated correctly
		ht.UnmarshalPage(w.Body, &records)
		firstTrade = records[0]

		// 	ensure created_at is populated correctly
		l := history.Ledger{}
		hq := history.Q{Session: ht.AuroraSession()}
		ht.Require.NoError(hq.LedgerBySequence(&l, 9))

		ht.Assert.WithinDuration(l.ClosedAt, records[0].LedgerCloseTime, 1*time.Second)
	}

	// reverse order
	w = ht.Get("/trades?order=desc")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)
		ht.UnmarshalPage(w.Body, &records)

		// ensure that ordering is indeed reversed
		ht.Assert.Equal(firstTrade, records[len(records)-1])
	}

	var q = make(url.Values)
	q.Add("base_asset_type", "credit_alphanum4")
	q.Add("base_asset_code", "USD")
	q.Add("base_asset_issuer", "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	q.Add("counter_asset_type", "credit_alphanum4")
	q.Add("counter_asset_code", "EUR")
	q.Add("counter_asset_issuer", "GCQPYGH4K57XBDENKKX55KDTWOTK5WDWRQOH2LHEDX3EKVIQRLMESGBG")

	w = ht.GetWithParams("/trades", q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)

		records := []map[string]interface{}{}
		ht.UnmarshalPage(w.Body, &records)

		ht.Assert.Contains(records[0], "base_amount")
		ht.Assert.Contains(records[0], "counter_amount")
	}

	q = make(url.Values)
	q.Add("base_asset_type", "credit_alphanum4")
	q.Add("base_asset_code", "EUR")
	q.Add("base_asset_issuer", "GCQPYGH4K57XBDENKKX55KDTWOTK5WDWRQOH2LHEDX3EKVIQRLMESGBG")
	q.Add("counter_asset_type", "credit_alphanum4")
	q.Add("counter_asset_code", "USD")
	q.Add("counter_asset_issuer", "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")

	w = ht.GetWithParams("/trades", q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)

		records := []map[string]interface{}{}
		ht.UnmarshalPage(w.Body, &records)

		ht.Assert.Contains(records[0], "base_amount")
		ht.Assert.Contains(records[0], "counter_amount")
	}

	// For offer
	w = ht.Get("/offers/1/trades")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)
	}

	w = ht.Get("/offers/2/trades")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(0, w.Body)
	}

	// for an account
	w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/trades")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)
	}

	// for other account
	w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/trades")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)
		records := []map[string]interface{}{}
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.Contains(records[0], "base_amount")
		ht.Assert.Contains(records[0], "counter_amount")
	}

	//test paging from account 1
	w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/trades?order=desc&limit=1")
	var links hal.Links
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		links = ht.UnmarshalPage(w.Body, &records)
	}

	w = ht.Get(links.Next.Href)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		prevRecord := records[0]
		links = ht.UnmarshalPage(w.Body, &records)
		ht.Assert.NotEqual(prevRecord, records[0])
	}

	//test paging from account 2
	w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/trades?order=desc&limit=1")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		links = ht.UnmarshalPage(w.Body, &records)
	}

	w = ht.Get(links.Next.Href)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		prevRecord := records[0]
		links = ht.UnmarshalPage(w.Body, &records)
		ht.Assert.NotEqual(prevRecord, records[0])
	}
}

// setAssetQuery adds an asset filter with a given prefix to a query
func setAssetQuery(q *url.Values, prefix string, asset xdr.Asset) {
	var assetType, assetCode, assetFilter string
	asset.Extract(&assetType, &assetCode, &assetFilter)
	q.Add(prefix+"asset_type", assetType)
	q.Add(prefix+"asset_code", assetCode)
	q.Add(prefix+"asset_issuer", assetFilter)
}

//testPrice ensures that the price float string is equal to the rational price
func testPrice(t *HTTPT, priceStr string, priceR xdr.Price) {
	price, err := strconv.ParseFloat(priceStr, 64)
	if t.Assert.NoError(err) {
		t.Assert.Equal(price, float64(priceR.N)/float64(priceR.D))
	}
}

func testTradeAggregationPrices(t *HTTPT, record aurora.TradeAggregation) {
	testPrice(t, record.High, record.HighR)
	testPrice(t, record.Low, record.LowR)
	testPrice(t, record.Open, record.OpenR)
	testPrice(t, record.Close, record.CloseR)
}

const minute = int64(time.Minute / time.Millisecond)
const hour = int64(time.Hour / time.Millisecond)
const day = int64(24 * time.Hour / time.Millisecond)
const week = int64(7 * 24 * time.Hour / time.Millisecond)
const aggregationPath = "/trade_aggregations"

func TestTradeActions_Aggregation(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	const numOfTrades = 10

	//a realistic millis (since epoch) value to start the test from
	//it represents a round hour and is bigger than a max int32
	const start = int64(1510693200000)

	dbQ := &Q{ht.AuroraSession()}
	ass1, ass2, err := PopulateTestTrades(dbQ, start, numOfTrades, minute, 0)
	ht.Require.NoError(err)

	//add other trades as noise, to ensure asset filtering is working
	_, _, err = PopulateTestTrades(dbQ, start, numOfTrades, minute, numOfTrades)
	ht.Require.NoError(err)

	var records []aurora.TradeAggregation
	var record aurora.TradeAggregation
	var nextLink string

	q := make(url.Values)
	setAssetQuery(&q, "base_", ass1)
	setAssetQuery(&q, "counter_", ass2)

	q.Add("start_time", strconv.FormatInt(start, 10))
	q.Add("end_time", strconv.FormatInt(start+hour, 10))
	q.Add("order", "asc")

	//test no resolution provided
	w := ht.GetWithParams(aggregationPath, q)
	ht.Assert.Equal(400, w.Code)

	//test illegal resolution
	if history.StrictResolutionFiltering {
		q.Add("resolution", strconv.FormatInt(hour/2, 10))
		w = ht.GetWithParams(aggregationPath, q)
		ht.Assert.Equal(400, w.Code)
	}

	//test one bucket for all trades
	q.Set("resolution", strconv.FormatInt(hour, 10))
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		ht.UnmarshalPage(w.Body, &records)
		record = records[0] //Save the single aggregation record for next test
		testTradeAggregationPrices(ht, record)
		ht.Assert.Equal("0.0005500", records[0].BaseVolume)
	}

	//test reverse one bucket - make sure values don't change
	q.Set("order", "desc")
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.Equal(record, records[0])
	}

	//Test bucket per trade
	q.Set("order", "asc")
	q.Set("resolution", strconv.FormatInt(minute, 10))
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		if ht.Assert.PageOf(numOfTrades, w.Body) {
			//test that asset filters work
			ht.UnmarshalPage(w.Body, &records)
			ht.Assert.Equal(int64(1), records[0].TradeCount)
			ht.Assert.Equal("0.0000100", records[0].BaseVolume)
			ht.Assert.Equal("1.0000000", records[0].Average)
		}
	}

	//test partial range by modifying endTime to be one minute above half range.
	//half of the results are expected
	endTime := start + (numOfTrades/2)*minute
	q.Set("end_time", strconv.FormatInt(endTime, 10))
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(numOfTrades/2, w.Body)
	}

	//test that page limit works
	limit := 3
	q.Add("limit", strconv.Itoa(limit))
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(limit, w.Body)
	}

	//test that next page delivers the correct amount of records
	w = ht.GetWithParams(aggregationPath, q)
	nextLink = ht.UnmarshalNext(w.Body)
	//make sure the next link is a full url and not just a path
	ht.Assert.Equal(true, strings.HasPrefix(nextLink, "http://localhost"))
	w = ht.Get(nextLink)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(numOfTrades/2-limit, w.Body)
		ht.UnmarshalPage(w.Body, &records)
		//test for expected value on timestamp of first record on next page
		ht.Assert.Equal(start+int64(limit)*minute, records[0].Timestamp)
	}

	//test direction (desc)
	q.Set("order", "desc")
	w = ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		if ht.Assert.PageOf(limit, w.Body) {
			ht.UnmarshalPage(w.Body, &records)
			ht.Assert.Equal(int64(start+(numOfTrades/2-1)*minute), records[0].Timestamp)
		}
	}

	//test next link desc
	w = ht.GetWithParams(aggregationPath, q)
	nextLink = ht.UnmarshalNext(w.Body)
	w = ht.Get(nextLink)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(numOfTrades/2-limit, w.Body)
	}

	//test next next link empty
	w = ht.GetWithParams(aggregationPath, q)
	nextLink = ht.UnmarshalNext(w.Body)
	w = ht.Get(nextLink)
	nextLink = ht.UnmarshalNext(w.Body)
	w = ht.Get(nextLink)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(0, w.Body)
	}
}

func TestTradeActions_AmountsExceedInt64(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()
	dbQ := &Q{ht.AuroraSession()}

	const start = int64(1510693200000)

	acc1 := GetTestAccount()
	acc2 := GetTestAccount()
	ass1 := GetTestAsset("usd")
	ass2 := GetTestAsset("euro")
	for i := 1; i <= 3; i++ {
		timestamp := hcnetTime.MillisFromInt64(start + (minute * int64(i-1)))
		err := IngestTestTrade(
			dbQ, ass1, ass2, acc1, acc2, int64(9131689504000000000), int64(9131689504000000000), timestamp, int64(i))
		ht.Require.NoError(err)
	}

	var records []aurora.TradeAggregation

	q := make(url.Values)
	setAssetQuery(&q, "base_", ass1)
	setAssetQuery(&q, "counter_", ass2)

	q.Add("start_time", strconv.FormatInt(start, 10))
	q.Add("end_time", strconv.FormatInt(start+hour, 10))
	q.Add("order", "asc")
	q.Set("resolution", strconv.FormatInt(hour, 10))

	w := ht.GetWithParams(aggregationPath, q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.Equal("2739506851200.0000000", records[0].BaseVolume)
		ht.Assert.Equal("2739506851200.0000000", records[0].CounterVolume)
	}
}

func TestTradeActions_IndexRegressions(t *testing.T) {
	t.Run("Regression:  https://github.com/hcnet/go/services/aurora/internal/issues/318", func(t *testing.T) {
		ht := StartHTTPTest(t, "trades")
		defer ht.Finish()

		var q = make(url.Values)
		q.Add("base_asset_type", "credit_alphanum4")
		q.Add("base_asset_code", "EUR")
		q.Add("base_asset_issuer", "GCQPYGH4K57XBDENKKX55KDTWOTK5WDWRQOH2LHEDX3EKVIQRLMESGBG")
		q.Add("counter_asset_type", "native")

		w := ht.Get("/trades?" + q.Encode())

		ht.Assert.Equal(404, w.Code) //This used to be 200 with length 0
	})

	t.Run("Regression for nil prices: https://github.com/hcnet/go/issues/357", func(t *testing.T) {
		ht := StartHTTPTest(t, "trades")
		defer ht.Finish()

		w := ht.Get("/trades")
		ht.Require.Equal(200, w.Code)

		_ = ht.AuroraDB.MustExec("UPDATE history_trades SET price_n = NULL, price_d = NULL")
		w = ht.Get("/trades")
		ht.Assert.Equal(200, w.Code, "nil-price trades failed")
	})
}

// TestTradeActions_AggregationOrdering checks that open/close aggregation
// fields are correct for multiple trades that occur in the same ledger
// https://github.com/hcnet/go/issues/215
func TestTradeActions_AggregationOrdering(t *testing.T) {

	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	seller := GetTestAccount()
	buyer := GetTestAccount()
	ass1 := GetTestAsset("usd")
	ass2 := GetTestAsset("euro")

	dbQ := &Q{ht.AuroraSession()}
	IngestTestTrade(dbQ, ass1, ass2, seller, buyer, 1, 3, 0, 3)
	IngestTestTrade(dbQ, ass1, ass2, seller, buyer, 1, 1, 0, 1)
	IngestTestTrade(dbQ, ass1, ass2, seller, buyer, 1, 2, 0, 2)

	q := make(url.Values)
	setAssetQuery(&q, "base_", ass1)
	setAssetQuery(&q, "counter_", ass2)

	q.Add("start_time", "0")
	q.Add("end_time", "60000")
	q.Add("order", "asc")
	q.Add("resolution", "60000")

	var records []aurora.TradeAggregation
	w := ht.GetWithParams("/trade_aggregations", q)
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.Equal("1.0000000", records[0].Open)
		ht.Assert.Equal("3.0000000", records[0].Close)
	}
}

func assertOfferType(ht *HTTPT, offerId string, idType OfferIDType) {
	offerIdInt64, _ := strconv.ParseInt(offerId, 10, 64)
	_, offerType := DecodeOfferID(offerIdInt64)
	ht.Assert.Equal(offerType, idType)
}

// TestTradeActions_SyntheticOfferIds loads the offer_ids scenario and ensures that synthetic offer
// ids are created when necessary and not when unnecessary
func TestTradeActions_SyntheticOfferIds(t *testing.T) {
	ht := StartHTTPTest(t, "offer_ids")
	defer ht.Finish()
	var records []aurora.Trade
	w := ht.Get("/trades")
	if ht.Assert.Equal(200, w.Code) {
		if ht.Assert.PageOf(4, w.Body) {
			ht.UnmarshalPage(w.Body, &records)
			assertOfferType(ht, records[0].BaseOfferID, TOIDType)
			assertOfferType(ht, records[1].BaseOfferID, TOIDType)
			assertOfferType(ht, records[2].BaseOfferID, CoreOfferIDType)
			assertOfferType(ht, records[3].BaseOfferID, CoreOfferIDType)
		}
	}
}

func TestTradeActions_AssetValidation(t *testing.T) {
	ht := StartHTTPTest(t, "trades")
	defer ht.Finish()

	var q = make(url.Values)
	q.Add("base_asset_type", "native")

	w := ht.GetWithParams("/trades", q)
	ht.Assert.Equal(400, w.Code)

	extras := ht.UnmarshalExtras(w.Body)
	ht.Assert.Equal("base_asset_type,counter_asset_type", extras["invalid_field"])
	ht.Assert.Equal("this endpoint supports asset pairs but only one asset supplied", extras["reason"])
}

func TestTradeActions_AggregationInvalidOffset(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()
	dbQ := &Q{ht.AuroraSession()}
	ass1, ass2, err := PopulateTestTrades(dbQ, 0, 100, hour, 1)
	ht.Require.NoError(err)

	q := make(url.Values)
	setAssetQuery(&q, "base_", ass1)
	setAssetQuery(&q, "counter_", ass2)
	q.Add("order", "asc")

	testCases := []struct {
		offset     int64
		resolution int64
		startTime  int64
		endTime    int64
	}{
		{offset: minute, resolution: hour},                                            // Test invalid offset value that's not hour aligned
		{offset: 25 * hour, resolution: week},                                         // Test invalid offset value that's greater than 24 hours
		{offset: 3 * hour, resolution: hour},                                          // Test invalid offset value that's greater than the resolution
		{offset: 3 * hour, startTime: 28 * hour, endTime: 26 * hour, resolution: day}, // Test invalid end time that's less than the start time
		{offset: 3 * hour, startTime: 6 * hour, endTime: 26 * hour, resolution: day},  // Test invalid end time that's less than the offset-adjusted start time
		{offset: 1 * hour, startTime: 5 * hour, endTime: 3 * hour, resolution: day},   // Test invalid end time that's less than the offset-adjusted start time
		{offset: 3 * hour, endTime: 1 * hour, resolution: day},                        // Test invalid end time that's less than the offset
		{startTime: 3 * minute, endTime: 1 * minute, resolution: minute},              // Test invalid end time that's less than the start time (no offset)
	}

	for _, tc := range testCases {
		t.Run("Testing invalid offset parameters", func(t *testing.T) {
			q.Add("offset", strconv.FormatInt(tc.offset, 10))
			q.Add("resolution", strconv.FormatInt(tc.resolution, 10))
			q.Add("start_time", strconv.FormatInt(tc.startTime, 10))
			if tc.endTime != 0 {
				q.Add("end_time", strconv.FormatInt(tc.endTime, 10))
			}
			w := ht.GetWithParams(aggregationPath, q)
			ht.Assert.Equal(400, w.Code)
		})
	}
}

func TestTradeActions_AggregationOffset(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()
	dbQ := &Q{ht.AuroraSession()}
	// One trade every hour
	ass1, ass2, err := PopulateTestTrades(dbQ, 0, 100, hour, 1)
	ht.Require.NoError(err)

	q := make(url.Values)
	setAssetQuery(&q, "base_", ass1)
	setAssetQuery(&q, "counter_", ass2)
	q.Add("order", "asc")

	q.Set("resolution", strconv.FormatInt(day, 10))
	testCases := []struct {
		offset             int64
		startTime          int64
		endTime            int64
		expectedTimestamps []int64
	}{
		{offset: 2 * hour, expectedTimestamps: []int64{2 * hour, 26 * hour, 50 * hour, 74 * hour, 98 * hour}}, //Test with no start time
		{offset: 1 * hour, startTime: 25 * hour, expectedTimestamps: []int64{25 * hour, 49 * hour, 73 * hour, 97 * hour}},
		{offset: 3 * hour, startTime: 10 * hour, expectedTimestamps: []int64{27 * hour, 51 * hour, 75 * hour, 99 * hour}},
		{offset: 6 * hour, startTime: 1 * hour, expectedTimestamps: []int64{6 * hour, 30 * hour, 54 * hour, 78 * hour}},
		{offset: 18 * hour, startTime: 30 * hour, expectedTimestamps: []int64{42 * hour, 66 * hour, 90 * hour}},
		{offset: 10 * hour, startTime: 35 * hour, expectedTimestamps: []int64{58 * hour, 82 * hour}},
		{offset: 18 * hour, startTime: 96 * hour, expectedTimestamps: []int64{}}, // No results since last timestamp is at 100
		{offset: 1 * hour, startTime: 5 * hour, endTime: 95 * hour, expectedTimestamps: []int64{25 * hour, 49 * hour}},
		{offset: 1 * hour, startTime: 5 * hour, endTime: 26 * hour, expectedTimestamps: []int64{}}, // end time and start time should both be at 25 hours
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing trade aggregations bucket with offset %d (hour) start time %d (hour)",
			tc.offset/hour, tc.startTime/hour), func(t *testing.T) {
			q.Set("offset", strconv.FormatInt(tc.offset, 10))
			if tc.startTime != 0 {
				q.Set("start_time", strconv.FormatInt(tc.startTime, 10))
			}
			if tc.endTime != 0 {
				q.Set("end_time", strconv.FormatInt(tc.endTime, 10))
			}
			w := ht.GetWithParams(aggregationPath, q)
			if ht.Assert.Equal(200, w.Code) {
				ht.Assert.PageOf(len(tc.expectedTimestamps), w.Body)
				var records []aurora.TradeAggregation
				ht.UnmarshalPage(w.Body, &records)
				if len(records) > 0 {
					for i, record := range records {
						ht.Assert.Equal(tc.expectedTimestamps[i], record.Timestamp)
					}
				}
			}
		})
	}
}
