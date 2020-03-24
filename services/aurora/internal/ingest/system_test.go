package ingest

import (
	"fmt"
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestBackfill(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()
	is := sys(tt, Config{EnableAssetStats: false, CursorName: "HORIZON"})

	err := is.ReingestSingle(10)
	tt.Require.NoError(err)
	tt.UpdateLedgerState()

	// ensure 1 ledger
	var found int
	err = tt.AuroraSession().GetRaw(&found, "SELECT COUNT(*) FROM history_ledgers")
	tt.Require.NoError(err)
	tt.Require.Equal(1, found)

	err = is.Backfill(3)
	if tt.Assert.NoError(err) {
		err = tt.AuroraSession().GetRaw(&found, "SELECT COUNT(*) FROM history_ledgers")
		tt.Require.NoError(err)

		tt.Assert.Equal(4, found, "expected 4 ledgers to be in history database, but got %d", found)
	}
}

func TestClearAll(t *testing.T) {
	tt := test.Start(t).Scenario("kahuna")
	defer tt.Finish()
	is := sys(tt, Config{EnableAssetStats: false, CursorName: "HORIZON"})

	err := is.ClearAll()

	tt.Require.NoError(err)

	// ensure all tables are cleared
	tables := []TableName{
		AssetStatsTableName,
		AccountsTableName,
		AssetsTableName,
		EffectsTableName,
		LedgersTableName,
		OperationParticipantsTableName,
		OperationsTableName,
		TradesTableName,
		TransactionParticipantsTableName,
		TransactionsTableName,
	}

	for _, tableName := range tables {
		ensureEmpty(tt, tableName)
	}
}

func ensureEmpty(tt *test.T, tableName TableName) {
	var found int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", string(tableName))
	err := tt.AuroraSession().GetRaw(&found, query)
	tt.Require.NoError(err)
	tt.Assert.Equal(0, found)
}
