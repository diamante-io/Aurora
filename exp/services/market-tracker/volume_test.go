package main

import (
	"testing"
	"time"

	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/stretchr/testify/assert"
)

var pts int64 = 1594668800
var res = 15 * 60 // 15 minutes, in seconds

var xp1 = xlmPrice{
	timestamp: pts,
	price:     1.00,
}

var xp2 = xlmPrice{
	timestamp: pts - int64(res),
	price:     2.00,
}

var xp3 = xlmPrice{
	timestamp: pts - int64(2*res),
	price:     3.00,
}

var xp4 = xlmPrice{
	timestamp: pts - int64(3*res),
	price:     4.00,
}

var prices = []xlmPrice{xp1, xp2, xp3, xp4}

func TestTotalRecordsBaseVolume(t *testing.T) {
	res := 15 * 60
	ta1 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(res/2))}
	ta1.BaseVolume = "100.0"
	ta2 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(3*res/2))}
	ta2.BaseVolume = "200.0"
	ta3 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(5*res/2))}
	ta3.BaseVolume = "300.0"
	tas := []hProtocol.TradeAggregation{ta1, ta2, ta3}

	pt := time.Unix(pts, 0)
	total, err := totalRecordsBaseVolume(tas, time.Unix(pts-int64(res*3), 0), pt)
	assert.NoError(t, err)
	assert.Equal(t, 600., total)

	total, err = totalRecordsBaseVolume(tas, time.Unix(pts-int64(res*2), 0), pt)
	assert.Error(t, err)
	assert.Equal(t, 0., total)
}

func TestTotalRecordsCounterVolume(t *testing.T) {
	res := 15 * 60
	ta1 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(res/2))}
	ta1.CounterVolume = "100.0"
	ta2 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(3*res/2))}
	ta2.CounterVolume = "200.0"
	ta3 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(5*res/2))}
	ta3.CounterVolume = "300.0"
	tas := []hProtocol.TradeAggregation{ta1, ta2, ta3}

	pt := time.Unix(pts, 0)
	total, err := totalRecordsCounterVolume(tas, time.Unix(pts-int64(res*3), 0), pt)
	assert.NoError(t, err)
	assert.Equal(t, 600., total)

	total, err = totalRecordsCounterVolume(tas, time.Unix(pts-int64(res*2), 0), pt)
	assert.Error(t, err)
	assert.Equal(t, 0., total)
}

func TestTotalRecordsTradeCount(t *testing.T) {
	res := 15 * 60
	ta1 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(res/2))}
	ta1.TradeCount = int64(100)
	ta2 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(3*res/2))}
	ta2.TradeCount = int64(200)
	ta3 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(5*res/2))}
	ta3.TradeCount = int64(300)
	tas := []hProtocol.TradeAggregation{ta1, ta2, ta3}

	pt := time.Unix(pts, 0)
	total, err := totalRecordsTradeCount(tas, time.Unix(pts-int64(res*3), 0), pt)
	assert.NoError(t, err)
	assert.Equal(t, 600., total)

	total, err = totalRecordsTradeCount(tas, time.Unix(pts-int64(res*2), 0), pt)
	assert.Error(t, err)
	assert.Equal(t, 0., total)
}

func TestAddVolumeHistory(t *testing.T) {
	// every 15 min over 30 days
	numIntervals := 4 * 24 * 30
	vh := []volumeHist{}
	i := 0
	for i < numIntervals {
		s := pts - int64(i*15*60)
		h := volumeHist{
			start:                  s,
			end:                    s - 15*60,
			baseVolumeBaseAsset:    100.0,
			baseVolumeUsd:          10.0,
			counterVolumeBaseAsset: 100.0,
			counterVolumeUsd:       10.0,
		}
		vh = append(vh, h)
		i++
	}

	// one day, in seconds
	end := pts - 24*60*60
	baseBase := addBaseVolumeBaseAssetHistory(vh, end)
	assert.Equal(t, 9600., baseBase)
	baseUsd := addBaseVolumeUsdHistory(vh, end)
	assert.Equal(t, 960., baseUsd)
	counterBase := addCounterVolumeBaseHistory(vh, end)
	assert.Equal(t, 9600., counterBase)
	counterUsd := addCounterVolumeUsdHistory(vh, end)
	assert.Equal(t, 960., counterUsd)
}

func TestConstructVolumeHistory(t *testing.T) {
	res := 15 * 60
	ta1 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(res/2))}
	ta1.CounterVolume = "200.0"
	ta1.BaseVolume = "100.0"

	ta2 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(3*res/2))}
	ta2.CounterVolume = "400.0"
	ta2.BaseVolume = "200.0"

	ta3 := hProtocol.TradeAggregation{Timestamp: 1000 * (pts - int64(5*res/2))}
	ta3.CounterVolume = "600.0"
	ta3.BaseVolume = "300.0"

	tas := []hProtocol.TradeAggregation{ta1, ta2, ta3}
	start := time.Unix(pts-24*60*60, 0)
	end := time.Unix(pts, 0)

	errPrices := []xlmPrice{}
	assetUsdPrice := 10.0
	volumeHist, err := constructVolumeHistory(tas, errPrices, assetUsdPrice, start, end, res)
	assert.Error(t, err)
	assert.Equal(t, 0, len(volumeHist))

	volumeHist, err = constructVolumeHistory(tas, prices, assetUsdPrice, start, end, res)
	assert.NoError(t, err)
	assert.Equal(t, 24*4, len(volumeHist))

	assert.Equal(t, pts-int64(res), volumeHist[0].start)
	assert.Equal(t, pts, volumeHist[0].end)
	assert.Equal(t, 100.0, volumeHist[0].baseVolumeBaseAsset)
	assert.Equal(t, 10.0, volumeHist[0].baseVolumeUsd)
	assert.Equal(t, 3000.0, volumeHist[0].counterVolumeBaseAsset)
	assert.Equal(t, 300.0, volumeHist[0].counterVolumeUsd)

	assert.Equal(t, pts-int64(2*res), volumeHist[1].start)
	assert.Equal(t, pts-int64(res), volumeHist[1].end)
	assert.Equal(t, 200.0, volumeHist[1].baseVolumeBaseAsset)
	assert.Equal(t, 20.0, volumeHist[1].baseVolumeUsd)
	assert.Equal(t, 10000.0, volumeHist[1].counterVolumeBaseAsset)
	assert.Equal(t, 1000.0, volumeHist[1].counterVolumeUsd)

	assert.Equal(t, pts-int64(3*res), volumeHist[2].start)
	assert.Equal(t, pts-int64(2*res), volumeHist[2].end)
	assert.Equal(t, 300.0, volumeHist[2].baseVolumeBaseAsset)
	assert.Equal(t, 30.0, volumeHist[2].baseVolumeUsd)
	assert.Equal(t, 21000.0, volumeHist[2].counterVolumeBaseAsset)
	assert.Equal(t, 2100.0, volumeHist[2].counterVolumeUsd)
}

func TestFindTimestampPriceIndex(t *testing.T) {
	idx := -1
	ts := pts + int64(res/2)
	newIdx := findTimestampPriceIndex(ts, prices, idx)
	assert.Equal(t, -1, newIdx)

	ts = pts - int64(res/2)
	newIdx = findTimestampPriceIndex(ts, prices, idx)
	assert.Equal(t, 0, newIdx)

	ts = pts - (int64(res) + 1)
	newIdx = findTimestampPriceIndex(ts, prices, idx)
	assert.Equal(t, 1, newIdx)

	ts = pts - (int64(2*res) + 1)
	newIdx = findTimestampPriceIndex(ts, prices, idx)
	assert.Equal(t, 2, newIdx)
}

func TestCalcWeightedPrice(t *testing.T) {
	idx := -1
	ts := pts - int64(res/2)
	wp, err := calcWeightedPrice(ts, idx, prices)
	assert.Error(t, err)

	ts = pts + int64(res/2)
	wp, err = calcWeightedPrice(ts, idx, prices)
	assert.NoError(t, err)
	assert.Equal(t, 1., wp)

	idx = 0
	wp, err = calcWeightedPrice(ts, idx, prices)
	assert.Error(t, err)

	ts = pts - int64(res/2)
	wp, err = calcWeightedPrice(ts, idx, prices)
	assert.NoError(t, err)
	assert.Equal(t, 1.5, wp)

	idx = 4
	ts = pts - int64(res/2)
	wp, err = calcWeightedPrice(ts, idx, prices)
	assert.Error(t, err)

	ts = pts - int64(res*5)
	wp, err = calcWeightedPrice(ts, idx, prices)
}
