package resourceadapter

import (
	"testing"

	"github.com/hcnet/go/protocols/aurora/operations"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/support/test"
	"github.com/stretchr/testify/assert"
)

// TestPopulateOperation_Successful tests operation object population.
func TestPopulateOperation_Successful(t *testing.T) {
	ctx, _ := test.ContextWithLogBuffer()

	var (
		dest   operations.Base
		row    history.Operation
		val    bool
		ledger = history.Ledger{}
	)

	dest = operations.Base{}
	row = history.Operation{TransactionSuccessful: nil}

	PopulateBaseOperation(ctx, &dest, row, nil, ledger)
	assert.True(t, dest.TransactionSuccessful)
	assert.Nil(t, dest.Transaction)

	dest = operations.Base{}
	val = true
	row = history.Operation{TransactionSuccessful: &val}

	PopulateBaseOperation(ctx, &dest, row, nil, ledger)
	assert.True(t, dest.TransactionSuccessful)
	assert.Nil(t, dest.Transaction)

	dest = operations.Base{}
	val = false
	row = history.Operation{TransactionSuccessful: &val}

	PopulateBaseOperation(ctx, &dest, row, nil, ledger)
	assert.False(t, dest.TransactionSuccessful)
	assert.Nil(t, dest.Transaction)
}

// TestPopulateOperation_WithTransaction tests PopulateBaseOperation when passing both an operation and a transaction.
func TestPopulateOperation_WithTransaction(t *testing.T) {
	ctx, _ := test.ContextWithLogBuffer()

	var (
		dest           operations.Base
		operationsRow  history.Operation
		val            bool
		ledger         = history.Ledger{}
		transactionRow history.Transaction
	)

	dest = operations.Base{}
	val = true
	operationsRow = history.Operation{TransactionSuccessful: &val}
	transactionRow = history.Transaction{Successful: &val, MaxFee: 10000, FeeCharged: 100}

	PopulateBaseOperation(ctx, &dest, operationsRow, &transactionRow, ledger)
	assert.True(t, dest.TransactionSuccessful)
	assert.True(t, dest.Transaction.Successful)
	assert.Equal(t, int32(100), dest.Transaction.FeePaid)
}
