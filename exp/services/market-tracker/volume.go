package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	hProtocol "github.com/diamnet/go/protocols/aurora"
)

// Volume stores volume of a various base pair in both XLM and USD.
// It also includes metadata about associated trades.
type Volume struct {
	BaseVolumeBaseAsset    prometheus.Gauge
	BaseVolumeUsd          prometheus.Gauge
	CounterVolumeBaseAsset prometheus.Gauge
	CounterVolumeUsd       prometheus.Gauge
	TradeCount             prometheus.Gauge
	TradeAvgAmt            prometheus.Gauge
}

type xlmPrice struct {
	timestamp int64
	price     float64
}

type volumeHist struct {
	start                  int64
	end                    int64
	numTrades              float64
	baseVolumeBaseAsset    float64
	baseVolumeUsd          float64
	counterVolumeBaseAsset float64
	counterVolumeUsd       float64
}

func trackVolumes(cfg Config, c trackerClient, watchedTPsPtr *[]prometheusWatchedTP) {
	watchedTPs := *watchedTPsPtr
	volumeMap := initVolumes(cfg, c, watchedTPs)

	go func() {
		updateVolume(cfg, c, watchedTPsPtr, volumeMap)
	}()
}

func initVolumes(cfg Config, c trackerClient, watchedTPs []prometheusWatchedTP) map[string][]volumeHist {
	xlmReq := mustCreateXlmPriceRequest()
	xlmPriceHist, err := getXlmPriceHistory(xlmReq)
	if err != nil {
		fmt.Printf("got error when getting xlm price history: %s\n", err)
	}

	volumeHistMap := make(map[string][]volumeHist)
	end := time.Now()
	start := end.Add(time.Duration(-24 * time.Hour))
	res := 15 * 60                                     // resolution length, in seconds
	cRes := time.Duration(res*1000) * time.Millisecond // aurora request must be in milliseconds

	for i, wtp := range watchedTPs {
		// TODO: Calculate volume for assets with non-native counter.
		if wtp.TradePair.SellingAsset.Code != "XLM" && wtp.TradePair.SellingAsset.IssuerAddress != "" {
			continue
		}

		currency := watchedTPs[i].TradePair.BuyingAsset.Currency
		trueAssetUsdPrice, err := updateAssetUsdPrice(currency)
		if err != nil {
			fmt.Printf("error while getting asset price: %s\n", err)
			return make(map[string][]volumeHist)
		}

		taps, err := c.getAggTradesForTradePair(wtp.TradePair, start, end, cRes)
		if err != nil {
			fmt.Printf("got error getting agg trades for pair %s\n: %s", wtp.TradePair.String(), err)
		}

		records := getAggRecords(taps)
		volumeHist, err := constructVolumeHistory(records, xlmPriceHist, trueAssetUsdPrice, start, end, res)
		if err != nil {
			fmt.Printf("got error constructing volume history for pair %s\n: %s", wtp.TradePair.String(), err)
		}

		volumeHistMap[wtp.TradePair.String()] = volumeHist

		day := end.Add(time.Duration(-24 * time.Hour)).Unix()
		watchedTPs[i].Volume.BaseVolumeBaseAsset.Set(addBaseVolumeBaseAssetHistory(volumeHist, day))
		watchedTPs[i].Volume.BaseVolumeUsd.Set(addBaseVolumeUsdHistory(volumeHist, day))
		watchedTPs[i].Volume.CounterVolumeBaseAsset.Set(addCounterVolumeBaseHistory(volumeHist, day))
		watchedTPs[i].Volume.CounterVolumeUsd.Set(addCounterVolumeUsdHistory(volumeHist, day))
		watchedTPs[i].Volume.TradeCount.Set(addTradeCount(volumeHist, day))
		watchedTPs[i].Volume.TradeAvgAmt.Set(addTradeAvg(volumeHist, day))
	}
	return volumeHistMap
}

func updateVolume(cfg Config, c trackerClient, watchedTPsPtr *[]prometheusWatchedTP, volumeHistMap map[string][]volumeHist) {
	req := mustCreateXlmPriceRequest()
	historyUnit := time.Duration(15 * 60 * time.Second) // length of each individual unit of volume history
	cRes := time.Duration(60*1000) * time.Millisecond   // aurora client requests have a 1 minute resolution, in milliseconds
	day := time.Duration(24 * 60 * 60 * time.Second)    // number of seconds in a day
	watchedTPs := *watchedTPsPtr
	forLoopDuration := time.Duration(0)

	priceCache := createPriceCache(watchedTPs)
	for {
		time.Sleep(historyUnit - forLoopDuration) // wait before starting the update

		xlmUsdPrice, err := getLatestXlmPrice(req)
		if err != nil {
			fmt.Printf("error while getting latest price: %s", err)
		}

		end := time.Now()
		start := end.Add(-1 * historyUnit)
		for i, wtp := range watchedTPs {
			// TODO: Calculate volume for assets with non-native counter.
			if wtp.TradePair.SellingAsset.Code != "XLM" && wtp.TradePair.SellingAsset.IssuerAddress != "" {
				continue
			}

			trueAssetUsdPrice := 0.0
			currency := wtp.TradePair.BuyingAsset.Currency
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

			tps := wtp.TradePair.String()
			taps, err := c.getAggTradesForTradePair(wtp.TradePair, start, end, cRes)
			if err != nil {
				fmt.Printf("got error getting agg trades for pair %s\n: %s", tps, err)
				return
			}

			records := getAggRecords(taps)
			sts := start.Unix()
			ets := end.Unix()
			counterVolume, err := totalRecordsCounterVolume(records, start, end)
			if err != nil {
				fmt.Printf("got error aggregating xlm volume for pair %s\n: %s", tps, err)
			}

			baseVolume, err := totalRecordsBaseVolume(records, start, end)
			if err != nil {
				fmt.Printf("got error aggregating base volume for pair %s\n: %s", tps, err)
			}

			numTrades, err := totalRecordsTradeCount(records, start, end)
			if err != nil {
				fmt.Printf("got error aggregating trade counts for pair %s\n: %s", tps, err)
			}

			latestVolume := volumeHist{
				start:                  sts,
				end:                    ets,
				baseVolumeBaseAsset:    baseVolume,
				baseVolumeUsd:          baseVolume / trueAssetUsdPrice,
				counterVolumeBaseAsset: counterVolume * xlmUsdPrice * trueAssetUsdPrice,
				counterVolumeUsd:       counterVolume * xlmUsdPrice,
				numTrades:              numTrades,
			}

			// get the volumes of the oldest history unit, for both the day and month
			vh := volumeHistMap[tps]
			oldestVolume := vh[int(day/historyUnit)]

			// remove the oldest volume, store the newest one
			vh = vh[:len(vh)-1]
			vh = append([]volumeHist{latestVolume}, vh...)
			volumeHistMap[tps] = vh

			// update the volume metrics using the difference between the latest and oldest
			// history units' volumes, as appropriate for that metric
			watchedTPs[i].Volume.BaseVolumeBaseAsset.Add(latestVolume.baseVolumeBaseAsset - oldestVolume.baseVolumeBaseAsset)
			watchedTPs[i].Volume.BaseVolumeUsd.Add(latestVolume.baseVolumeUsd - oldestVolume.baseVolumeUsd)
			watchedTPs[i].Volume.CounterVolumeBaseAsset.Add(latestVolume.counterVolumeBaseAsset - oldestVolume.counterVolumeBaseAsset)
			watchedTPs[i].Volume.CounterVolumeUsd.Add(latestVolume.counterVolumeUsd - oldestVolume.counterVolumeUsd)
			watchedTPs[i].Volume.TradeCount.Add(latestVolume.numTrades - oldestVolume.numTrades)
			watchedTPs[i].Volume.TradeAvgAmt.Add(latestVolume.counterVolumeUsd/latestVolume.numTrades - oldestVolume.counterVolumeUsd/oldestVolume.numTrades)
		}

		forLoopDuration = time.Now().Sub(end)
	}
}

func getAggRecords(taps []hProtocol.TradeAggregationsPage) (records []hProtocol.TradeAggregation) {
	for _, tap := range taps {
		records = append(records, tap.Embedded.Records...)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Timestamp > records[j].Timestamp })
	return
}

func constructVolumeHistory(tas []hProtocol.TradeAggregation, xlmPrices []xlmPrice, assetPrice float64, start, end time.Time, res int) ([]volumeHist, error) {
	if len(xlmPrices) < 2 {
		return []volumeHist{}, fmt.Errorf("mis-formed xlm price history from diamnet expert")
	}

	volumeHistory := []volumeHist{}
	priceIdx := -1
	recordIdx := 0
	currEnd := end
	for currEnd.After(start) {
		// find the weighted price for the current interval
		cets := currEnd.Unix()
		csts := cets - int64(res)
		priceIdx = findTimestampPriceIndex(csts, xlmPrices, priceIdx)

		weightedXlmUsdPrice, err := calcWeightedPrice(csts, priceIdx, xlmPrices)
		if err != nil {
			return []volumeHist{}, err
		}

		// find total volume of records in this interval
		// TODO: This loop does not correctly include records before the start
		// time. however, that should not happen, given that we define start before
		// calling the aurora client.
		currBaseVolume := 0.0
		currCounterVolume := 0.0
		currTradeCount := 0.0
		for recordIdx < len(tas) {
			r := tas[recordIdx]
			rts := r.Timestamp / 1000
			if rts < csts {
				// if record is before timeframe, break
				break
			} else if rts > cets {
				// if record is after timeframe, continue to next
				// record, since this could be in range
				recordIdx++
				continue
			} else {
				recordBaseVolume, err := strconv.ParseFloat(r.BaseVolume, 64)
				if err != nil {
					return []volumeHist{}, err
				}

				recordCounterVolume, err := strconv.ParseFloat(r.CounterVolume, 64)
				if err != nil {
					return []volumeHist{}, err
				}

				currBaseVolume += recordBaseVolume
				currCounterVolume += recordCounterVolume
				currTradeCount += float64(r.TradeCount)
				recordIdx++
			}
		}

		currVolume := volumeHist{
			start:                  csts,
			end:                    cets,
			numTrades:              currTradeCount,
			baseVolumeBaseAsset:    currBaseVolume,
			baseVolumeUsd:          currBaseVolume / assetPrice,
			counterVolumeBaseAsset: currCounterVolume * weightedXlmUsdPrice * assetPrice,
			counterVolumeUsd:       weightedXlmUsdPrice * currCounterVolume,
		}

		currEnd = currEnd.Add(time.Duration(-1*res) * time.Second)
		volumeHistory = append(volumeHistory, currVolume)
	}
	return volumeHistory, nil
}

func addBaseVolumeBaseAssetHistory(history []volumeHist, end int64) (baseVolume float64) {
	for _, vh := range history {
		if vh.end < end {
			break
		}
		baseVolume += vh.baseVolumeBaseAsset
	}
	return
}

func addBaseVolumeUsdHistory(history []volumeHist, end int64) (baseVolume float64) {
	for _, vh := range history {
		if vh.end < end {
			break
		}
		baseVolume += vh.baseVolumeUsd
	}
	return
}

func addCounterVolumeBaseHistory(history []volumeHist, end int64) (counterVolume float64) {
	for _, vh := range history {
		if vh.end < end {
			break
		}
		counterVolume += vh.counterVolumeBaseAsset
	}
	return
}

func addCounterVolumeUsdHistory(history []volumeHist, end int64) (counterVolume float64) {
	for _, vh := range history {
		if vh.end < end {
			break
		}
		counterVolume += vh.counterVolumeUsd
	}
	return
}

func addTradeCount(history []volumeHist, end int64) (tradeCount float64) {
	for _, vh := range history {
		if vh.end < end {
			break
		}
		tradeCount += float64(vh.numTrades)
	}
	return
}

func addTradeAvg(history []volumeHist, end int64) (tradeAvg float64) {
	totalAmt := 0.
	totalTrades := 0.
	for _, vh := range history {
		if vh.end < end {
			break
		}
		totalAmt += vh.counterVolumeUsd
		totalTrades += float64(vh.numTrades)
	}
	tradeAvg = totalAmt / totalTrades
	return
}

func totalRecordsCounterVolume(tas []hProtocol.TradeAggregation, start, end time.Time) (float64, error) {
	tv := 0.0
	for _, ta := range tas {
		ts := time.Unix(ta.Timestamp/1000, 0) // timestamps are milliseconds since epoch time
		if ts.Before(start) || ts.After(end) {
			return 0.0, fmt.Errorf("record at timestamp %v is out of time bounds %v to %v", ts, start, end)
		}

		cv, err := strconv.ParseFloat(ta.CounterVolume, 64)
		if err != nil {
			return 0.0, err
		}

		tv += cv
	}
	return tv, nil
}

func totalRecordsBaseVolume(tas []hProtocol.TradeAggregation, start, end time.Time) (float64, error) {
	tv := 0.0
	for _, ta := range tas {
		ts := time.Unix(ta.Timestamp/1000, 0) // timestamps are milliseconds since epoch time
		if ts.Before(start) || ts.After(end) {
			return 0.0, fmt.Errorf("record at timestamp %v is out of time bounds %v to %v", ts, start, end)
		}

		bv, err := strconv.ParseFloat(ta.BaseVolume, 64)
		if err != nil {
			return 0.0, err
		}

		tv += bv
	}
	return tv, nil
}

func totalRecordsTradeCount(tas []hProtocol.TradeAggregation, start, end time.Time) (float64, error) {
	ttc := 0.0
	for _, ta := range tas {
		ts := time.Unix(ta.Timestamp/1000, 0) // timestamps are milliseconds since epoch time
		if ts.Before(start) || ts.After(end) {
			return 0, fmt.Errorf("record at timestamp %v is out of time bounds %v to %v", ts, start, end)
		}

		ttc += float64(ta.TradeCount)
	}
	return ttc, nil
}

// findTimestampPriceIndex iterates through an array of timestamps and prices, and returns the
// index of the oldest such pair that is more recent than the given timestamp.
// This assumes those pairs are sorted by decreasing timestamp, (i.e. most recent first).
func findTimestampPriceIndex(timestamp int64, prices []xlmPrice, startIndex int) int {
	index := startIndex
	if index < 0 {
		if timestamp > prices[0].timestamp {
			return index
		}
		index = 0
	}

	for index < len(prices)-1 {
		if prices[index].timestamp >= timestamp && timestamp >= prices[index+1].timestamp {
			break
		}
		index++
	}
	return index
}

func calcWeightedPrice(timestamp int64, startIndex int, prices []xlmPrice) (float64, error) {
	// we expect prices sorted in decreasing timestamp (i.e., most recent first)
	// TODO: Use resolution to weight prices.
	if startIndex < 0 {
		if timestamp < prices[0].timestamp {
			return 0.0, fmt.Errorf("update price index before calculating price")
		}
		return prices[0].price, nil
	} else if startIndex >= len(prices)-1 {
		if timestamp > prices[len(prices)-1].timestamp {
			return 0.0, fmt.Errorf("update price index before calculating price")
		}
		return prices[len(prices)-1].price, nil
	}

	if timestamp > prices[startIndex].timestamp || timestamp < prices[startIndex+1].timestamp {
		return 0.0, fmt.Errorf("update price index before calculating price")
	}

	avgPrice := (prices[startIndex].price + prices[startIndex+1].price) / 2
	return avgPrice, nil
}
