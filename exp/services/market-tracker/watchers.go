package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func configPrometheusWatchers(tps []TradePair) (watchedTPs []prometheusWatchedTP) {
	for _, tp := range tps {
		labels := prometheus.Labels{
			"tradePair":    fmt.Sprintf("%s", tp),
			"buyingAsset":  fmt.Sprintf("%s", tp.BuyingAsset),
			"sellingAsset": fmt.Sprintf("%s", tp.SellingAsset),
		}

		pwtp := prometheusWatchedTP{
			TradePair: tp,
			Spread:    createSpread(labels),
			Volume:    createVolume(labels),
			Slippage:  createSlippage(labels),
			Orderbook: createOrderbook(labels),
			FairValue: createFairValue(labels),
		}
		watchedTPs = append(watchedTPs, pwtp)
	}
	return
}

func createSpread(labels prometheus.Labels) Spread {
	return Spread{
		Top:  createSpreadGauge("", "", labels),
		D100: createSpreadGauge("_100", "at depth $100", labels),
		D1K:  createSpreadGauge("_1K", "at depth $1000", labels),
		D5K:  createSpreadGauge("_5K", "at depth $5000", labels),
		D25K: createSpreadGauge("_25K", "at depth $25,000", labels),
		D50K: createSpreadGauge("_50K", "at depth $50,000", labels),
	}
}

func createSpreadGauge(depthShort, depthLong string, labels prometheus.Labels) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name:        fmt.Sprintf("diamnet_market_tracker_spread%s", depthShort),
		ConstLabels: labels,
		Help:        fmt.Sprintf("Percentage market spread %s", depthLong),
	})
}

func createVolume(labels prometheus.Labels) Volume {
	return Volume{
		BaseVolumeBaseAsset: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_volume_base_base",
			ConstLabels: labels,
			Help:        "Base asset trading volume, in base asset, over last 1d",
		}),
		BaseVolumeUsd: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_volume_base_usd",
			ConstLabels: labels,
			Help:        "Base asset trading volume, in USD, over last 1d",
		}),
		CounterVolumeBaseAsset: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_volume_counter_base",
			ConstLabels: labels,
			Help:        "Counter asset trading volume, in base asset, over last 1d",
		}),
		CounterVolumeUsd: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_volume_counter_usd",
			ConstLabels: labels,
			Help:        "Counter asset trading volume, in USD, over last 1d",
		}),
		TradeCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_tradecount",
			ConstLabels: labels,
			Help:        "Number of trades over last 1d",
		}),
		TradeAvgAmt: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_tradeavg",
			ConstLabels: labels,
			Help:        "Average trade amount over last 1d",
		}),
	}
}

func createSlippage(labels prometheus.Labels) Slippage {
	return Slippage{
		BidD100: createSlippageGauge("bid", "100", labels),
		AskD100: createSlippageGauge("ask", "100", labels),
		BidD1K:  createSlippageGauge("bid", "1K", labels),
		AskD1K:  createSlippageGauge("ask", "1K", labels),
		BidD5K:  createSlippageGauge("bid", "5K", labels),
		AskD5K:  createSlippageGauge("ask", "5K", labels),
	}
}

func createSlippageGauge(orderType, depth string, labels prometheus.Labels) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name:        fmt.Sprintf("diamnet_market_tracker_%s_slippage_%s", orderType, depth),
		ConstLabels: labels,
		Help:        fmt.Sprintf("Slippage of %s at depth %s", orderType, depth),
	})
}

func createOrderbook(labels prometheus.Labels) Orderbook {
	return Orderbook{
		NumBids: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_numbids",
			ConstLabels: labels,
			Help:        "Number of bids in the orderbook",
		}),
		NumAsks: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_numasks",
			ConstLabels: labels,
			Help:        "Number of asks in the orderbook",
		}),
		BidBaseVolume: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_bidbasevol",
			ConstLabels: labels,
			Help:        "Volume of bids in the orderbook in base currency",
		}),
		BidUsdVolume: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_bidusdvol",
			ConstLabels: labels,
			Help:        "Volume of bids in the orderbook in USD",
		}),
		AskBaseVolume: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_askbasevol",
			ConstLabels: labels,
			Help:        "Volume of asks in the orderbook in base",
		}),
		AskUsdVolume: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_askusdvol",
			ConstLabels: labels,
			Help:        "Volume of asks in the orderbook in USD",
		}),
	}
}

func createFairValue(labels prometheus.Labels) FairValue {
	return FairValue{
		Percent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_fmp",
			ConstLabels: labels,
			Help:        "Pct difference of DEX value from fair market value",
		}),
		RefPrice: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_refprice",
			ConstLabels: labels,
			Help:        "Reference price of real asset (in USD)",
		}),
		DexPrice: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diamnet_market_tracker_price",
			ConstLabels: labels,
			Help:        "Mid-market price on the DEX in USD",
		}),
	}
}
