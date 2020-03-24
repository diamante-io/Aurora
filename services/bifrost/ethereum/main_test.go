package ethereum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEthToWei(t *testing.T) {
	hugeEth, ok := new(big.Int).SetString("100000000123456789012345678", 10)
	assert.True(t, ok)

	tests := []struct {
		amount         string
		expectedAmount *big.Int
		expectedError  string
	}{
		{"", nil, "Could not convert to *big.Rat"},
		{"test", nil, "Could not convert to *big.Rat"},
		{"0.0000000000000000001", nil, "Invalid precision"},
		{"1.2345678912345678901", nil, "Invalid precision"},

		{"0", big.NewInt(0), ""},
		{"0.00", big.NewInt(0), ""},
		{"0.000000000000000001", big.NewInt(1), ""},
		{"1", big.NewInt(1000000000000000000), ""},
		{"1.00", big.NewInt(1000000000000000000), ""},
		{"1.234567890123456789", big.NewInt(1234567890123456789), ""},
		{"100000000.123456789012345678", hugeEth, ""},
	}

	for _, test := range tests {
		returnedAmount, err := EthToWei(test.amount)
		if test.expectedError != "" {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, 0, returnedAmount.Cmp(test.expectedAmount))
		}
	}
}
