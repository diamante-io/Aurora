package simplepath

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/paths"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func TestFinder(t *testing.T) {
	tt := test.Start(t).Scenario("paths")
	defer tt.Finish()

	finder := &Finder{
		Q: &core.Q{Session: tt.CoreSession()},
	}

	native := makeAsset(xdr.AssetTypeAssetTypeNative, "", "")
	usd := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"USD",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	eur := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"EUR",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	inter1 := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"1",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	inter21 := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"21",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	inter22 := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"22",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")

	query := paths.Query{
		DestinationAsset:  eur,
		DestinationAmount: xdr.Int64(200000000), // 20.0000000
		SourceAssets:      []xdr.Asset{usd},
	}

	p, err := finder.Find(query, MaxPathLength)
	if tt.Assert.NoError(err) {
		tt.Assert.Len(p, 3)

		// Consuming offers:
		// - selling 10 USD for EUR, price = 0.5
		// - selling 10 USD for EUR, price = 0.5
		tt.Assert.Equal(p[0].Source.String(), usd.String())
		tt.Assert.Equal(p[0].Destination.String(), eur.String())
		tt.Assert.Equal(p[0].SourceAmount, xdr.Int64(100000000)) // 10.0000000
		tt.Assert.Equal(p[0].DestinationAmount, xdr.Int64(200000000))
		tt.Assert.Len(p[0].Path, 0)

		// Consuming offers:
		// - selling 20 USD for `1`, price = 1
		// - selling 20 `1` for EUR, price = 1
		tt.Assert.Equal(p[1].Source.String(), usd.String())
		tt.Assert.Equal(p[1].Destination.String(), eur.String())
		tt.Assert.Equal(p[1].SourceAmount, xdr.Int64(200000000))
		tt.Assert.Equal(p[1].DestinationAmount, xdr.Int64(200000000))
		if tt.Assert.Len(p[1].Path, 1) {
			tt.Assert.Equal(p[1].Path[0].String(), inter1.String())
		}

		// Consuming offers:
		// - selling 20 USD for `21`, price = 1
		// - selling 20 `21` for `22`, price = 1
		// - selling 20 `22` for EUR, price = 1
		tt.Assert.Equal(p[2].Source.String(), usd.String())
		tt.Assert.Equal(p[2].Destination.String(), eur.String())
		tt.Assert.Equal(p[2].SourceAmount, xdr.Int64(200000000))
		tt.Assert.Equal(p[2].DestinationAmount, xdr.Int64(200000000))
		if tt.Assert.Len(p[2].Path, 2) {
			tt.Assert.Equal(p[2].Path[0].String(), inter21.String())
			tt.Assert.Equal(p[2].Path[1].String(), inter22.String())
		}
	}

	query.DestinationAmount = xdr.Int64(200000001)
	p, err = finder.Find(query, MaxPathLength)
	if tt.Assert.NoError(err) {
		tt.Assert.Len(p, 2)

		tt.Assert.Equal(p[0].Source.String(), usd.String())
		tt.Assert.Equal(p[0].Destination.String(), eur.String())
		tt.Assert.Equal(p[0].SourceAmount, xdr.Int64(100000001))
		tt.Assert.Equal(p[0].DestinationAmount, xdr.Int64(200000001))
		tt.Assert.Len(p[0].Path, 0)

		tt.Assert.Equal(p[1].Source.String(), usd.String())
		tt.Assert.Equal(p[1].Destination.String(), eur.String())
		tt.Assert.Equal(p[1].SourceAmount, xdr.Int64(200000001))
		tt.Assert.Equal(p[1].DestinationAmount, xdr.Int64(200000001))
		if tt.Assert.Len(p[1].Path, 2) {
			tt.Assert.Equal(p[1].Path[0].String(), inter21.String())
			tt.Assert.Equal(p[1].Path[1].String(), inter22.String())
		}
	}

	query.DestinationAmount = xdr.Int64(500000001)
	p, err = finder.Find(query, MaxPathLength)
	if tt.Assert.NoError(err) {
		tt.Assert.Len(p, 0)
	}

	//  regression: paths that involve native currencies can be found

	query = paths.Query{
		DestinationAsset:  native,
		DestinationAmount: xdr.Int64(1),
		SourceAssets:      []xdr.Asset{usd, native},
	}
	p, err = finder.Find(query, MaxPathLength)
	if tt.Assert.NoError(err) {
		tt.Assert.Len(p, 2)
	}

	// In the past the order of trades was reversed, ex. the first trade was to sell
	// destination amount, selling source, buying second asset on the path. The algorithm
	// like this is wrong.
	//
	// Consider the following path: AAA -> BBB -> CCC, destination amount = 10 and the
	// following offers:
	//
	// offer :trader, {for:["AAA", :gateway], sell:["BBB", :gateway]}, 1, 11
	// offer :trader, {for:["BBB", :gateway], sell:["CCC", :gateway]}, 10, 0.1
	//
	// For such order books the old algorithm would not find a path as it is not
	// possible to buy 1 BBB with 10 AAA and 1 BBB is needed to buy 10 CCC.
	aaa := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"AAA",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	bbb := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"BBB",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")
	ccc := makeAsset(
		xdr.AssetTypeAssetTypeCreditAlphanum4,
		"CCC",
		"GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN")

	query = paths.Query{
		DestinationAsset:  ccc,
		DestinationAmount: xdr.Int64(100000000), // 10.0
		SourceAssets:      []xdr.Asset{aaa},
	}
	p, err = finder.Find(query, MaxPathLength)
	if tt.Assert.NoError(err) {
		if tt.Assert.Len(p, 1) {
			tt.Assert.Equal(p[0].Source.String(), aaa.String())
			tt.Assert.Equal(p[0].Destination.String(), ccc.String())
			tt.Assert.Equal(p[0].SourceAmount, xdr.Int64(110000000))      // 11.0
			tt.Assert.Equal(p[0].DestinationAmount, xdr.Int64(100000000)) // 10.0
			if tt.Assert.Len(p[0].Path, 1) {
				tt.Assert.Equal(p[0].Path[0].String(), bbb.String())
			}
		}
	}
}
