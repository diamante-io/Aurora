package bitcoin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionAmount(t *testing.T) {
	tests := []struct {
		amount                int64
		expectedDiamNetAmount string
	}{
		{1, "0.0000000"},
		{4, "0.0000000"},
		{5, "0.0000001"},
		{10, "0.0000001"},
		{12345674, "0.1234567"},
		{12345678, "0.1234568"},
		{100000000, "1.0000000"},
		{2100000000000000, "21000000.0000000"},
	}

	for _, test := range tests {
		transaction := Transaction{ValueSat: test.amount}
		amount := transaction.ValueToDiamNet()
		assert.Equal(t, test.expectedDiamNetAmount, amount)
	}
}
