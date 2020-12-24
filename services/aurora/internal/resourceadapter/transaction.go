package resourceadapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/guregu/null"
	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/httpx"
	"github.com/diamnet/go/support/render/hal"
)

// Populate fills out the details
func PopulateTransaction(
	ctx context.Context,
	dest *protocol.Transaction,
	row history.Transaction,
) {
	dest.ID = row.TransactionHash
	dest.PT = row.PagingToken()
	// Check db2/history.Transaction.Successful field comment for more information.
	if row.Successful == nil {
		dest.Successful = true
	} else {
		dest.Successful = *row.Successful
	}
	dest.Hash = row.TransactionHash
	dest.Ledger = row.LedgerSequence
	dest.LedgerCloseTime = row.LedgerCloseTime
	dest.Account = row.Account
	dest.AccountSequence = row.AccountSequence
	dest.FeePaid = row.FeeCharged

	dest.FeeCharged = row.FeeCharged
	dest.MaxFee = row.MaxFee

	dest.OperationCount = row.OperationCount
	dest.EnvelopeXdr = row.TxEnvelope
	dest.ResultXdr = row.TxResult
	dest.ResultMetaXdr = row.TxMeta
	dest.FeeMetaXdr = row.TxFeeMeta
	dest.MemoType = row.MemoType
	dest.Memo = row.Memo.String
	dest.Signatures = strings.Split(row.SignatureString, ",")
	dest.ValidBefore = timeString(dest, row.ValidBefore)
	dest.ValidAfter = timeString(dest, row.ValidAfter)

	lb := hal.LinkBuilder{Base: httpx.BaseURL(ctx)}
	dest.Links.Account = lb.Link("/accounts", dest.Account)
	dest.Links.Ledger = lb.Link("/ledgers", fmt.Sprintf("%d", dest.Ledger))
	dest.Links.Operations = lb.PagedLink("/transactions", dest.ID, "operations")
	dest.Links.Effects = lb.PagedLink("/transactions", dest.ID, "effects")
	dest.Links.Self = lb.Link("/transactions", dest.ID)
	dest.Links.Succeeds = lb.Linkf("/transactions?order=desc&cursor=%s", dest.PT)
	dest.Links.Precedes = lb.Linkf("/transactions?order=asc&cursor=%s", dest.PT)
}

func timeString(res *protocol.Transaction, in null.Int) string {
	if !in.Valid {
		return ""
	}

	return time.Unix(in.Int64, 0).UTC().Format(time.RFC3339)
}
