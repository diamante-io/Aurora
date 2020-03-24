package reap

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestDeleteUnretainedHistory(t *testing.T) {
	tt := test.Start(t).Scenario("kahuna")
	defer tt.Finish()

	db := tt.AuroraSession()

	sys := New(0, db)

	var (
		prev int
		cur  int
	)
	err := db.GetRaw(&prev, `SELECT COUNT(*) FROM history_ledgers`)
	tt.Require.NoError(err)

	err = sys.DeleteUnretainedHistory()
	if tt.Assert.NoError(err) {
		err = db.GetRaw(&cur, `SELECT COUNT(*) FROM history_ledgers`)
		tt.Require.NoError(err)
		tt.Assert.Equal(prev, cur, "Ledgers deleted when RetentionCount == 0")
	}

	tt.UpdateLedgerState()
	sys.RetentionCount = 10
	err = sys.DeleteUnretainedHistory()
	if tt.Assert.NoError(err) {
		err = db.GetRaw(&cur, `SELECT COUNT(*) FROM history_ledgers`)
		tt.Require.NoError(err)
		tt.Assert.Equal(10, cur)
	}

	tt.UpdateLedgerState()
	sys.RetentionCount = 1
	err = sys.DeleteUnretainedHistory()
	if tt.Assert.NoError(err) {
		err = db.GetRaw(&cur, `SELECT COUNT(*) FROM history_ledgers`)
		tt.Require.NoError(err)
		tt.Assert.Equal(1, cur)
	}
}
