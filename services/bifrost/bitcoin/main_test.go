package bitcoin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtcToSat(t *testing.T) {
	tests := []struct {
		amount         string
		expectedAmount int64
		expectedError  string
	}{
		{"", 0, "Could not convert to *big.Rat"},
		{"test", 0, "Could not convert to *big.Rat"},
		{"0.000000001", 0, "Invalid precision"},
		{"1.234567891", 0, "Invalid precision"},

		{"0", 0, ""},
		{"0.00", 0, ""},
		{"1", 100000000, ""},
		{"1.00", 100000000, ""},
		{"1.23456789", 123456789, ""},
		{"1.234567890", 123456789, ""},
		{"0.00000001", 1, ""},
		{"21000000.12345678", 2100000012345678, ""},
	}

	for _, test := range tests {
		returnedAmount, err := BtcToSat(test.amount)
		if test.expectedError != "" {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.expectedError)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expectedAmount, returnedAmount)
		}
	}
}
