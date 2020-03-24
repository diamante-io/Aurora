package expingest

import (
	"sort"
	"testing"

	"github.com/hcnet/go/exp/orderbook"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

var (
	issuer   = xdr.MustAddress("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
	usdAsset = xdr.Asset{
		Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
		AlphaNum4: &xdr.AssetAlphaNum4{
			AssetCode: [4]byte{'u', 's', 'd', 0},
			Issuer:    issuer,
		},
	}

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

func TestLoadOrderBookGraphFromEmptyDB(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	q := &history.Q{tt.AuroraSession()}
	graph := orderbook.NewOrderBookGraph()

	err := loadOrderBookGraphFromDB(q, graph)
	tt.Assert.NoError(err)
	tt.Assert.True(graph.IsEmpty())
}

func TestLoadOrderBookGraph(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	q := &history.Q{tt.AuroraSession()}
	graph := orderbook.NewOrderBookGraph()

	tt.Assert.NoError(q.UpdateLastLedgerExpIngest(123))
	tt.Assert.NoError(q.UpsertOffer(eurOffer, 123))
	tt.Assert.NoError(q.UpsertOffer(twoEurOffer, 123))

	err := loadOrderBookGraphFromDB(q, graph)
	tt.Assert.NoError(err)
	tt.Assert.False(graph.IsEmpty())

	offers := graph.Offers()
	sort.Slice(offers, func(i, j int) bool {
		return offers[i].OfferId < offers[j].OfferId
	})
	expectedOffers := []xdr.OfferEntry{
		eurOffer, twoEurOffer,
	}
	tt.Assert.Equal(expectedOffers, offers)
}
