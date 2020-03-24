package codes

import (
	"testing"

	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		Input    interface{}
		Expected string
		Err      error
	}{
		{xdr.TransactionResultCodeTxSuccess, "tx_success", nil},
		{xdr.OperationResultCodeOpBadAuth, "op_bad_auth", nil},
		{xdr.CreateAccountResultCodeCreateAccountLowReserve, "op_low_reserve", nil},
		{xdr.PaymentResultCodePaymentSrcNoTrust, "op_src_no_trust", nil},
		{0, "", ErrUnknownCode},
	}

	for _, test := range tests {
		actual, err := String(test.Input)

		if test.Err != nil {
			assert.NotNil(t, err)
			assert.Equal(t, test.Err.Error(), err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, test.Expected, actual)
		}
	}
}

//TODO: op_inner refers to inner result code
//TODO: non op_inner uses the outer result code
//TODO: one test for each operation type
