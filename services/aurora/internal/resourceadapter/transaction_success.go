package resourceadapter

import (
	"context"

	protocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/services/aurora/internal/txsub"
	"github.com/hcnet/go/support/render/hal"
)

// Populate fills out the details
func PopulateTransactionSuccess(ctx context.Context, dest *protocol.TransactionSuccess, result txsub.Result) {
	dest.Hash = result.Hash
	dest.Ledger = result.LedgerSequence
	dest.Env = result.EnvelopeXDR
	dest.Result = result.ResultXDR
	dest.Meta = result.ResultMetaXDR

	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	dest.Links.Transaction = lb.Link("/transactions", result.Hash)
}
