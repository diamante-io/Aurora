// Package results provides an implementation of the txsub.ResultProvider interface
// backed using the SQL databases used by both diamnet core and aurora
package results

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/txsub"
	"github.com/diamnet/go/xdr"
)

// DB provides transactio submission results by querying the
// connected aurora and diamnet core databases.
type DB struct {
	Core    *core.Q
	History *history.Q
}

var _ txsub.ResultProvider = &DB{}

// ResultByHash implements txsub.ResultProvider
func (rp *DB) ResultByHash(ctx context.Context, hash string) txsub.Result {
	historyLatest := ledger.CurrentState().HistoryLatest

	// query history database
	var hr history.Transaction
	err := rp.History.TransactionByHash(&hr, hash)
	if err == nil {
		return txResultFromHistory(hr)
	}

	if !rp.History.NoRows(err) {
		return txsub.Result{Err: err}
	}

	// query core database
	var cr core.Transaction
	// In the past we were searching for the transaction in core DB *after* the
	// latest ingested ledger. This was incorrect because history DB contains
	// successful transactions only. So it was possible that the transaction was
	// never found and clients were receiving Timeout errors.
	// However we can't change it to simply find a transaction by hash because
	// `txhistory` table does not have an index on `txid` field. Because of this
	// we query the last 120 ledgers (~10 minutes) to not kill the DB by searching
	// for a value on a table with millions of rows but also to support returning
	// the failed tx result (when resubmitting) for 10 minutes (or before core
	// clears `txhistory` table, whatever is first).
	// If you are modifying the code here, please do not make this error again.
	err = rp.Core.TransactionByHashAfterLedger(&cr, hash, historyLatest-120)
	if err == nil {
		return txResultFromCore(cr)
	}

	if !rp.Core.NoRows(err) {
		return txsub.Result{Err: err}
	}

	// if no result was found in either db, return ErrNoResults
	return txsub.Result{Err: txsub.ErrNoResults}
}

func txResultFromHistory(tx history.Transaction) txsub.Result {
	var txResult xdr.TransactionResult
	err := xdr.SafeUnmarshalBase64(tx.TxResult, &txResult)
	if err == nil {
		if txResult.Result.Code != xdr.TransactionResultCodeTxSuccess {
			err = &txsub.FailedTransactionError{
				ResultXDR: tx.TxResult,
			}
		}
	}

	return txsub.Result{
		Err:            err,
		Hash:           tx.TransactionHash,
		LedgerSequence: tx.LedgerSequence,
		EnvelopeXDR:    tx.TxEnvelope,
		ResultXDR:      tx.TxResult,
		ResultMetaXDR:  tx.TxMeta,
	}
}

func txResultFromCore(tx core.Transaction) txsub.Result {
	// re-encode result to base64
	var raw bytes.Buffer
	_, err := xdr.Marshal(&raw, tx.Result.Result)

	if err != nil {
		return txsub.Result{Err: err}
	}

	trx := base64.StdEncoding.EncodeToString(raw.Bytes())

	// if result is success, send a normal resposne
	if tx.Result.Result.Result.Code == xdr.TransactionResultCodeTxSuccess {
		return txsub.Result{
			Hash:           tx.TransactionHash,
			LedgerSequence: tx.LedgerSequence,
			EnvelopeXDR:    tx.EnvelopeXDR(),
			ResultXDR:      trx,
			ResultMetaXDR:  tx.ResultMetaXDR(),
		}
	}

	// if failed, produce a FailedTransactionError
	return txsub.Result{
		Err: &txsub.FailedTransactionError{
			ResultXDR: trx,
		},
		Hash:           tx.TransactionHash,
		LedgerSequence: tx.LedgerSequence,
		EnvelopeXDR:    tx.EnvelopeXDR(),
		ResultXDR:      trx,
		ResultMetaXDR:  tx.ResultMetaXDR(),
	}
}
