package history

import (
	"encoding/json"
	"fmt"
	"math"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/toid"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// UnmarshalDetails unmarshals the details of this effect into `dest`
func (r *Effect) UnmarshalDetails(dest interface{}) error {
	if !r.DetailsString.Valid {
		return nil
	}

	return errors.Wrap(json.Unmarshal([]byte(r.DetailsString.String), &dest), "unmarshal effect details failed")
}

// ID returns a lexically ordered id for this effect record
func (r *Effect) ID() string {
	return fmt.Sprintf("%019d-%010d", r.HistoryOperationID, r.Order)
}

// LedgerSequence return the ledger in which the effect occurred.
func (r *Effect) LedgerSequence() int32 {
	id := toid.Parse(r.HistoryOperationID)
	return id.LedgerSequence
}

// PagingToken returns a cursor for this effect
func (r *Effect) PagingToken() string {
	return fmt.Sprintf("%d-%d", r.HistoryOperationID, r.Order)
}

// Effects provides a helper to filter rows from the `history_effects`
// table with pre-defined filters.  See `TransactionsQ` methods for the
// available filters.
func (q *Q) Effects() *EffectsQ {
	return &EffectsQ{
		parent: q,
		sql:    selectEffect,
	}
}

// ForAccount filters the operations collection to a specific account
func (q *EffectsQ) ForAccount(aid string) *EffectsQ {
	var account Account
	q.Err = q.parent.AccountByAddress(&account, aid)
	if q.Err != nil {
		return q
	}

	q.sql = q.sql.Where("heff.history_account_id = ?", account.ID)

	return q
}

// ForLedger filters the query to only effects in a specific ledger,
// specified by its sequence.
func (q *EffectsQ) ForLedger(seq int32) *EffectsQ {
	var ledger Ledger
	q.Err = q.parent.LedgerBySequence(&ledger, seq)
	if q.Err != nil {
		return q
	}

	start := toid.ID{LedgerSequence: seq}
	end := toid.ID{LedgerSequence: seq + 1}
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForOperation filters the query to only effects in a specific operation,
// specified by its id.
func (q *EffectsQ) ForOperation(id int64) *EffectsQ {
	start := toid.Parse(id)
	end := start
	end.IncOperationOrder()
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// ForOrderBook filters the query to only effects whose details indicate that
// the effect is for a specific asset pair.
func (q *EffectsQ) ForOrderBook(selling, buying xdr.Asset) *EffectsQ {
	q.orderBookFilter(selling, "sold_")
	if q.Err != nil {
		return q
	}
	q.orderBookFilter(buying, "bought_")
	if q.Err != nil {
		return q
	}

	return q
}

// ForTransaction filters the query to only effects in a specific
// transaction, specified by the transactions's hex-encoded hash.
func (q *EffectsQ) ForTransaction(hash string) *EffectsQ {
	var tx Transaction
	q.Err = q.parent.TransactionByHash(&tx, hash)
	if q.Err != nil {
		return q
	}

	start := toid.Parse(tx.ID)
	end := start
	end.TransactionOrder++
	q.sql = q.sql.Where(
		"heff.history_operation_id >= ? AND heff.history_operation_id < ?",
		start.ToInt64(),
		end.ToInt64(),
	)

	return q
}

// OfType filters the query to only effects of the given type.
func (q *EffectsQ) OfType(typ EffectType) *EffectsQ {
	q.sql = q.sql.Where("heff.type = ?", typ)
	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *EffectsQ) Page(page db2.PageQuery) *EffectsQ {
	if q.Err != nil {
		return q
	}

	op, idx, err := page.CursorInt64Pair(db2.DefaultPairSep)
	if err != nil {
		q.Err = err
		return q
	}

	if idx > math.MaxInt32 {
		idx = math.MaxInt32
	}

	// NOTE: Remember to test the queries below with EXPLAIN / EXPLAIN ANALYZE
	// before changing them.
	// This condition is using multicolumn index and it's easy to write it in a way that
	// DB will perform a full table scan.
	switch page.Order {
	case "asc":
		q.sql = q.sql.
			Where(`(
					 heff.history_operation_id >= ?
				AND (
					 heff.history_operation_id > ? OR
					(heff.history_operation_id = ? AND heff.order > ?)
				))`, op, op, op, idx).
			OrderBy("heff.history_operation_id asc, heff.order asc")
	case "desc":
		q.sql = q.sql.
			Where(`(
					 heff.history_operation_id <= ?
				AND (
					 heff.history_operation_id < ? OR
					(heff.history_operation_id = ? AND heff.order < ?)
				))`, op, op, op, idx).
			OrderBy("heff.history_operation_id desc, heff.order desc")
	}

	q.sql = q.sql.Limit(page.Limit)
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *EffectsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

// OfType filters the query to only effects of the given type.
func (q *EffectsQ) orderBookFilter(a xdr.Asset, prefix string) {
	var typ, code, iss string
	q.Err = a.Extract(&typ, &code, &iss)
	if q.Err != nil {
		return
	}

	if a.Type == xdr.AssetTypeAssetTypeNative {
		clause := fmt.Sprintf(`
				(heff.details->>'%sasset_type' = ?
		AND heff.details ?? '%sasset_code' = false
		AND heff.details ?? '%sasset_issuer' = false)`, prefix, prefix, prefix)
		q.sql = q.sql.Where(clause, typ)
		return
	}

	clause := fmt.Sprintf(`
		(heff.details->>'%sasset_type' = ?
	AND heff.details->>'%sasset_code' = ?
	AND heff.details->>'%sasset_issuer' = ?)`, prefix, prefix, prefix)
	q.sql = q.sql.Where(clause, typ, code, iss)
}

var selectEffect = sq.
	Select("heff.*, hacc.address").
	From("history_effects heff").
	LeftJoin("history_accounts hacc ON hacc.id = heff.history_account_id")
