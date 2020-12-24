package resourceadapter

import (
	"context"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

func PopulateOrderBookSummary(
	ctx context.Context,
	dest *protocol.OrderBookSummary,
	selling xdr.Asset,
	buying xdr.Asset,
	row core.OrderBookSummary,
) error {

	err := PopulateAsset(ctx, &dest.Selling, selling)
	if err != nil {
		return err
	}
	err = PopulateAsset(ctx, &dest.Buying, buying)
	if err != nil {
		return err
	}

	err = populatePriceLevels(&dest.Bids, row.Bids())
	if err != nil {
		return err
	}
	err = populatePriceLevels(&dest.Asks, row.Asks())
	if err != nil {
		return err
	}

	return nil
}

func populatePriceLevels(destp *[]protocol.PriceLevel, rows []core.OrderBookSummaryPriceLevel) error {
	*destp = make([]protocol.PriceLevel, len(rows))
	dest := *destp

	for i, row := range rows {
		amount, err := row.AmountAsString()
		if err != nil {
			return errors.Wrap(err, "Error converting PriceLevel.Amount: "+row.Amount)
		}
		dest[i] = protocol.PriceLevel{
			Price:  row.PriceAsString(),
			Amount: amount,
			PriceR: protocol.Price{
				N: row.Pricen,
				D: row.Priced,
			},
		}
	}

	return nil
}
