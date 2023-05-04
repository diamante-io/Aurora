package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Spread tracks the percent spread at various market depths.
type Spread struct {
	Top  prometheus.Gauge
	D100 prometheus.Gauge
	D1K  prometheus.Gauge
	D5K  prometheus.Gauge
	D25K prometheus.Gauge
	D50K prometheus.Gauge
}

// Slippage tracks the bid and ask slippages at various market depths.
type Slippage struct {
	BidD100 prometheus.Gauge
	AskD100 prometheus.Gauge
	BidD1K  prometheus.Gauge
	AskD1K  prometheus.Gauge
	BidD5K  prometheus.Gauge
	AskD5K  prometheus.Gauge
}

// FairValue tracks the reference price,
type FairValue struct {
	Percent  prometheus.Gauge
	RefPrice prometheus.Gauge
	DexPrice prometheus.Gauge
}

func trackSpreads(cfg Config, c trackerClient, watchedTPsPtr *[]prometheusWatchedTP) {
	watchedTPs := *watchedTPsPtr
	priceCache := createPriceCache(watchedTPs)
	req := mustCreateXlmPriceRequest()
	go func() {
		for {
			xlmPrice, err := getLatestXlmPrice(req)
			if err != nil {
				fmt.Printf("error while getting latest price: %s", err)
			}

			for i, wtp := range watchedTPs {
				obStats, err := c.getOrderBookForTradePair(wtp.TradePair)
				if err != nil {
					fmt.Printf("error while getting orderbook stats for asset pair %s: %s", wtp.TradePair, err)
				}

				spreadPct := calcSpreadPctForOrderBook(obStats)
				if err != nil {
					fmt.Printf("error while processing asset pair %s: %s", wtp.TradePair, err)
				}

				watchedTPs[i].Spread.Top.Set(spreadPct)

				// we only compute spreads at various depths for xlm-based pairs,
				// because our usd prices are in terms of xlm.
				if wtp.TradePair.SellingAsset.Code != "XLM" {
					continue
				}

				trueAssetUsdPrice := 0.0
				currency := watchedTPs[i].TradePair.BuyingAsset.Currency
				if priceCache[currency].updated.Before(time.Now().Add(-1 * time.Hour)) {
					trueAssetUsdPrice, err = updateAssetUsdPrice(currency)
					if err != nil {
						fmt.Printf("error while getting asset price: %s\n", err)
						return
					}

					priceCache[currency] = cachedPrice{
						price:   trueAssetUsdPrice,
						updated: time.Now(),
					}
				} else {
					trueAssetUsdPrice = priceCache[currency].price
				}

				usdBids, err := convertBids(obStats.Bids, xlmPrice, trueAssetUsdPrice)
				if err != nil {
					fmt.Printf("error while converting bids to USD: %s", err)
					continue
				}

				usdAsks, err := convertAsks(obStats.Asks, xlmPrice, trueAssetUsdPrice)
				if err != nil {
					fmt.Printf("error while converting asks to USD: %s", err)
					continue
				}

				watchedTPs[i].FairValue.DexPrice.Set(calcMidPrice(usdBids, usdAsks))

				watchedTPs[i].Orderbook.BidBaseVolume.Set(getOrdersBaseVolume(usdBids))
				watchedTPs[i].Orderbook.BidUsdVolume.Set(getOrdersUsdVolume(usdBids))
				watchedTPs[i].Orderbook.AskBaseVolume.Set(getOrdersBaseVolume(usdAsks))
				watchedTPs[i].Orderbook.AskUsdVolume.Set(getOrdersUsdVolume(usdAsks))
				watchedTPs[i].Orderbook.NumBids.Set(float64(len(usdBids)))
				watchedTPs[i].Orderbook.NumAsks.Set(float64(len(usdAsks)))

				watchedTPs[i].Spread.D100.Set(calcSpreadPctAtDepth(usdBids, usdAsks, 100.))
				watchedTPs[i].Spread.D1K.Set(calcSpreadPctAtDepth(usdBids, usdAsks, 1000.))
				watchedTPs[i].Spread.D5K.Set(calcSpreadPctAtDepth(usdBids, usdAsks, 5000.))
				watchedTPs[i].Spread.D25K.Set(calcSpreadPctAtDepth(usdBids, usdAsks, 25000.))
				watchedTPs[i].Spread.D50K.Set(calcSpreadPctAtDepth(usdBids, usdAsks, 50000.))

				watchedTPs[i].Slippage.BidD100.Set(calcSlippageAtDepth(usdBids, usdAsks, 100., true))
				watchedTPs[i].Slippage.AskD100.Set(calcSlippageAtDepth(usdBids, usdAsks, 100., false))
				watchedTPs[i].Slippage.BidD1K.Set(calcSlippageAtDepth(usdBids, usdAsks, 1000., true))
				watchedTPs[i].Slippage.AskD1K.Set(calcSlippageAtDepth(usdBids, usdAsks, 1000., false))
				watchedTPs[i].Slippage.BidD5K.Set(calcSlippageAtDepth(usdBids, usdAsks, 5000., true))
				watchedTPs[i].Slippage.AskD5K.Set(calcSlippageAtDepth(usdBids, usdAsks, 5000., false))

				watchedTPs[i].FairValue.Percent.Set(calcFairValuePct(usdBids, usdAsks, trueAssetUsdPrice))
				watchedTPs[i].FairValue.RefPrice.Set(trueAssetUsdPrice)
			}

			time.Sleep(time.Duration(cfg.CheckIntervalSeconds) * time.Second)
		}
	}()
}
