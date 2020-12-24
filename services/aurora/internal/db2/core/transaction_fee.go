package core

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/xdr"
)

// ChangesXDR returns the XDR encoded changes for this transaction fee
func (fee *TransactionFee) ChangesXDR() string {
	out, err := xdr.MarshalBase64(fee.Changes)
	if err != nil {
		panic(err)
	}
	return out
}

// TransactionFeesByLedger is a query that loads all rows from `txfeehistory`
// where ledgerseq matches `Sequence.`
func (q *Q) TransactionFeesByLedger(dest interface{}, seq int32) error {
	sql := sq.Select("ctxfh.*").
		From("txfeehistory ctxfh").
		OrderBy("ctxfh.txindex ASC").
		Where("ctxfh.ledgerseq = ?", seq)

	return q.Select(dest, sql)
}
