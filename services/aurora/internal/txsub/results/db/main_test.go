package results

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/services/aurora/internal/txsub"
)

func TestResultProvider(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	rp := &DB{
		Core:    &core.Q{Session: tt.CoreSession()},
		History: &history.Q{Session: tt.AuroraSession()},
	}

	// Regression: ensure a transaction that is not ingested still returns the
	// result
	hash := "2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d"
	ret := rp.ResultByHash(tt.Ctx, hash)

	tt.Require.NoError(ret.Err)
	tt.Assert.Equal(hash, ret.Hash)
}

func TestResultFailed(t *testing.T) {
	tt := test.Start(t).Scenario("failed_transactions")
	defer tt.Finish()

	rp := &DB{
		Core:    &core.Q{Session: tt.CoreSession()},
		History: &history.Q{Session: tt.AuroraSession()},
	}

	hash := "aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf"

	// Ignore core db results
	_, err := tt.CoreSession().ExecRaw(
		`DELETE FROM txhistory WHERE txid = ?`,
		hash,
	)
	tt.Require.NoError(err)

	ret := rp.ResultByHash(tt.Ctx, hash)

	tt.Require.Error(ret.Err)
	tt.Assert.Equal("AAAAAAAAAGT/////AAAAAQAAAAAAAAAB/////gAAAAA=", ret.Err.(*txsub.FailedTransactionError).ResultXDR)
	tt.Assert.Equal(hash, ret.Hash)
}
