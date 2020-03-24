package ethereum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionAmount(t *testing.T) {
	tests := []struct {
		amount                string
		expectedHcNetAmount string
	}{
		{"1", "0.0000000"},
		{"1234567890123345678", "1.2345679"},
		{"1000000000000000000", "1.0000000"},
		{"150000000000000000000000000", "150000000.0000000"},
	}

	for _, test := range tests {
		bigAmount, ok := new(big.Int).SetString(test.amount, 10)
		assert.True(t, ok)
		transaction := Transaction{ValueWei: bigAmount}
		amount := transaction.ValueToHcNet()
		assert.Equal(t, test.expectedHcNetAmount, amount)
	}
}
