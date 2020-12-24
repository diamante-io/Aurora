package aurora

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/protocols/aurora/operations"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
)

func TestPaymentActions(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	w := ht.Get("/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(4, w.Body)
	}

	// filtered by ledger
	w = ht.Get("/ledgers/1/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(0, w.Body)
	}

	w = ht.Get("/ledgers/3/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
	}

	// filtered by account
	w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
	}

	// switch scenarios
	ht.T.Scenario("pathed_payment")

	// filtered by transaction
	w = ht.Get("/transactions/b52f16ffb98c047e33b9c2ec30880330cde71f85b3443dae2c5cb86c7d4d8452/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(0, w.Body)
	}

	// 400 for invalid tx hash
	w = ht.Get("/transactions/ /payments")
	ht.Assert.Equal(400, w.Code)

	w = ht.Get("/transactions/invalid/payments")
	ht.Assert.Equal(400, w.Code)

	w = ht.Get("/transactions/1d2a4be72470658f68db50eef29ea0af3f985ce18b5c218f03461d40c47dc29/payments")
	ht.Assert.Equal(400, w.Code)

	w = ht.Get("/transactions/1d2a4be72470658f68db50eef29ea0af3f985ce18b5c218f03461d40c47dc29222/payments")
	ht.Assert.Equal(400, w.Code)

	w = ht.Get("/transactions/1d2a4be72470658f68db50eef29ea0af3f985ce18b5c218f03461d40c47dc292/payments")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)

		// test for existence of source_amount in path payment details
		var records []map[string]interface{}
		ht.UnmarshalPage(w.Body, &records)
		ht.Assert.Equal("10.0000000", records[0]["source_amount"])
	}

	// Regression: negative cursor
	w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/payments?cursor=-23667108046966785&order=asc&limit=100")
	ht.Assert.Equal(400, w.Code)
}

func TestPaymentActions_Includetransactions(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	w := ht.Get("/payments")
	ht.Assert.Equal(200, w.Code)
	withoutTransactions := []operations.Base{}
	ht.UnmarshalPage(w.Body, &withoutTransactions)

	w = ht.Get("/payments?join=transactions")
	ht.Assert.Equal(200, w.Code)
	withTransactions := []operations.Base{}
	ht.UnmarshalPage(w.Body, &withTransactions)

	for i, operation := range withTransactions {
		getTransaction := ht.Get("/transactions/" + operation.Transaction.ID)
		ht.Assert.Equal(200, getTransaction.Code)
		var getTransactionResponse aurora.Transaction
		err := json.Unmarshal(getTransaction.Body.Bytes(), &getTransactionResponse)

		ht.Require.NoError(err, "failed to parse body")
		tx := operation.Transaction
		ht.Assert.Equal(*tx, getTransactionResponse)

		withTransactions[i].Transaction = nil
	}

	ht.Assert.Equal(withoutTransactions, withTransactions)
}

func TestPaymentActions_Show_Failed(t *testing.T) {
	ht := StartHTTPTest(t, "failed_transactions")
	defer ht.Finish()

	// Should show successful transactions only
	w := ht.Get("/payments?limit=200")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		successful := 0
		failed := 0

		for _, op := range records {
			if op.TransactionSuccessful {
				successful++
			} else {
				failed++
			}
		}

		ht.Assert.Equal(5, successful)
		ht.Assert.Equal(0, failed)
	}

	// Should show all transactions: both successful and failed
	w = ht.Get("/payments?limit=200&include_failed=true")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		successful := 0
		failed := 0

		for _, op := range records {
			if op.TransactionSuccessful {
				successful++
			} else {
				failed++
			}
		}

		ht.Assert.Equal(5, successful)
		ht.Assert.Equal(1, failed)
	}

	w = ht.Get("/transactions/aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf/payments")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		ht.Assert.Equal(1, len(records))
		for _, op := range records {
			ht.Assert.False(op.TransactionSuccessful)
		}
	}

	w = ht.Get("/transactions/56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1/payments")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		ht.Assert.Equal(1, len(records))
		for _, op := range records {
			ht.Assert.True(op.TransactionSuccessful)
		}
	}

	// NULL value
	_, err := ht.AuroraSession().ExecRaw(
		`UPDATE history_transactions SET successful = NULL WHERE transaction_hash = ?`,
		"56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1",
	)
	ht.Require.NoError(err)

	w = ht.Get("/transactions/56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1/payments")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		ht.Assert.Equal(1, len(records))
		for _, op := range records {
			ht.Assert.True(op.TransactionSuccessful)
		}
	}
}

func TestPayment_CreatedAt(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	w := ht.Get("/ledgers/3/payments")
	records := []operations.Base{}
	ht.UnmarshalPage(w.Body, &records)

	l := history.Ledger{}
	hq := history.Q{Session: ht.AuroraSession()}
	ht.Require.NoError(hq.LedgerBySequence(&l, 3))

	ht.Assert.WithinDuration(l.ClosedAt, records[0].LedgerCloseTime, 1*time.Second)
}

// TestPaymentActions_Show_Extra_TxID tests if failed transactions are not returned
// when `tx_id` GET param is present. This was happening because `base.GetString()`
// method retuns values from the query when URL param is not present.
func TestPaymentActions_Show_Extra_TxID(t *testing.T) {
	ht := StartHTTPTest(t, "failed_transactions")
	defer ht.Finish()

	w := ht.Get("/accounts/GBXGQJWVLWOYHFLVTKWV5FGHA3LNYY2JQKM7OAJAUEQFU6LPCSEFVXON/payments?limit=200&tx_id=abc")

	if ht.Assert.Equal(200, w.Code) {
		records := []operations.Base{}
		ht.UnmarshalPage(w.Body, &records)

		successful := 0
		failed := 0

		for _, op := range records {
			if op.TransactionSuccessful {
				successful++
			} else {
				failed++
			}
		}

		ht.Assert.Equal(2, successful)
		ht.Assert.Equal(0, failed)
	}
}
