package history

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/support/errors"
)

// LedgerBySequence loads the single ledger at `seq` into `dest`
func (q *Q) LedgerBySequence(dest interface{}, seq int32) error {
	sql := selectLedger.
		Limit(1).
		Where("sequence = ?", seq)

	return q.Get(dest, sql)
}

// Ledgers provides a helper to filter rows from the `history_ledgers` table
// with pre-defined filters.  See `LedgersQ` methods for the available filters.
func (q *Q) Ledgers() *LedgersQ {
	return &LedgersQ{
		parent: q,
		sql:    selectLedger,
	}
}

// LedgersBySequence loads the a set of ledgers identified by the sequences
// `seqs` into `dest`.
func (q *Q) LedgersBySequence(dest interface{}, seqs ...int32) error {
	if len(seqs) == 0 {
		return errors.New("no sequence arguments provided")
	}
	in := fmt.Sprintf("sequence IN (%s)", sq.Placeholders(len(seqs)))

	whereArgs := make([]interface{}, len(seqs))
	for i, s := range seqs {
		whereArgs[i] = s
	}

	sql := selectLedger.Where(in, whereArgs...)

	return q.Select(dest, sql)
}

// LedgerCapacityUsageStats returns ledger capacity stats for the last 5 ledgers.
// Currently, we hard code the query to return the last 5 ledgers.
// TODO: make the number of ledgers configurable.
func (q *Q) LedgerCapacityUsageStats(currentSeq int32, dest *LedgerCapacityUsageStats) error {
	const ledgers int32 = 5
	return q.GetRaw(dest, `
		SELECT ROUND(SUM(CAST(operation_count as decimal))/SUM(max_tx_set_size), 2) as ledger_capacity_usage FROM
			(SELECT
			  hl.sequence, COALESCE(SUM(ht.operation_count), 0) as operation_count, hl.max_tx_set_size
			  FROM history_ledgers hl
			  LEFT JOIN history_transactions ht ON ht.ledger_sequence = hl.sequence
			  WHERE hl.sequence > $1 AND hl.sequence <= $2
			  GROUP BY hl.sequence, hl.max_tx_set_size) as a
	`, currentSeq-ledgers, currentSeq)
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *LedgersQ) Page(page db2.PageQuery) *LedgersQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "hl.id")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *LedgersQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

var selectLedger = sq.Select(
	"hl.id",
	"hl.sequence",
	"hl.importer_version",
	"hl.ledger_hash",
	"hl.previous_ledger_hash",
	"hl.transaction_count",
	"hl.successful_transaction_count",
	"hl.failed_transaction_count",
	"hl.operation_count",
	"hl.closed_at",
	"hl.created_at",
	"hl.updated_at",
	"hl.total_coins",
	"hl.fee_pool",
	"hl.base_fee",
	"hl.base_reserve",
	"hl.max_tx_set_size",
	"hl.protocol_version",
	"hl.ledger_header",
).From("history_ledgers hl")
