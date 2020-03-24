package aurora

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/hcnet/go/exp/orderbook"
	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/render/problem"
	"github.com/hcnet/go/services/aurora/internal/simplepath"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func TestPathActions_Index(t *testing.T) {
	ht := StartHTTPTest(t, "paths")
	defer ht.Finish()

	// no query args
	w := ht.Get("/paths")
	ht.Assert.Equal(400, w.Code)

	// happy path
	var q = make(url.Values)

	q.Add(
		"destination_account",
		"GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V",
	)
	q.Add(
		"source_account",
		"GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP",
	)
	q.Add(
		"destination_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add("destination_asset_type", "credit_alphanum4")
	q.Add("destination_asset_code", "EUR")
	q.Add("destination_amount", "10")

	for _, uri := range []string{"/paths", "/paths/strict-receive"} {
		w = ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(200, w.Code)
		ht.Assert.PageOf(3, w.Body)
	}
}

func TestPathActionsStillIngesting(t *testing.T) {
	ht := StartHTTPTest(t, "paths")
	defer ht.Finish()

	orderBookGraph := orderbook.NewOrderBookGraph()
	ht.App.paths = simplepath.NewInMemoryFinder(orderBookGraph)

	var q = make(url.Values)

	q.Add(
		"destination_account",
		"GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V",
	)
	q.Add(
		"source_account",
		"GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP",
	)
	q.Add(
		"destination_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add("destination_asset_type", "credit_alphanum4")
	q.Add("destination_asset_code", "EUR")
	q.Add("destination_amount", "10")

	for _, uri := range []string{"/paths", "/paths/strict-receive"} {
		w := ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(problem.StillIngesting.Status, w.Code)
	}
}

func loadOffers(tt *test.T, orderBookGraph *orderbook.OrderBookGraph, fromAddress string) {
	coreQ := &core.Q{tt.CoreSession()}
	offers := []core.Offer{}
	pageQuery := db2.PageQuery{
		Order: db2.OrderAscending,
		Limit: 100,
	}
	err := coreQ.OffersByAddress(&offers, fromAddress, pageQuery)
	tt.Assert.NoError(err)
	for _, offer := range offers {

		orderBookGraph.AddOffer(xdr.OfferEntry{
			SellerId: xdr.MustAddress(offer.SellerID),
			OfferId:  xdr.Int64(offer.OfferID),
			Selling:  offer.SellingAsset,
			Buying:   offer.BuyingAsset,
			Amount:   offer.Amount,
			Price:    xdr.Price{N: xdr.Int32(offer.Price * 100), D: 100},
		})
	}
	tt.Assert.NoError(orderBookGraph.Apply())
}

func TestPathActionsInMemoryFinder(t *testing.T) {
	ht := StartHTTPTest(t, "paths")
	defer ht.Finish()

	orderBookGraph := orderbook.NewOrderBookGraph()

	loadOffers(ht.T, orderBookGraph, "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL")
	loadOffers(ht.T, orderBookGraph, "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")

	var q = make(url.Values)

	q.Add(
		"destination_account",
		"GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V",
	)
	q.Add(
		"source_account",
		"GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP",
	)
	q.Add(
		"destination_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add("destination_asset_type", "credit_alphanum4")
	q.Add("destination_asset_code", "EUR")
	q.Add("destination_amount", "10")

	for _, uri := range []string{"/paths", "/paths/strict-receive"} {
		ht.App.paths = simplepath.NewInMemoryFinder(orderBookGraph)

		w := ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(http.StatusOK, w.Code)
		inMemoryResponse := []aurora.Path{}
		ht.UnmarshalPage(w.Body, &inMemoryResponse)

		ht.App.paths = &simplepath.Finder{ht.App.CoreQ()}

		w = ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(http.StatusOK, w.Code)
		dbResponse := []aurora.Path{}
		ht.UnmarshalPage(w.Body, &dbResponse)

		ht.Assert.Equal(inMemoryResponse, dbResponse)
	}
}

func TestPathActionsEmptySourceAcount(t *testing.T) {
	ht := StartHTTPTest(t, "paths")
	defer ht.Finish()

	orderBookGraph := orderbook.NewOrderBookGraph()

	loadOffers(ht.T, orderBookGraph, "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL")
	loadOffers(ht.T, orderBookGraph, "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")

	var q = make(url.Values)

	q.Add(
		"destination_account",
		"GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V",
	)
	q.Add(
		"source_account",
		// there is no account associated with this address
		"GD5PM5X7Q5MM54ERO2P5PXW3HD6HVZI5IRZGEDWS4OPFBGHNTF6XOWQO",
	)
	q.Add(
		"destination_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add("destination_asset_type", "credit_alphanum4")
	q.Add("destination_asset_code", "EUR")
	q.Add("destination_amount", "10")

	for _, uri := range []string{"/paths", "/paths/strict-receive"} {
		ht.App.paths = simplepath.NewInMemoryFinder(orderBookGraph)

		w := ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(http.StatusOK, w.Code)
		inMemoryResponse := []aurora.Path{}
		ht.UnmarshalPage(w.Body, &inMemoryResponse)
		ht.Assert.Empty(inMemoryResponse)

		ht.App.paths = &simplepath.Finder{ht.App.CoreQ()}

		w = ht.Get(uri + "?" + q.Encode())
		ht.Assert.Equal(http.StatusOK, w.Code)
		dbResponse := []aurora.Path{}
		ht.UnmarshalPage(w.Body, &dbResponse)
		ht.Assert.Empty(dbResponse)
	}
}

func TestPathActionsStrictSend(t *testing.T) {
	ht := StartHTTPTest(t, "paths")
	defer ht.Finish()

	orderBookGraph := orderbook.NewOrderBookGraph()

	loadOffers(ht.T, orderBookGraph, "GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL")
	loadOffers(ht.T, orderBookGraph, "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")

	ht.App.paths = simplepath.NewInMemoryFinder(orderBookGraph)

	var q = make(url.Values)

	q.Add(
		"source_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add(
		"source_account",
		"GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP",
	)
	q.Add(
		"destination_asset_issuer",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
	)
	q.Add("source_asset_type", "credit_alphanum4")
	q.Add("source_asset_code", "USD")
	q.Add("source_amount", "10")
	q.Add("destination_asset_type", "credit_alphanum4")
	q.Add("destination_asset_code", "EUR")

	w := ht.Get("/paths/strict-send?" + q.Encode())
	ht.Assert.Equal(http.StatusOK, w.Code)
	inMemoryResponse := []aurora.Path{}
	ht.UnmarshalPage(w.Body, &inMemoryResponse)
	ht.Assert.Len(inMemoryResponse, 4)
	for i, path := range inMemoryResponse {
		ht.Assert.Equal(path.SourceAssetCode, "USD")
		ht.Assert.Equal(path.SourceAssetType, "credit_alphanum4")
		ht.Assert.Equal(path.SourceAssetIssuer, "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
		ht.Assert.Equal(path.SourceAmount, "10.0000000")

		ht.Assert.Equal(path.DestinationAssetCode, "EUR")
		ht.Assert.Equal(path.DestinationAssetType, "credit_alphanum4")
		ht.Assert.Equal(path.DestinationAssetIssuer, "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
		if i > 1 {
			previous, err := strconv.ParseFloat(inMemoryResponse[i-1].DestinationAmount, 64)
			ht.Assert.NoError(err)

			current, err := strconv.ParseFloat(path.DestinationAmount, 64)
			ht.Assert.NoError(err)

			ht.Assert.True(previous >= current)
		}
	}
}
