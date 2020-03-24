package resourceadapter

import (
	"context"
	"testing"

	protocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/db2/assets"
	"github.com/stretchr/testify/assert"
)

func TestLargeAmount(t *testing.T) {
	row := assets.AssetStatsR{
		SortKey:     "",
		Type:        "credit_alphanum4",
		Code:        "XIM",
		Issuer:      "GBZ35ZJRIKJGYH5PBKLKOZ5L6EXCNTO7BKIL7DAVVDFQ2ODJEEHHJXIM",
		Amount:      "100000000000000000000", // 10T
		NumAccounts: 429,
		Flags:       0,
		Toml:        "https://xim.com/.well-known/hcnet.toml",
	}
	var res protocol.AssetStat
	err := PopulateAssetStat(context.Background(), &res, row)
	assert.NoError(t, err)

	assert.Equal(t, "credit_alphanum4", res.Type)
	assert.Equal(t, "XIM", res.Code)
	assert.Equal(t, "GBZ35ZJRIKJGYH5PBKLKOZ5L6EXCNTO7BKIL7DAVVDFQ2ODJEEHHJXIM", res.Issuer)
	assert.Equal(t, "10000000000000.0000000", res.Amount)
	assert.Equal(t, int32(429), res.NumAccounts)
	assert.Equal(t, "https://xim.com/.well-known/hcnet.toml", res.Links.Toml.Href)
}
