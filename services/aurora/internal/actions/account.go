package actions

import (
	"context"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
)

// AccountInfo returns the information about an account identified by addr.
func AccountInfo(ctx context.Context, cq *core.Q, addr string) (*protocol.Account, error) {
	var (
		coreRecord     core.Account
		coreData       []core.AccountData
		coreSigners    []core.Signer
		coreTrustlines []core.Trustline
		resource       protocol.Account
	)

	err := cq.AccountByAddress(&coreRecord, addr)
	if err != nil {
		return nil, errors.Wrap(err, "getting core account record")
	}

	err = cq.AllDataByAddress(&coreData, addr)
	if err != nil {
		return nil, errors.Wrap(err, "getting core account data")
	}

	err = cq.SignersByAddress(&coreSigners, addr)
	if err != nil {
		return nil, errors.Wrap(err, "getting core signer")
	}

	err = cq.TrustlinesByAddress(&coreTrustlines, addr)
	if err != nil {
		return nil, errors.Wrap(err, "getting core trustline")
	}

	err = resourceadapter.PopulateAccount(
		ctx,
		&resource,
		coreRecord,
		coreData,
		coreSigners,
		coreTrustlines,
	)

	return &resource, errors.Wrap(err, "populating account")
}

// AccountPage returns a page containing the account records that
// have `signer` as a signer.
// This doesn't return full account details resource because of the
// limitations of existing ingestion architecture. In a future, when
// the new ingestion system is fully integrated, this endpoint can be
// used to find accounts for signer but also accounts for assets,
// home domain, inflation_dest etc.
func AccountPage(ctx context.Context, hq history.QSigners, signer string, pq db2.PageQuery) (hal.Page, error) {
	records, err := hq.AccountsForSigner(signer, pq)
	if err != nil {
		return hal.Page{}, errors.Wrap(err, "loading account records")
	}

	page := hal.Page{
		Cursor: pq.Cursor,
		Order:  pq.Order,
		Limit:  pq.Limit,
	}

	for _, record := range records {
		var res protocol.AccountSigner
		resourceadapter.PopulateAccountSigner(ctx, &res, record)
		page.Add(res)
	}

	page.FullURL = fullURL(ctx)
	page.PopulateLinks()
	return page, nil
}
