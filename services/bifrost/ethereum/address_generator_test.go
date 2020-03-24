package ethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressGenerator(t *testing.T) {
	// Generated using https://iancoleman.github.io/bip39/
	// Root key:
	// xprv9s21ZrQH143K2Cfj4mDZBcEecBmJmawReGwwoAou2zZzG45bM6cFPJSvobVTCB55L6Ld2y8RzC61CpvadeAnhws3CHsMFhNjozBKGNgucYm
	// Derivation Path m/44'/60'/0'/0:
	// xprv9zy5o7z1GMmYdaeQdmabWFhUf52Ytbpe3G5hduA4SghboqWe7aDGWseN8BJy1GU72wPjkCbBE1hvbXYqpCecAYdaivxjNnBoSNxwYD4wHpW
	// xpub6DxSCdWu6jKqr4isjo7bsPeDD6s3J4YVQV1JSHZg12Eagdqnf7XX4fxqyW2sLhUoFWutL7tAELU2LiGZrEXtjVbvYptvTX5Eoa4Mamdjm9u
	generator, err := NewAddressGenerator("xpub6DxSCdWu6jKqr4isjo7bsPeDD6s3J4YVQV1JSHZg12Eagdqnf7XX4fxqyW2sLhUoFWutL7tAELU2LiGZrEXtjVbvYptvTX5Eoa4Mamdjm9u")
	assert.NoError(t, err)

	expectedChildren := []struct {
		index   uint32
		address string
	}{
		{0, "0x044d22459b0Ce2eBa60B47ee411F8B6a8f91dF52"},
		{1, "0xc881d34F83001A0c96C422594ea9EBE0c0114973"},
		{2, "0x61203C142Fe744499D819ca5d36753F4461e174D"},
		{3, "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D"},
		{4, "0x314d1281f7cf78E5EC28DB62194Ef80a91f13b61"},
		{5, "0xD63eC9c99459BB2D9688CC71Eb849fDA142d55C5"},
		{6, "0xd977D20405c549a36A50Be06AE3B754155Fb3dDa"},
		{7, "0xC0dbAe13052CD4F4B9B674496a72Fc02d05aF442"},
		{8, "0x9A44d3447821885Ea60eb708c0EB0e50493Add0F"},
		{9, "0x82ae892Dfe0bED4c4b83780a00F4723D71c19b1D"},

		{100, "0x7aB27448C69aD3e10A754899151d006285DD0f60"},
		{101, "0xBCdcD65F4Db02CBc99FfbC3Bc045e4BC180f302f"},
		{102, "0x3A0A72A9644DCE86Df1C16Fdeb665a574009d9c4"},
		{103, "0xA21666e6BDF9F58EB422098fbe0850e440E9f53C"},
		{104, "0x208cDC0DF05321E4d3Ae2b6473cdf0c928e2fCd0"},
		{105, "0x30C70532a99f7Ea1186a476A22601966945a2624"},
		{106, "0xc345cA1D1347B6A7a532AFd19ccf30228e438D67"},
		{107, "0x35F58b0A6104E78998ca295C17F37Fc545Ed1E1c"},
		{108, "0x0079022194FF43188c3f3E571c503C15bAb4E3F3"},
		{109, "0x0f7203D9Aa1395D37CCee478E4E24e2bfDe54879"},

		{1000, "0x02D2DEbeA9A27F964aBEcE49a5d64062637Bd6C5"},
	}

	for _, child := range expectedChildren {
		address, err := generator.Generate(child.index)
		assert.NoError(t, err)
		assert.Equal(t, child.address, address)
	}
}
