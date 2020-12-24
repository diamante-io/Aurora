package aurora

import (
	"testing"

	"github.com/diamnet/go/protocols/aurora/effects"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/test"
)

func TestEffectActions_Index(t *testing.T) {

	t.Run("omnibus test", func(t *testing.T) {

		ht := StartHTTPTest(t, "base")
		defer ht.Finish()

		w := ht.Get("/effects?limit=20")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(11, w.Body)
		}

		// test streaming, regression for https://github.com/diamnet/go/services/aurora/internal/issues/147
		w = ht.Get("/effects?limit=2", test.RequestHelperStreaming)
		ht.Assert.Equal(200, w.Code)

		// filtered by ledger
		w = ht.Get("/ledgers/1/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(0, w.Body)
		}

		w = ht.Get("/ledgers/2/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(9, w.Body)
		}

		w = ht.Get("/ledgers/3/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(2, w.Body)
		}

		// filtered by account
		w = ht.Get("/accounts/GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(3, w.Body)
		}

		w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(2, w.Body)
		}

		w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(3, w.Body)
		}

		// filtered by transaction
		w = ht.Get("/transactions/2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(3, w.Body)
		}

		// filtered by operation
		w = ht.Get("/operations/8589938689/effects")
		if ht.Assert.Equal(200, w.Code) {
			ht.Assert.PageOf(3, w.Body)
		}

		// Check extra params
		w = ht.Get("/ledgers/100/effects?account_id=GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
		ht.Assert.Equal(400, w.Code)
		w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/effects?ledger_id=5")
		ht.Assert.Equal(400, w.Code)

		// before history
		ht.ReapHistory(1)
		w = ht.Get("/effects?order=desc&cursor=8589938689-1")
		ht.Assert.Equal(410, w.Code)
		ht.Logger.Error(w.Body.String())
	})

	t.Run("Effect resource props", func(t *testing.T) {
		ht := StartHTTPTest(t, "base")
		defer ht.Finish()

		// created_at
		w := ht.Get("/ledgers/2/effects")
		if ht.Assert.Equal(200, w.Code) {
			var result []effects.Base
			_ = ht.UnmarshalPage(w.Body, &result)
			ht.Require.NotEmpty(result, "unexpected empty response")

			e1 := result[0]

			var ledger2 history.Ledger
			err := ht.AuroraDB.Get(&ledger2, "SELECT * FROM history_ledgers WHERE sequence = 2")
			ht.Require.NoError(err, "failed to load ledger")

			ht.Assert.Equal(ledger2.ClosedAt.UTC(), e1.LedgerCloseTime.UTC())
		}
	})
}
