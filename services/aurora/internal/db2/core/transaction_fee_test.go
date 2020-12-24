package core

import (
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
)

func TestTransactionFeesByLedger(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	var fees []TransactionFee
	err := q.TransactionFeesByLedger(&fees, 2)

	if tt.Assert.NoError(err) {
		tt.Assert.Len(fees, 3)
	}
}
