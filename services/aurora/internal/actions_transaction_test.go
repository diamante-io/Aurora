package aurora

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/txsub"
	"github.com/diamnet/go/services/aurora/internal/txsub/sequence"
)

func TestTransactionActions_Show(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	w := ht.Get("/transactions/2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.Equal(
			"2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d",
			actual.Hash,
		)
	}

	// missing tx
	w = ht.Get("/transactions/not_real")
	ht.Assert.Equal(404, w.Code)
}

func TestTransactionActions_Show_Failed(t *testing.T) {
	ht := StartHTTPTest(t, "failed_transactions")
	defer ht.Finish()

	// Failed single
	w := ht.Get("/transactions/aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.Equal(
			"aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf",
			actual.Hash,
		)

		ht.Assert.False(actual.Successful)
	}

	// Successful single
	w = ht.Get("/transactions/56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.Equal(
			"56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1",
			actual.Hash,
		)

		ht.Assert.True(actual.Successful)
	}

	// Should show successful transactions only
	w = ht.Get("/transactions?limit=200")

	if ht.Assert.Equal(200, w.Code) {
		records := []aurora.Transaction{}
		ht.UnmarshalPage(w.Body, &records)

		successful := 0
		failed := 0

		for _, tx := range records {
			if tx.Successful {
				successful++
			} else {
				failed++
			}
		}

		ht.Assert.Equal(8, successful)
		ht.Assert.Equal(0, failed)
	}

	// Should show all transactions: both successful and failed
	w = ht.Get("/transactions?limit=200&include_failed=true")

	if ht.Assert.Equal(200, w.Code) {
		records := []aurora.Transaction{}
		ht.UnmarshalPage(w.Body, &records)

		successful := 0
		failed := 0

		for _, tx := range records {
			if tx.Successful {
				successful++
			} else {
				failed++
			}
		}

		ht.Assert.Equal(8, successful)
		ht.Assert.Equal(1, failed)
	}

	w = ht.Get("/transactions/aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.False(actual.Successful)
	}

	w = ht.Get("/transactions/56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.True(actual.Successful)
	}

	// NULL value
	_, err := ht.AuroraSession().ExecRaw(
		`UPDATE history_transactions SET successful = NULL WHERE transaction_hash = ?`,
		"56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1",
	)
	ht.Require.NoError(err)

	w = ht.Get("/transactions/56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Transaction
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)

		ht.Assert.True(actual.Successful)
	}
}

func TestTransactionActions_Index(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	w := ht.Get("/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(4, w.Body)
	}

	// filtered by ledger
	w = ht.Get("/ledgers/1/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(0, w.Body)
	}

	w = ht.Get("/ledgers/2/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(3, w.Body)
	}

	w = ht.Get("/ledgers/3/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
	}

	// missing ledger
	w = ht.Get("/ledgers/100/transactions")
	ht.Assert.Equal(404, w.Code)

	// filtering by account
	w = ht.Get("/accounts/GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(3, w.Body)
	}

	w = ht.Get("/accounts/GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(1, w.Body)
	}

	w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/transactions")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.PageOf(2, w.Body)
	}

	// Check extra params
	w = ht.Get("/ledgers/100/transactions?account_id=GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
	ht.Assert.Equal(400, w.Code)
	w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/transactions?ledger_id=5")
	ht.Assert.Equal(400, w.Code)
	w = ht.Get("/accounts/GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU/transactions?cursor=limit=order=")
	ht.Assert.Equal(400, w.Code)

	// regression: https://github.com/diamnet/go/services/aurora/internal/issues/365
	w = ht.Get("/transactions?limit=200")
	ht.Require.Equal(200, w.Code)
	w = ht.Get("/transactions?limit=201")
	ht.Assert.Equal(400, w.Code)
	w = ht.Get("/transactions?limit=0")
	ht.Assert.Equal(400, w.Code)

}

func TestTransactionActions_Post(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	form := url.Values{"tx": []string{"AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAArqN6LeOagjxMaUP96Bzfs9e0corNZXzBWJkFoK7kvkwAAAAAO5rKAAAAAAAAAAABVvwF9wAAAECDzqvkQBQoNAJifPRXDoLhvtycT3lFPCQ51gkdsFHaBNWw05S/VhW0Xgkr0CBPE4NaFV2Kmcs3ZwLmib4TRrML"}}

	// existing transaction
	w := ht.Post("/transactions", form)
	ht.Assert.Equal(200, w.Code)

	// sequence buffer full
	ht.App.submitter.Results = &txsub.MockResultProvider{
		Results: []txsub.Result{
			{Err: sequence.ErrNoMoreRoom},
		},
	}
	w = ht.Post("/transactions", form)
	ht.Assert.Equal(503, w.Code)
}

func TestTransactionActions_PostSuccessful(t *testing.T) {
	ht := StartHTTPTest(t, "failed_transactions")
	defer ht.Finish()

	// 56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1
	form := url.Values{"tx": []string{"AAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAZAAAAAIAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAbmgm1V2dg5V1mq1elMcG1txjSYKZ9wEgoSBaeW8UiFoAAAABVVNEAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAA7msoAAAAAAAAAAAGu5L5MAAAAQEnKDbDYvKkJjYK0arvhFln+GK0+7Ay6g0a+1hjRRelEAe4wmjeqNcRg2m4Cn7t4AjJzAsDQI0iXahGboJPINAw="}}

	w := ht.Post("/transactions", form)
	ht.Assert.Equal(200, w.Code)
	ht.Assert.Contains(string(w.Body.Bytes()), `"result_xdr": "AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA="`)
}

func TestTransactionActions_PostFailed(t *testing.T) {
	ht := StartHTTPTest(t, "failed_transactions")
	defer ht.Finish()

	// aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf
	form := url.Values{"tx": []string{"AAAAAG5oJtVdnYOVdZqtXpTHBtbcY0mCmfcBIKEgWnlvFIhaAAAAZAAAAAIAAAACAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAO2C/AO45YBD3tHVFO1R3A0MekP8JR6nN1A9eWidyItUAAAABVVNEAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAB3NZQAAAAAAAAAAAFvFIhaAAAAQKcGS9OsVnVHCVIH04C9ZKzzKYBRdCmy+Jwmzld7QcALOxZUcAgkuGfoSdvXpH38mNvrqQiaMsSNmTJWYRzHvgo="}}

	w := ht.Post("/transactions", form)
	ht.Assert.Equal(400, w.Code)
	ht.Assert.Contains(string(w.Body.Bytes()), "op_underfunded")
	ht.Assert.Contains(string(w.Body.Bytes()), `"result_xdr": "AAAAAAAAAGT/////AAAAAQAAAAAAAAAB/////gAAAAA="`)
}
