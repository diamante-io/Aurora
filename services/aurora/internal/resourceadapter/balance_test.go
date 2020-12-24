package resourceadapter

import (
	"testing"

	. "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestPopulateBalance(t *testing.T) {
	testAssetCode1 := "TEST_ASSET_1"
	testAssetCode2 := "TEST_ASSET_2"
	authorizedTrustline := core.Trustline{
		Accountid: "testID",
		Assettype: xdr.AssetTypeAssetTypeCreditAlphanum12,
		Issuer:    "",
		Assetcode: testAssetCode1,
		Tlimit:    100,
		Balance:   10,
		Flags:     1,
	}
	unauthorizedTrustline := core.Trustline{
		Accountid: "testID",
		Assettype: xdr.AssetTypeAssetTypeCreditAlphanum12,
		Issuer:    "",
		Assetcode: testAssetCode2,
		Tlimit:    100,
		Balance:   10,
		Flags:     2,
	}

	want := Balance{}
	err := PopulateBalance(&want, authorizedTrustline)
	assert.NoError(t, err)
	assert.Equal(t, "credit_alphanum12", want.Type)
	assert.Equal(t, "0.0000010", want.Balance)
	assert.Equal(t, "0.0000100", want.Limit)
	assert.Equal(t, "", want.Issuer)
	assert.Equal(t, testAssetCode1, want.Code)
	assert.Equal(t, true, *want.IsAuthorized)

	want = Balance{}
	err = PopulateBalance(&want, unauthorizedTrustline)
	assert.NoError(t, err)
	assert.Equal(t, "credit_alphanum12", want.Type)
	assert.Equal(t, "0.0000010", want.Balance)
	assert.Equal(t, "0.0000100", want.Limit)
	assert.Equal(t, "", want.Issuer)
	assert.Equal(t, testAssetCode2, want.Code)
	assert.Equal(t, false, *want.IsAuthorized)
}

func TestPopulateNativeBalance(t *testing.T) {
	want := Balance{}
	err := PopulateNativeBalance(&want, 10, 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, "native", want.Type)
	assert.Equal(t, "0.0000010", want.Balance)
	assert.Equal(t, "0.0000010", want.BuyingLiabilities)
	assert.Equal(t, "0.0000010", want.SellingLiabilities)
	assert.Equal(t, "", want.Limit)
	assert.Equal(t, "", want.Issuer)
	assert.Equal(t, "", want.Code)
	assert.Nil(t, want.IsAuthorized)
}
