package resourceadapter

import (
	"context"
	"fmt"
	"math/big"

	"github.com/diamnet/go/amount"
	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/httpx"
	"github.com/diamnet/go/support/render/hal"
)

func PopulateOffer(ctx context.Context, dest *protocol.Offer, row core.Offer, ledger *history.Ledger) {
	dest.ID = row.OfferID
	dest.PT = row.PagingToken()
	dest.Seller = row.SellerID
	dest.Amount = amount.String(row.Amount)
	dest.PriceR.N = row.Pricen
	dest.PriceR.D = row.Priced
	dest.Price = row.PriceAsString()

	row.SellingAsset.MustExtract(&dest.Selling.Type, &dest.Selling.Code, &dest.Selling.Issuer)
	row.BuyingAsset.MustExtract(&dest.Buying.Type, &dest.Buying.Code, &dest.Buying.Issuer)

	dest.LastModifiedLedger = row.Lastmodified
	if ledger != nil {
		dest.LastModifiedTime = &ledger.ClosedAt
	}
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	dest.Links.Self = lb.Linkf("/offers/%d", row.OfferID)
	dest.Links.OfferMaker = lb.Linkf("/accounts/%s", row.SellerID)
}

// PopulateHistoryOffer constructs an offer response struct from an offer row extracted from the
// the aurora offers table. Note that the only difference between PopulateHistoryOffer and PopulateOffer
// is that PopulateHistoryOffer takes an offer row from the aurora database whereas PopulateOffer
// takes an offer row from the diamnet core database. Once the experimental aurora ingestion system
// is fully rolled out there will be no need to query offers from the diamnet core database and
// we will be able to remove PopulateOffer
func PopulateHistoryOffer(ctx context.Context, dest *protocol.Offer, row history.Offer, ledger *history.Ledger) {
	dest.ID = int64(row.OfferID)
	dest.PT = fmt.Sprintf("%d", row.OfferID)
	dest.Seller = row.SellerID
	dest.Amount = amount.String(row.Amount)
	dest.PriceR.N = row.Pricen
	dest.PriceR.D = row.Priced
	dest.Price = big.NewRat(int64(row.Pricen), int64(row.Priced)).FloatString(7)

	row.SellingAsset.MustExtract(&dest.Selling.Type, &dest.Selling.Code, &dest.Selling.Issuer)
	row.BuyingAsset.MustExtract(&dest.Buying.Type, &dest.Buying.Code, &dest.Buying.Issuer)

	dest.LastModifiedLedger = int32(row.LastModifiedLedger)
	if ledger != nil {
		dest.LastModifiedTime = &ledger.ClosedAt
	}
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	dest.Links.Self = lb.Linkf("/offers/%d", row.OfferID)
	dest.Links.OfferMaker = lb.Linkf("/accounts/%s", row.SellerID)
}
