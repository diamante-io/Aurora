package ticker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	auroraclient "github.com/hcnet/go/clients/auroraclient"
	hProtocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/ticker/internal/scraper"
	"github.com/hcnet/go/services/ticker/internal/tickerdb"
	hlog "github.com/hcnet/go/support/log"
)

// StreamTrades constantly streams and ingests new trades directly from aurora.
func StreamTrades(
	ctx context.Context,
	s *tickerdb.TickerSession,
	c *auroraclient.Client,
	l *hlog.Entry,
) error {
	sc := scraper.ScraperConfig{
		Client: c,
		Logger: l,
		Ctx:    &ctx,
	}
	handler := func(trade hProtocol.Trade) {
		l.Infof("New trade arrived. ID: %v; Close Time: %v\n", trade.ID, trade.LedgerCloseTime)
		scraper.NormalizeTradeAssets(&trade)
		bID, cID, err := findBaseAndCounter(s, trade)
		if err != nil {
			l.Errorln(err)
			return
		}
		dbTrade, err := hProtocolTradeToDBTrade(trade, bID, cID)
		if err != nil {
			l.Errorln(err)
			return
		}

		err = s.BulkInsertTrades([]tickerdb.Trade{dbTrade})
		if err != nil {
			l.Errorln("Could not insert trade in database: ", trade.ID)
		}
	}

	// Ensure we start streaming from the last stored trade
	lastTrade, err := s.GetLastTrade()
	if err != nil {
		return err
	}

	cursor := lastTrade.AuroraID
	return sc.StreamNewTrades(cursor, handler)
}

// BackfillTrades ingest the most recent trades (limited to numDays) directly from Aurora
// into the database.
func BackfillTrades(
	s *tickerdb.TickerSession,
	c *auroraclient.Client,
	l *hlog.Entry,
	numHours int,
	limit int,
) error {
	sc := scraper.ScraperConfig{
		Client: c,
		Logger: l,
	}
	now := time.Now()
	since := now.Add(time.Hour * -time.Duration(numHours))
	trades, err := sc.FetchAllTrades(since, limit)
	if err != nil {
		return err
	}

	var dbTrades []tickerdb.Trade

	for _, trade := range trades {
		bID, cID, err := findBaseAndCounter(s, trade)
		if err != nil {
			continue
		}

		dbTrade, err := hProtocolTradeToDBTrade(trade, bID, cID)
		if err != nil {
			l.Errorln("Could not convert entry to DB Trade: ", err)
			continue
		}
		dbTrades = append(dbTrades, dbTrade)
	}

	l.Infof("Inserting %d entries in the database.\n", len(dbTrades))
	err = s.BulkInsertTrades(dbTrades)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

// findBaseAndCounter tries to find the Base and Counter assets IDs in the database,
// and returns an error if it doesn't find any.
func findBaseAndCounter(s *tickerdb.TickerSession, trade hProtocol.Trade) (bID int32, cID int32, err error) {
	bFound, bID, err := s.GetAssetByCodeAndIssuerAccount(
		trade.BaseAssetCode,
		trade.BaseAssetIssuer,
	)
	if err != nil {
		return
	}

	cFound, cID, err := s.GetAssetByCodeAndIssuerAccount(
		trade.CounterAssetCode,
		trade.CounterAssetIssuer,
	)
	if err != nil {
		return
	}

	if !bFound || !cFound {
		err = errors.New("base or counter asset no found")
		return
	}

	return
}

// hProtocolTradeToDBTrade converts from a hProtocol.Trade to a tickerdb.Trade
func hProtocolTradeToDBTrade(
	hpt hProtocol.Trade,
	baseAssetID int32,
	counterAssetID int32,
) (trade tickerdb.Trade, err error) {
	fBaseAmount, err := strconv.ParseFloat(hpt.BaseAmount, 64)
	if err != nil {
		return
	}
	fCounterAmount, err := strconv.ParseFloat(hpt.CounterAmount, 64)
	if err != nil {
		return
	}

	fPrice := float64(hpt.Price.D) / float64(hpt.Price.N)

	trade = tickerdb.Trade{
		AuroraID:       hpt.ID,
		LedgerCloseTime: hpt.LedgerCloseTime,
		OfferID:         hpt.OfferID,
		BaseOfferID:     hpt.BaseOfferID,
		BaseAccount:     hpt.BaseAccount,
		BaseAmount:      fBaseAmount,
		BaseAssetID:     baseAssetID,
		CounterOfferID:  hpt.CounterOfferID,
		CounterAccount:  hpt.CounterAccount,
		CounterAmount:   fCounterAmount,
		CounterAssetID:  counterAssetID,
		BaseIsSeller:    hpt.BaseIsSeller,
		Price:           fPrice,
	}

	return
}
