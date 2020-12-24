package aurora

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/diamnet/go/xdr"
)

func TestOfferActions_Show(t *testing.T) {
	var (
		issuer = xdr.MustAddress("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")

		nativeAsset = xdr.Asset{
			Type: xdr.AssetTypeAssetTypeNative,
		}

		eurAsset = xdr.Asset{
			Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
			AlphaNum4: &xdr.AssetAlphaNum4{
				AssetCode: [4]byte{'e', 'u', 'r', 0},
				Issuer:    issuer,
			},
		}
		eurOffer = xdr.OfferEntry{
			SellerId: issuer,
			OfferId:  xdr.Int64(4),
			Buying:   eurAsset,
			Selling:  nativeAsset,
			Price: xdr.Price{
				N: 1,
				D: 1,
			},
			Flags:  1,
			Amount: xdr.Int64(500),
		}
		twoEurOffer = xdr.OfferEntry{
			SellerId: issuer,
			OfferId:  xdr.Int64(5),
			Buying:   eurAsset,
			Selling:  nativeAsset,
			Price: xdr.Price{
				N: 2,
				D: 1,
			},
			Flags:  2,
			Amount: xdr.Int64(500),
		}
	)

	ht := StartHTTPTest(t, "base")
	ht.App.config.EnableExperimentalIngestion = true
	defer ht.Finish()
	q := &history.Q{ht.AuroraSession()}

	ht.Assert.NoError(q.UpdateLastLedgerExpIngest(3))
	ht.Assert.NoError(q.UpsertOffer(eurOffer, 3))
	ht.Assert.NoError(q.UpsertOffer(twoEurOffer, 20))

	w := ht.Get(fmt.Sprintf("/offers/%v", eurOffer.OfferId))

	if ht.Assert.Equal(200, w.Code) {
		var result aurora.Offer
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)
		ht.Assert.Equal(int64(eurOffer.OfferId), result.ID)
		ht.Assert.Equal("native", result.Selling.Type)
		ht.Assert.Equal("credit_alphanum4", result.Buying.Type)
		ht.Assert.Equal(issuer.Address(), result.Seller)
		ht.Assert.Equal(issuer.Address(), result.Buying.Issuer)
		ht.Assert.Equal(int32(3), result.LastModifiedLedger)

		lastModifiedTime, err := time.Parse("2006-01-02 15:04:05", "2019-06-03 16:34:02")
		ht.Require.NoError(err)
		ht.Assert.Equal(lastModifiedTime, *result.LastModifiedTime)
	}

	w = ht.Get(fmt.Sprintf("/offers/%v", twoEurOffer.OfferId))

	if ht.Assert.Equal(200, w.Code) {
		var result aurora.Offer
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)
		ht.Assert.Equal(int32(20), result.LastModifiedLedger)
		ht.Assert.Nil(result.LastModifiedTime)
	}
}

func TestOfferActions_OfferDoesNotExist(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	ht.App.config.EnableExperimentalIngestion = true
	defer ht.Finish()
	q := &history.Q{ht.AuroraSession()}
	ht.Assert.NoError(q.UpdateLastLedgerExpIngest(3))

	w := ht.Get("/offers/123456")

	ht.Assert.Equal(404, w.Code)
}

func TestOfferActionsStillIngesting_Show(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	ht.App.config.EnableExperimentalIngestion = true
	defer ht.Finish()
	q := &history.Q{ht.AuroraSession()}
	ht.Assert.NoError(q.UpdateLastLedgerExpIngest(0))

	w := ht.Get("/offers/123456")
	ht.Assert.Equal(problem.StillIngesting.Status, w.Code)
}

func TestOfferActions_Index(t *testing.T) {
	ht := StartHTTPTest(t, "trades")
	defer ht.Finish()

	w := ht.Get(
		"/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/offers",
	)

	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(3, w.Body)

		//test last modified timestamp
		var records []map[string]interface{}
		ht.UnmarshalPage(w.Body, &records)

		// Test asset fields population
		ht.Assert.Equal("credit_alphanum4", records[2]["selling"].(map[string]interface{})["asset_type"])
		ht.Assert.Equal("EUR", records[2]["selling"].(map[string]interface{})["asset_code"])
		ht.Assert.Equal("GCQPYGH4K57XBDENKKX55KDTWOTK5WDWRQOH2LHEDX3EKVIQRLMESGBG", records[2]["selling"].(map[string]interface{})["asset_issuer"])

		ht.Assert.Equal("credit_alphanum4", records[2]["buying"].(map[string]interface{})["asset_type"])
		ht.Assert.Equal("USD", records[2]["buying"].(map[string]interface{})["asset_code"])
		ht.Assert.Equal("GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4", records[2]["buying"].(map[string]interface{})["asset_issuer"])

		t2018, err := time.Parse("2006-01-02", "2018-01-01")
		ht.Assert.NoError(err)
		recordTime, err := time.Parse("2006-01-02T15:04:05Z", records[2]["last_modified_time"].(string))
		ht.Assert.True(recordTime.After(t2018))
		ht.Assert.EqualValues(8, records[2]["last_modified_ledger"])
	}
}

func TestOfferActions_IndexNoLedgerData(t *testing.T) {
	ht := StartHTTPTest(t, "trades")
	defer ht.Finish()

	// Remove ledger data
	_, err := ht.App.HistoryQ().ExecRaw("DELETE FROM history_ledgers WHERE sequence=?", 8)
	ht.Assert.NoError(err)

	w := ht.Get(
		"/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/offers",
	)

	// Since 0.15.0 Aurora returns empty data instead of 500 error
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(3, w.Body)

		//test last modified timestamp
		var records []map[string]interface{}
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.NotEmpty(records[2]["last_modified_ledger"])
		ht.Assert.Nil(records[2]["last_modified_time"])
	}
}

func TestOfferActions_SSE(t *testing.T) {
	tt := test.Start(t).Scenario("trades")
	defer tt.Finish()

	ctx := context.Background()
	stream := sse.NewStream(ctx, httptest.NewRecorder())
	oa := OffersByAccountAction{Action: *NewTestAction(ctx, "/foo/bar?account_id=GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2")}

	oa.SSE(stream)
	tt.Require.NoError(oa.Err)

	_, err := tt.CoreSession().ExecRaw(
		`DELETE FROM offers WHERE sellerid = ?`,
		"GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2",
	)
	tt.Require.NoError(err)

	oa.SSE(stream)
	tt.Require.NoError(oa.Err)
}
