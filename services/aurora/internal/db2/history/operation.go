package history

import (
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-errors/errors"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/toid"
	"github.com/hcnet/go/xdr"
)

func (t *Operation) IsTransactionSuccessful() bool {
	if t.TransactionSuccessful == nil {
		return true
	}

	return *t.TransactionSuccessful
}

// LedgerSequence return the ledger in which the effect occurred.
func (r *Operation) LedgerSequence() int32 {
	id := toid.Parse(r.ID)
	return id.LedgerSequence
}

// UnmarshalDetails unmarshals the details of this operation into `dest`
func (r *Operation) UnmarshalDetails(dest interface{}) error {
	if !r.DetailsString.Valid {
		return nil
	}

	err := json.Unmarshal([]byte(r.DetailsString.String), &dest)
	if err != nil {
		err = errors.Wrap(err, 1)
	}

	return err
}

// OperationFeeStats returns operation fee stats for the last 5 ledgers.
// Currently, we hard code the query to return the last 5 ledgers worth of transactions.
// TODO: make the number of ledgers configurable.
func (q *Q) OperationFeeStats(currentSeq int32, dest *FeeStats) error {
	return q.GetRaw(dest, `
		SELECT
			ceil(min(max_fee/operation_count))::bigint AS "min",
			ceil(mode() within group (order by max_fee/operation_count))::bigint AS "mode",
			ceil(percentile_cont(0.10) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p10",
			ceil(percentile_cont(0.20) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p20",
			ceil(percentile_cont(0.30) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p30",
			ceil(percentile_cont(0.40) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p40",
			ceil(percentile_cont(0.50) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p50",
			ceil(percentile_cont(0.60) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p60",
			ceil(percentile_cont(0.70) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p70",
			ceil(percentile_cont(0.80) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p80",
			ceil(percentile_cont(0.90) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p90",
			ceil(percentile_cont(0.95) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p95",
			ceil(percentile_cont(0.99) WITHIN GROUP (ORDER BY max_fee/operation_count))::bigint AS "p99"
		FROM history_transactions
		WHERE ledger_sequence > $1 AND ledger_sequence <= $2
	`, currentSeq-5, currentSeq)
}

// Operations provides a helper to filter the operations table with pre-defined
// filters.  See `OperationsQ` for the available filters.
func (q *Q) Operations() *OperationsQ {
	query := &OperationsQ{
		parent:              q,
		opIdCol:             "hop.id",
		includeFailed:       false,
		includeTransactions: false,
		sql:                 selectOperation,
	}

	return query
}

// OperationByID returns an Operation and optionally a Transaction given an operation id
func (q *Q) OperationByID(includeTransactions bool, id int64) (Operation, *Transaction, error) {
	sql := selectOperation.
		Limit(1).
		Where("hop.id = ?", id)

	var operation Operation
	err := q.Get(&operation, sql)
	if err != nil {
		return operation, nil, err
	}

	if includeTransactions {
		var transaction Transaction
		if err := q.TransactionByHash(&transaction, operation.TransactionHash); err != nil {
			return operation, nil, err
		}

		return operation, &transaction, err
	}
	return operation, nil, err
}

// ForAccount filters the operations collection to a specific account
func (q *OperationsQ) ForAccount(aid string) *OperationsQ {
	var account Account
	q.Err = q.parent.AccountByAddress(&account, aid)
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.Join(
		"history_operation_participants hopp ON "+
			"hopp.history_operation_id = hop.id",
	).Where("hopp.history_account_id = ?", account.ID)

	// in order to use history_operation_participants.hist_op_p_id index
	q.opIdCol = "hopp.history_operation_id"

	return q
}

// ForLedger filters the query to a only operations in a specific ledger,
// specified by its sequence.
func (q *OperationsQ) ForLedger(seq int32) *OperationsQ {
	var ledger Ledger
	q.Err = q.parent.LedgerBySequence(&ledger, seq)
	if q.Err != nil {
		return q
	}

	start := toid.ID{LedgerSequence: seq}
	end := toid.ID{LedgerSequence: seq + 1}
	q.sql = q.sql.Where(
		"hop.id >= ? AND hop.id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForTransaction filters the query to a only operations in a specific
// transaction, specified by the transactions's hex-encoded hash.
func (q *OperationsQ) ForTransaction(hash string) *OperationsQ {
	var tx Transaction
	q.Err = q.parent.TransactionByHash(&tx, hash)
	if q.Err != nil {
		return q
	}

	start := toid.Parse(tx.ID)
	end := start
	end.TransactionOrder++
	q.sql = q.sql.Where(
		"hop.id >= ? AND hop.id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// OnlyPayments filters the query being built to only include operations that
// are in the "payment" class of operations:  CreateAccountOps, Payments, and
// PathPayments.
func (q *OperationsQ) OnlyPayments() *OperationsQ {
	q.sql = q.sql.Where(sq.Eq{"hop.type": []xdr.OperationType{
		xdr.OperationTypeCreateAccount,
		xdr.OperationTypePayment,
		xdr.OperationTypePathPayment,
		xdr.OperationTypeAccountMerge,
	}})
	return q
}

// IncludeFailed changes the query to include failed transactions.
func (q *OperationsQ) IncludeFailed() *OperationsQ {
	q.includeFailed = true
	return q
}

// IncludeTransactions changes the query to fetch transaction data in addition to operation records.
func (q *OperationsQ) IncludeTransactions() *OperationsQ {
	q.includeTransactions = true
	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *OperationsQ) Page(page db2.PageQuery) *OperationsQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, q.opIdCol)
	return q
}

// Fetch returns results specified by a filtered operations query
func (q *OperationsQ) Fetch() ([]Operation, []Transaction, error) {
	if q.Err != nil {
		return nil, nil, q.Err
	}

	if !q.includeFailed {
		q.sql = q.sql.
			Where("(ht.successful = true OR ht.successful IS NULL)")
	}

	var operations []Operation
	var transactions []Transaction
	q.Err = q.parent.Select(&operations, q.sql)
	if q.Err != nil {
		return nil, nil, q.Err
	}
	set := map[int64]bool{}
	transactionIDs := []int64{}

	for _, o := range operations {
		var resultXDR xdr.TransactionResult
		err := xdr.SafeUnmarshalBase64(o.TxResult, &resultXDR)
		if err != nil {
			return nil, nil, err
		}

		if !set[o.TransactionID] {
			set[o.TransactionID] = true
			transactionIDs = append(transactionIDs, o.TransactionID)
		}

		if !q.includeFailed {
			if !o.IsTransactionSuccessful() {
				return nil, nil, errors.Errorf("Corrupted data! `include_failed=false` but returned transaction is failed: %s", o.TransactionHash)
			}

			if resultXDR.Result.Code != xdr.TransactionResultCodeTxSuccess {
				return nil, nil, errors.Errorf("Corrupted data! `include_failed=false` but returned transaction is failed: %s %s", o.TransactionHash, o.TxResult)
			}
		}

		// Check if `successful` equals resultXDR
		if o.IsTransactionSuccessful() && resultXDR.Result.Code != xdr.TransactionResultCodeTxSuccess {
			return nil, nil, errors.Errorf("Corrupted data! `successful=true` but returned transaction is not success: %s %s", o.TransactionHash, o.TxResult)
		}

		if !o.IsTransactionSuccessful() && resultXDR.Result.Code == xdr.TransactionResultCodeTxSuccess {
			return nil, nil, errors.Errorf("Corrupted data! `successful=false` but returned transaction is success: %s %s", o.TransactionHash, o.TxResult)
		}
	}

	if q.includeTransactions && len(transactionIDs) > 0 {
		transactionsByID, err := q.parent.TransactionsByIDs(transactionIDs...)
		if err != nil {
			return nil, nil, err
		}
		for _, o := range operations {
			if transaction, ok := transactionsByID[o.TransactionID]; !ok {
				return nil, nil, errors.Errorf("transaction with id %v could not be found", o.TransactionID)
			} else {
				transactions = append(transactions, transaction)
			}
		}
	}

	return operations, transactions, nil
}

var selectOperation = sq.Select(
	"hop.id, " +
		"hop.transaction_id, " +
		"hop.application_order, " +
		"hop.type, " +
		"hop.details, " +
		"hop.source_account, " +
		"ht.transaction_hash, " +
		"ht.tx_result, " +
		"ht.successful as transaction_successful").
	From("history_operations hop").
	LeftJoin("history_transactions ht ON ht.id = hop.transaction_id")
