package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	hClient "github.com/diamnet/go/clients/auroraclient"
)

type prometheusWatchedTP struct {
	TradePair TradePair
	Spread    Spread
	Volume    Volume
	Slippage  Slippage
	Orderbook Orderbook
	FairValue FairValue
}

var watchedTradePairs []prometheusWatchedTP

func main() {
	cfg := loadConfig()
	c := trackerClient{hClient.DefaultPublicNetClient}
	watchedTPs := configPrometheusWatchers(cfg.TradePairs)
	trackSpreads(cfg, c, &watchedTPs)
	trackVolumes(cfg, c, &watchedTPs)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
