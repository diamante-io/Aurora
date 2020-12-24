package resourceadapter

import (
	"testing"

	. "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/support/test"
	"github.com/stretchr/testify/assert"
)

// TestPopulateTransaction_Successful tests transaction object population.
func TestPopulateTransaction_Successful(t *testing.T) {
	ctx, _ := test.ContextWithLogBuffer()

	var (
		dest Transaction
		row  history.Transaction
		val  bool
	)

	dest = Transaction{}
	row = history.Transaction{Successful: nil}

	PopulateTransaction(ctx, &dest, row)
	assert.True(t, dest.Successful)

	dest = Transaction{}
	val = true
	row = history.Transaction{Successful: &val}

	PopulateTransaction(ctx, &dest, row)
	assert.True(t, dest.Successful)

	dest = Transaction{}
	val = false
	row = history.Transaction{Successful: &val}

	PopulateTransaction(ctx, &dest, row)
	assert.False(t, dest.Successful)
}

// TestPopulateTransaction_Fee tests transaction object population.
func TestPopulateTransaction_Fee(t *testing.T) {
	ctx, _ := test.ContextWithLogBuffer()

	var (
		dest Transaction
		row  history.Transaction
	)

	dest = Transaction{}
	row = history.Transaction{MaxFee: 10000, FeeCharged: 100}

	PopulateTransaction(ctx, &dest, row)
	assert.Equal(t, int32(100), dest.FeePaid)
}
