package aurora

import (
	"testing"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/protocols/aurora/base"
	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/diamnet/go/support/render/hal"
)

func TestAssetsActions(t *testing.T) {
	testDomain := struct {
		Toml hal.Link `json:"toml"`
	}{
		Toml: hal.NewLink("https://test.com/.well-known/diamnet.toml"),
	}
	empty := struct {
		Toml hal.Link `json:"toml"`
	}{
		Toml: hal.NewLink(""),
	}

	BTCGateway := aurora.AssetStat{
		Links: testDomain,
		Asset: base.Asset{
			Type:   "credit_alphanum4",
			Code:   "BTC",
			Issuer: "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
		},
		PT:          "BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
		Amount:      "100.9876000",
		NumAccounts: 1,
		Flags: aurora.AccountFlags{
			AuthRequired:  true,
			AuthRevocable: false,
		},
	}
	SCOTScott := aurora.AssetStat{
		Links: empty,
		Asset: base.Asset{
			Type:   "credit_alphanum4",
			Code:   "SCOT",
			Issuer: "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
		},
		PT:          "SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4",
		Amount:      "1000.0000000",
		NumAccounts: 1,
		Flags: aurora.AccountFlags{
			AuthRequired:  false,
			AuthRevocable: true,
		},
	}
	USDGateway := aurora.AssetStat{
		Links: testDomain,
		Asset: base.Asset{
			Type:   "credit_alphanum4",
			Code:   "USD",
			Issuer: "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
		},
		PT:          "USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
		Amount:      "300001.0434000",
		NumAccounts: 2,
		Flags: aurora.AccountFlags{
			AuthRequired:  true,
			AuthRevocable: false,
		},
	}

	testCases := []struct {
		path         string
		wantSelf     string
		wantPrevious string
		wantNext     string
		wantItems    []aurora.AssetStat
	}{
		{
			"/assets",
			"/assets?order=asc&limit=10&cursor=",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{BTCGateway, SCOTScott, USDGateway},
		},
		// limit
		{
			"/assets?limit=3",
			"/assets?order=asc&limit=3&cursor=",
			"/assets?order=desc&limit=3&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=3&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{BTCGateway, SCOTScott, USDGateway},
		}, {
			"/assets?limit=1",
			"/assets?order=asc&limit=1&cursor=",
			"/assets?order=desc&limit=1&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=1&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{BTCGateway},
		},
		// cursor
		{
			"/assets?cursor=0",
			"/assets?order=asc&limit=10&cursor=0",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{BTCGateway, SCOTScott, USDGateway},
		}, {
			"/assets?cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{SCOTScott, USDGateway},
		}, {
			"/assets?cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{USDGateway},
		}, {
			"/assets?cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4", // TODO NNS 2 - I think this should be cursor=current+1 but it returns cursor=3, is that a bug?
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{},
		},
		// order
		{
			"/assets?order=asc",
			"/assets?order=asc&limit=10&cursor=",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{BTCGateway, SCOTScott, USDGateway},
		}, {
			"/assets?order=desc",
			"/assets?order=desc&limit=10&cursor=",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{USDGateway, SCOTScott, BTCGateway},
		}, {
			"/assets?order=desc&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{SCOTScott, BTCGateway},
		}, {
			"/assets?order=desc&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4", // TODO NNS 2 - I think this should be cursor="/"
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			[]aurora.AssetStat{},
		},
		// asset_code
		{
			"/assets?asset_code=noexist",
			"/assets?order=asc&limit=10&cursor=&asset_code=noexist",
			"/assets?order=desc&limit=10&cursor=&asset_code=noexist", // TODO NNS 2 - imo, should be cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4
			"/assets?order=asc&limit=10&cursor=&asset_code=noexist",
			[]aurora.AssetStat{},
		}, {
			"/assets?asset_code=USD",
			"/assets?order=asc&limit=10&cursor=&asset_code=USD",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD",
			[]aurora.AssetStat{USDGateway},
		}, {
			"/assets?asset_code=BTC",
			"/assets?order=asc&limit=10&cursor=&asset_code=BTC",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC",
			"/assets?order=asc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC",
			[]aurora.AssetStat{BTCGateway},
		}, {
			"/assets?asset_code=SCOT",
			"/assets?order=asc&limit=10&cursor=&asset_code=SCOT",
			"/assets?order=desc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4&asset_code=SCOT",
			"/assets?order=asc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4&asset_code=SCOT",
			[]aurora.AssetStat{SCOTScott},
		},
		// asset_issuer
		{
			"/assets?asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{BTCGateway, USDGateway},
		}, {
			"/assets?asset_issuer=GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			"/assets?order=asc&limit=10&cursor=&asset_issuer=GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			"/assets?order=desc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4&asset_issuer=GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			"/assets?order=asc&limit=10&cursor=SCOT_GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU_credit_alphanum4&asset_issuer=GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			[]aurora.AssetStat{SCOTScott},
		}, {
			"/assets?asset_issuer=GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2",
			"/assets?order=asc&limit=10&cursor=&asset_issuer=GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2",
			"/assets?order=desc&limit=10&cursor=&asset_issuer=GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", // TODO NNS 2 - imo should be cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4
			"/assets?order=asc&limit=10&cursor=&asset_issuer=GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2",
			[]aurora.AssetStat{},
		},
		// combined
		{
			"/assets?asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4&cursor=-",
			"/assets?order=asc&limit=10&cursor=-&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{USDGateway},
		}, {
			"/assets?asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=USD&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{USDGateway},
		}, {
			"/assets?asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4",
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4", // TODO NNS 2 - imo, should be cursor=cursor+1 or lastCursor
			"/assets?order=asc&limit=10&cursor=USD_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{},
		}, {
			"/assets?asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=10&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{BTCGateway},
		}, {
			"/assets?asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4&limit=1&order=desc",
			"/assets?order=desc&limit=1&cursor=&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=asc&limit=1&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			"/assets?order=desc&limit=1&cursor=BTC_GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4_credit_alphanum4&asset_code=BTC&asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			[]aurora.AssetStat{BTCGateway},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.path, func(t *testing.T) {
			ht := StartHTTPTest(t, "ingest_asset_stats")
			defer ht.Finish()

			// Ugly but saves us time needed to change each `StartHTTPTest` occurence.
			appConfig := NewTestConfig()
			appConfig.EnableAssetStats = true

			ht.App = NewApp(appConfig)
			ht.RH = test.NewRequestHelper(ht.App.web.router)

			w := ht.Get(kase.path)
			ht.Assert.Equal(200, w.Code)
			ht.Assert.PageOf(len(kase.wantItems), w.Body)

			records := []aurora.AssetStat{}
			links := ht.UnmarshalPage(w.Body, &records)
			if ht.Assert.Equal(len(kase.wantItems), len(records)) {
				for i := range kase.wantItems {
					ht.Assert.Equal(kase.wantItems[i], records[i])
				}
			}

			ht.Assert.EqualUrlStrings("http://localhost"+kase.wantSelf, links.Self.Href)
			ht.Assert.EqualUrlStrings("http://localhost"+kase.wantPrevious, links.Prev.Href)
			ht.Assert.EqualUrlStrings("http://localhost"+kase.wantNext, links.Next.Href)
		})
	}
}

func TestInvalidAssetCode(t *testing.T) {
	ht := StartHTTPTest(t, "ingest_asset_stats")
	defer ht.Finish()

	// Ugly but saves us time needed to change each `StartHTTPTest` occurence.
	appConfig := NewTestConfig()
	appConfig.EnableAssetStats = true

	ht.App = NewApp(appConfig)
	ht.RH = test.NewRequestHelper(ht.App.web.router)

	w := ht.Get("/assets?asset_code=ABCDEFGHIJKL")
	ht.Assert.Equal(200, w.Code)

	w = ht.Get("/assets?asset_code=ABCDEFGHIJKLM")
	ht.Assert.Equal(400, w.Code)
}

func TestInvalidAssetIssuer(t *testing.T) {
	ht := StartHTTPTest(t, "ingest_asset_stats")
	defer ht.Finish()

	// Ugly but saves us time needed to change each `StartHTTPTest` occurence.
	appConfig := NewTestConfig()
	appConfig.EnableAssetStats = true

	ht.App = NewApp(appConfig)
	ht.RH = test.NewRequestHelper(ht.App.web.router)

	w := ht.Get("/assets?asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	ht.Assert.Equal(200, w.Code)

	w = ht.Get("/assets?asset_issuer=invalid")
	ht.Assert.Equal(400, w.Code)
}

func TestAssetStatsDisabledByDefault(t *testing.T) {
	ht := StartHTTPTest(t, "ingest_asset_stats")
	defer ht.Finish()

	w := ht.Get("/assets?asset_issuer=GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
	ht.Assert.Equal(404, w.Code)
}
