package resourceadapter

import (
	"context"
	"fmt"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/httpx"
	"github.com/diamnet/go/support/render/hal"
)

// PopulateAccountSigner fills out the resource's fields
func PopulateAccountSigner(
	ctx context.Context,
	dest *protocol.AccountSigner,
	has history.AccountSigner,
) {
	dest.ID = has.Account
	dest.AccountID = has.Account
	dest.PT = has.Account
	dest.Signer = protocol.Signer{
		Weight: has.Weight,
		Key:    has.Signer,
		Type:   protocol.MustKeyTypeFromAddress(has.Signer),
	}

	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	account := fmt.Sprintf("/accounts/%s", has.Account)
	dest.Links.Account = lb.Link(account)
}
