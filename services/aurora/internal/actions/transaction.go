package actions

import (
	"context"

	"github.com/diamnet/go/clients/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/xdr"
)

// TransactionPage returns a page containing the transaction records of an
// account/ledger identified by accountID/ledgerID into a page based on pq and
// includeFailedTx.
func TransactionPage(ctx context.Context, hq *history.Q, accountID string, ledgerID int32, includeFailedTx bool, pq db2.PageQuery) (hal.Page, error) {
	records, err := loadTransactionRecords(hq, accountID, ledgerID, includeFailedTx, pq)
	if err != nil {
		return hal.Page{}, errors.Wrap(err, "loading transaction records")
	}

	page := hal.Page{
		Cursor: pq.Cursor,
		Order:  pq.Order,
		Limit:  pq.Limit,
	}

	for _, record := range records {
		// TODO: make PopulateTransaction return aurora.Transaction directly.
		var res aurora.Transaction
		resourceadapter.PopulateTransaction(ctx, &res, record)
		page.Add(res)
	}

	page.FullURL = fullURL(ctx)
	page.PopulateLinks()
	return page, nil
}

// loadTransactionRecords returns a slice of transaction records of an
// account/ledger identified by accountID/ledgerID based on pq and
// includeFailedTx.
func loadTransactionRecords(hq *history.Q, accountID string, ledgerID int32, includeFailedTx bool, pq db2.PageQuery) ([]history.Transaction, error) {
	if accountID != "" && ledgerID != 0 {
		return nil, errors.New("conflicting exclusive fields are present: account_id and ledger_id")
	}

	var records []history.Transaction

	txs := hq.Transactions()
	switch {
	case accountID != "":
		txs.ForAccount(accountID)
	case ledgerID > 0:
		txs.ForLedger(ledgerID)
	}

	if includeFailedTx {
		txs.IncludeFailed()
	}

	err := txs.Page(pq).Select(&records)
	if err != nil {
		return nil, errors.Wrap(err, "executing transaction records query")
	}

	for _, t := range records {
		if !includeFailedTx {
			if !t.IsSuccessful() {
				return nil, errors.Errorf("Corrupted data! `include_failed=false` but returned transaction is failed: %s", t.TransactionHash)
			}

			var resultXDR xdr.TransactionResult
			err = xdr.SafeUnmarshalBase64(t.TxResult, &resultXDR)
			if err != nil {
				return nil, errors.Wrap(err, "unmarshalling tx result")
			}

			if resultXDR.Result.Code != xdr.TransactionResultCodeTxSuccess {
				return nil, errors.Errorf("Corrupted data! `include_failed=false` but returned transaction is failed: %s %s", t.TransactionHash, t.TxResult)
			}
		}
	}

	return records, nil
}

// StreamTransactions streams transaction records of an account/ledger
// identified by accountID/ledgerID based on pq and includeFailedTx.
func StreamTransactions(ctx context.Context, s *sse.Stream, hq *history.Q, accountID string, ledgerID int32, includeFailedTx bool, pq db2.PageQuery) error {
	allRecords, err := loadTransactionRecords(hq, accountID, ledgerID, includeFailedTx, pq)
	if err != nil {
		return errors.Wrap(err, "loading transaction records")
	}

	s.SetLimit(int(pq.Limit))
	records := allRecords[s.SentCount():]
	for _, record := range records {
		var res aurora.Transaction
		resourceadapter.PopulateTransaction(ctx, &res, record)
		s.Send(sse.Event{ID: res.PagingToken(), Data: res})
	}

	return nil
}

// TransactionResource returns a single transaction resource identified by txHash.
func TransactionResource(ctx context.Context, hq *history.Q, txHash string) (aurora.Transaction, error) {
	var (
		record   history.Transaction
		resource aurora.Transaction
	)
	err := hq.TransactionByHash(&record, txHash)
	if err != nil {
		return resource, errors.Wrap(err, "loading transaction record")
	}

	resourceadapter.PopulateTransaction(ctx, &resource, record)
	return resource, nil
}
