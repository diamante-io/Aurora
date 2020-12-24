package history

import (
	"database/sql"
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
)

func TestLedgerQueries(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	// Test LedgerBySequence
	var l Ledger
	err := q.LedgerBySequence(&l, 3)
	tt.Assert.NoError(err)

	err = q.LedgerBySequence(&l, 100000)
	tt.Assert.Equal(err, sql.ErrNoRows)

	// Test Ledgers()
	ls := []Ledger{}
	err = q.Ledgers().Select(&ls)

	if tt.Assert.NoError(err) {
		tt.Assert.Len(ls, 3)
	}

	// LedgersBySequence
	err = q.LedgersBySequence(&ls, 1, 2, 3)

	if tt.Assert.NoError(err) {
		tt.Assert.Len(ls, 3)

		foundSeqs := make([]int32, len(ls))
		for i := range ls {
			foundSeqs[i] = ls[i].Sequence
		}

		tt.Assert.Contains(foundSeqs, int32(1))
		tt.Assert.Contains(foundSeqs, int32(2))
		tt.Assert.Contains(foundSeqs, int32(3))
	}
}
