package core

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/xdr"
)

// DataXDR returns the base64 encoded ledger header
func (lh *LedgerHeader) DataXDR() string {
	out, err := xdr.MarshalBase64(lh.Data)
	if err != nil {
		panic(err)
	}
	return out
}

// LedgerHeaderBySequence is a query that loads a single row from the
// `ledgerheaders` table.
func (q *Q) LedgerHeaderBySequence(dest interface{}, seq int32) error {
	sql := sq.Select("clh.*").
		From("ledgerheaders clh").
		Limit(1).
		Where("clh.ledgerseq = ?", seq)

	return q.Get(dest, sql)
}
