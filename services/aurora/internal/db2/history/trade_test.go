package history_test

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2"
	. "github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func TestTradeQueries(t *testing.T) {
	tt := test.Start(t).Scenario("kahuna")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}
	var trades []Trade

	// All trades
	err := q.Trades().Page(db2.MustPageQuery("", false, "asc", 100)).Select(&trades)
	if tt.Assert.NoError(err) {
		tt.Assert.Len(trades, 4)
	}

	// Paging
	pq := db2.MustPageQuery(trades[0].PagingToken(), false, "asc", 1)
	var pt []Trade

	err = q.Trades().Page(pq).Select(&pt)
	if tt.Assert.NoError(err) {
		if tt.Assert.Len(pt, 1) {
			tt.Assert.Equal(trades[1], pt[0])
		}
	}

	// Cursor bounds checking
	pq = db2.MustPageQuery("", false, "desc", 1)
	err = q.Trades().Page(pq).Select(&pt)
	tt.Require.NoError(err)

	// test for asset pairs
	lumen, err := q.GetAssetID(xdr.MustNewNativeAsset())
	tt.Require.NoError(err)
	assetUSD, err := q.GetAssetID(xdr.MustNewCreditAsset("USD", "GB2QIYT2IAUFMRXKLSLLPRECC6OCOGJMADSPTRK7TGNT2SFR2YGWDARD"))
	tt.Require.NoError(err)
	assetEUR, err := q.GetAssetID(xdr.MustNewCreditAsset("EUR", "GAXMF43TGZHW3QN3REOUA2U5PW5BTARXGGYJ3JIFHW3YT6QRKRL3CPPU"))
	tt.Require.NoError(err)

	err = q.TradesForAssetPair(assetUSD, assetEUR).Select(&trades)
	tt.Require.NoError(err)
	tt.Assert.Len(trades, 0)

	assetUSD, err = q.GetAssetID(xdr.MustNewCreditAsset("USD", "GAXMF43TGZHW3QN3REOUA2U5PW5BTARXGGYJ3JIFHW3YT6QRKRL3CPPU"))
	tt.Require.NoError(err)

	err = q.TradesForAssetPair(lumen, assetUSD).Select(&trades)
	tt.Require.NoError(err)
	tt.Assert.Len(trades, 1)

	tt.Assert.Equal(xdr.Int64(2000000000), trades[0].BaseAmount)
	tt.Assert.Equal(xdr.Int64(1000000000), trades[0].CounterAmount)
	tt.Assert.Equal(true, trades[0].BaseIsSeller)

	// reverse assets
	err = q.TradesForAssetPair(assetUSD, lumen).Select(&trades)
	tt.Require.NoError(err)
	tt.Assert.Len(trades, 1)

	tt.Assert.Equal(xdr.Int64(1000000000), trades[0].BaseAmount)
	tt.Assert.Equal(xdr.Int64(2000000000), trades[0].CounterAmount)
	tt.Assert.Equal(false, trades[0].BaseIsSeller)
}
