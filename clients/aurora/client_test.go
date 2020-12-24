package aurora

import (
	"net/http"
	"testing"

	"github.com/diamnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	h := httptest.NewClient()
	auroraClient := &Client{
		URL:  "https://aurora.diamnet.org",
		HTTP: h,
	}

	// happy path
	h.On(
		"GET",
		"https://aurora.diamnet.org/trades/?base_asset_type=native&counter_asset_code=SLT&counter_asset_issuer=GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP&counter_asset_type=credit_alphanum4&limit=3&offer_id=0&order=asc&resolution=300000",
	).ReturnString(http.StatusOK, tradesNormalResponse)

	trades, err := auroraClient.LoadTrades(
		Asset{Type: "native"},
		Asset{"credit_alphanum4", "SLT", "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP"},
		0,
		300000,
		Limit(3),
		Order(OrderAsc),
	)

	require.NoError(t, err)
	assert.Equal(t, "61557992432078849-0", trades.Embedded.Records[0].ID)
	assert.Equal(t, "187430", trades.Embedded.Records[0].OfferID)
	assert.Equal(t, "GAKLCFRTFDXKOEEUSBS23FBSUUVJRMDQHGCHNGGGJZQRK7BCPIMHUC4P", trades.Embedded.Records[0].BaseAccount)
	assert.Equal(t, "0.1558000", trades.Embedded.Records[0].BaseAmount)
	assert.Equal(t, "native", trades.Embedded.Records[0].BaseAssetType)
	assert.Equal(t, "GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P", trades.Embedded.Records[0].CounterAccount)
	assert.Equal(t, "0.0100000", trades.Embedded.Records[0].CounterAmount)
	assert.Equal(t, "credit_alphanum4", trades.Embedded.Records[0].CounterAssetType)
	assert.Equal(t, "SLT", trades.Embedded.Records[0].CounterAssetCode)
	assert.Equal(t, "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP", trades.Embedded.Records[0].CounterAssetIssuer)

	// error case: wrong resolution
	trades, err = auroraClient.LoadTrades(
		Asset{Type: "native"},
		Asset{"credit_alphanum4", "SLT", "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP"},
		0,
		1234567,
	)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to load endpoint")
	}
}

var tradesNormalResponse = `{
"_links": {
"self": {
"href": "https://aurora.diamnet.org/trades?base_asset_code=&base_asset_issuer=&base_asset_type=native&counter_asset_code=SLT&counter_asset_issuer=GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP&counter_asset_type=credit_alphanum4&cursor=&limit=3&offer_id=0&order=asc&resolution=300000"
},
"next": {
"href": "https://aurora.diamnet.org/trades?base_asset_code=&base_asset_issuer=&base_asset_type=native&counter_asset_code=SLT&counter_asset_issuer=GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP&counter_asset_type=credit_alphanum4&cursor=61560277354680321-0&limit=3&offer_id=0&order=asc&resolution=300000"
},
"prev": {
"href": "https://aurora.diamnet.org/trades?base_asset_code=&base_asset_issuer=&base_asset_type=native&counter_asset_code=SLT&counter_asset_issuer=GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP&counter_asset_type=credit_alphanum4&cursor=61557992432078849-0&limit=3&offer_id=0&order=desc&resolution=300000"
}
},
"_embedded": {
"records": [
{
"_links": {
"self": {
"href": ""
},
"base": {
"href": "https://aurora.diamnet.org/accounts/GAKLCFRTFDXKOEEUSBS23FBSUUVJRMDQHGCHNGGGJZQRK7BCPIMHUC4P"
},
"counter": {
"href": "https://aurora.diamnet.org/accounts/GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P"
},
"operation": {
"href": "https://aurora.diamnet.org/operations/61557992432078849"
}
},
"id": "61557992432078849-0",
"paging_token": "61557992432078849-0",
"ledger_close_time": "2017-11-02T13:19:18Z",
"offer_id": "187430",
"base_account": "GAKLCFRTFDXKOEEUSBS23FBSUUVJRMDQHGCHNGGGJZQRK7BCPIMHUC4P",
"base_amount": "0.1558000",
"base_asset_type": "native",
"counter_account": "GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P",
"counter_amount": "0.0100000",
"counter_asset_type": "credit_alphanum4",
"counter_asset_code": "SLT",
"counter_asset_issuer": "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP",
"base_is_seller": false,
"price": {
"n": 50,
"d": 779
}
},
{
"_links": {
"self": {
"href": ""
},
"base": {
"href": "https://aurora.diamnet.org/accounts/GDHFBKVJO3VVQHPR56FV4YOILEGX3PYZBFNHVRJCDZTDULFPPWZHNNVL"
},
"counter": {
"href": "https://aurora.diamnet.org/accounts/GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P"
},
"operation": {
"href": "https://aurora.diamnet.org/operations/61560234405007361"
}
},
"id": "61560234405007361-0",
"paging_token": "61560234405007361-0",
"ledger_close_time": "2017-11-02T14:03:12Z",
"offer_id": "187430",
"base_account": "GDHFBKVJO3VVQHPR56FV4YOILEGX3PYZBFNHVRJCDZTDULFPPWZHNNVL",
"base_amount": "7.7900000",
"base_asset_type": "native",
"counter_account": "GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P",
"counter_amount": "0.5000000",
"counter_asset_type": "credit_alphanum4",
"counter_asset_code": "SLT",
"counter_asset_issuer": "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP",
"base_is_seller": false,
"price": {
"n": 50,
"d": 779
}
},
{
"_links": {
"self": {
"href": ""
},
"base": {
"href": "https://aurora.diamnet.org/accounts/GDLHSCWRFUNEEJL6PR67OZL7QVO2L57MKQOMS6LGKNLGZPX6KCHXREMP"
},
"counter": {
"href": "https://aurora.diamnet.org/accounts/GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P"
},
"operation": {
"href": "https://aurora.diamnet.org/operations/61560277354680321"
}
},
"id": "61560277354680321-0",
"paging_token": "61560277354680321-0",
"ledger_close_time": "2017-11-02T14:04:02Z",
"offer_id": "187430",
"base_account": "GDLHSCWRFUNEEJL6PR67OZL7QVO2L57MKQOMS6LGKNLGZPX6KCHXREMP",
"base_amount": "155.8000000",
"base_asset_type": "native",
"counter_account": "GAYDG77BFUUHYXC4IMFGXNBFDS5TMBB545Q6MON3EXYXHDOEJWU2LD2P",
"counter_amount": "10.0000000",
"counter_asset_type": "credit_alphanum4",
"counter_asset_code": "SLT",
"counter_asset_issuer": "GCKA6K5PCQ6PNF5RQBF7PQDJWRHO6UOGFMRLK3DYHDOI244V47XKQ4GP",
"base_is_seller": false,
"price": {
"n": 50,
"d": 779
}
}
]
}
}`
