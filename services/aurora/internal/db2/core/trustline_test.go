package core

import (
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/diamnet/go/xdr"
)

func assetsToSymbols(assets []xdr.Asset) []string {
	symbols := make([]string, len(assets))
	for i, asset := range assets {
		symbols[i] = asset.String()
	}
	return symbols
}

func TestAssetsForAddress(t *testing.T) {
	tt := test.Start(t).Scenario("order_books")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	assets, balances, err := q.AssetsForAddress(
		"GBXGQJWVLWOYHFLVTKWV5FGHA3LNYY2JQKM7OAJAUEQFU6LPCSEFVXON",
	)
	tt.Assert.NoError(err)
	assetSymbols := assetsToSymbols(assets)

	assetsToBalance := map[string]xdr.Int64{}
	for i, symbol := range assetSymbols {
		assetsToBalance[symbol] = balances[i]
	}

	expected := map[string]xdr.Int64{
		"credit_alphanum4/BTC/GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4": 60000000000,
		"credit_alphanum4/USD/GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4": 50000000000,
		"native": 99999999200,
	}

	tt.Assert.Equal(expected, assetsToBalance)
}

func TestAssetsForAddressWithoutAccount(t *testing.T) {
	tt := test.Start(t).Scenario("order_books")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	var account Account
	err := q.AccountByAddress(&account, "GD5PM5X7Q5MM54ERO2P5PXW3HD6HVZI5IRZGEDWS4OPFBGHNTF6XOWQO")
	tt.Assert.True(q.NoRows(err))

	assets, balances, err := q.AssetsForAddress(
		"GD5PM5X7Q5MM54ERO2P5PXW3HD6HVZI5IRZGEDWS4OPFBGHNTF6XOWQO",
	)
	tt.Assert.NoError(err)
	tt.Assert.Empty(assets)
	tt.Assert.Empty(balances)
}
