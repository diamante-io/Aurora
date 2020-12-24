package core

import (
	"encoding/base64"
	"fmt"

	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null"
	"github.com/diamnet/go/services/aurora/internal/utf8"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/xdr"
)

// Base64Signatures returns a slice of strings where each element is a base64
// encoded representation of a signature attached to this transaction.
func (tx *Transaction) Base64Signatures() []string {
	raw := tx.Envelope.Signatures
	results := make([]string, len(raw))

	for i := range raw {
		results[i] = base64.StdEncoding.EncodeToString(raw[i].Signature)
	}
	return results
}

// EnvelopeXDR returns the XDR encoded envelope for this transaction
func (tx *Transaction) EnvelopeXDR() string {
	out, err := xdr.MarshalBase64(tx.Envelope)
	if err != nil {
		panic(err)
	}
	return out
}

// Fee returns the max fee that was set for `tx`
func (tx *Transaction) MaxFee() int32 {
	return int32(tx.Envelope.Tx.Fee)
}

// FeeCharged returns the fee that was actually charged for `tx`
func (tx *Transaction) FeeCharged() int32 {
	return int32(tx.Result.Result.FeeCharged)
}

// IsSuccessful returns true when the transaction was successful.
func (tx *Transaction) IsSuccessful() bool {
	return tx.Result.Result.Result.Code == xdr.TransactionResultCodeTxSuccess
}

// Memo returns the memo for this transaction, if there is one.
func (tx *Transaction) Memo() null.String {
	var (
		value string
		valid bool
	)
	switch tx.Envelope.Tx.Memo.Type {
	case xdr.MemoTypeMemoNone:
		value, valid = "", false
	case xdr.MemoTypeMemoText:
		scrubbed := utf8.Scrub(tx.Envelope.Tx.Memo.MustText())
		notnull := strings.Join(strings.Split(scrubbed, "\x00"), "")
		value, valid = notnull, true
	case xdr.MemoTypeMemoId:
		value, valid = fmt.Sprintf("%d", tx.Envelope.Tx.Memo.MustId()), true
	case xdr.MemoTypeMemoHash:
		hash := tx.Envelope.Tx.Memo.MustHash()
		value, valid =
			base64.StdEncoding.EncodeToString(hash[:]),
			true
	case xdr.MemoTypeMemoReturn:
		hash := tx.Envelope.Tx.Memo.MustRetHash()
		value, valid =
			base64.StdEncoding.EncodeToString(hash[:]),
			true
	default:
		panic(fmt.Errorf("invalid memo type: %v", tx.Envelope.Tx.Memo.Type))
	}

	return null.NewString(value, valid)
}

// MemoType returns the memo type for this transaction
func (tx *Transaction) MemoType() string {
	switch tx.Envelope.Tx.Memo.Type {
	case xdr.MemoTypeMemoNone:
		return "none"
	case xdr.MemoTypeMemoText:
		return "text"
	case xdr.MemoTypeMemoId:
		return "id"
	case xdr.MemoTypeMemoHash:
		return "hash"
	case xdr.MemoTypeMemoReturn:
		return "return"
	default:
		panic(fmt.Errorf("invalid memo type: %v", tx.Envelope.Tx.Memo.Type))
	}
}

// ResultXDR returns the XDR encoded result for this transaction
func (tx *Transaction) ResultXDR() string {
	out, err := xdr.MarshalBase64(tx.Result.Result)
	if err != nil {
		panic(err)
	}
	return out
}

// ResultMetaXDR returns the XDR encoded result meta for this transaction
func (tx *Transaction) ResultMetaXDR() string {
	out, err := xdr.MarshalBase64(tx.ResultMeta)
	if err != nil {
		panic(err)
	}
	return out
}

// Sequence returns the sequence number for `tx`
func (tx *Transaction) Sequence() int64 {
	return int64(tx.Envelope.Tx.SeqNum)
}

// SourceAddress returns the strkey-encoded account id that paid the fee for
// `tx`.
func (tx *Transaction) SourceAddress() string {
	sa := tx.Envelope.Tx.SourceAccount
	pubkey := sa.MustEd25519()
	raw := make([]byte, 32)
	copy(raw, pubkey[:])
	return strkey.MustEncode(strkey.VersionByteAccountID, raw)
}

// TransactionByHashAfterLedger is a query that loads a single row from the `txhistory`.
func (q *Q) TransactionByHashAfterLedger(
	dest interface{},
	hash string,
	ledger int32,
) error {
	sql := sq.Select("ctxh.*").
		From("txhistory ctxh").
		Limit(1).
		Where("ctxh.txid = ?", hash).
		Where("ctxh.ledgerseq > ?", ledger)

	return q.Get(dest, sql)
}

// TransactionsByLedger is a query that loads all rows from `txhistory` where
// ledgerseq matches `Sequence.`
func (q *Q) TransactionsByLedger(dest interface{}, seq int32) error {
	sql := sq.Select("ctxh.*").
		From("txhistory ctxh").
		OrderBy("ctxh.txindex ASC").
		Where("ctxh.ledgerseq = ?", seq)

	return q.Select(dest, sql)
}
