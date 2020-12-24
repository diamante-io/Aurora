package resourceadapter

import (
	"context"
	"fmt"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/httpx"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
)

// PopulateAccount fills out the resource's fields
func PopulateAccount(
	ctx context.Context,
	dest *protocol.Account,
	ca core.Account,
	cd []core.AccountData,
	cs []core.Signer,
	ct []core.Trustline,
) error {
	dest.ID = ca.Accountid
	dest.AccountID = ca.Accountid
	dest.Sequence = ca.Seqnum
	dest.SubentryCount = ca.Numsubentries
	dest.InflationDestination = ca.Inflationdest.String
	dest.HomeDomain = ca.HomeDomain.String
	dest.LastModifiedLedger = ca.LastModified

	PopulateAccountFlags(&dest.Flags, ca)
	PopulateAccountThresholds(&dest.Thresholds, ca)

	// populate balances
	dest.Balances = make([]protocol.Balance, len(ct)+1)
	for i, tl := range ct {
		err := PopulateBalance(&dest.Balances[i], tl)
		if err != nil {
			return errors.Wrap(err, "populating balance")
		}
	}

	// add native balance
	err := PopulateNativeBalance(&dest.Balances[len(dest.Balances)-1], ca.Balance, ca.BuyingLiabilities, ca.SellingLiabilities)
	if err != nil {
		return errors.Wrap(err, "populating native balance")
	}

	// populate data
	dest.Data = make(map[string]string)
	for _, d := range cd {
		dest.Data[d.Key] = d.Value
	}

	// populate signers
	dest.Signers = make([]protocol.Signer, len(cs)+1)
	for i, s := range cs {
		PopulateSigner(ctx, &dest.Signers[i], s)
	}

	PopulateMasterSigner(&dest.Signers[len(dest.Signers)-1], ca)

	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	self := fmt.Sprintf("/accounts/%s", ca.Accountid)
	dest.Links.Self = lb.Link(self)
	dest.Links.Transactions = lb.PagedLink(self, "transactions")
	dest.Links.Operations = lb.PagedLink(self, "operations")
	dest.Links.Payments = lb.PagedLink(self, "payments")
	dest.Links.Effects = lb.PagedLink(self, "effects")
	dest.Links.Offers = lb.PagedLink(self, "offers")
	dest.Links.Trades = lb.PagedLink(self, "trades")
	dest.Links.Data = lb.Link(self, "data/{key}")
	dest.Links.Data.PopulateTemplated()
	return nil
}
