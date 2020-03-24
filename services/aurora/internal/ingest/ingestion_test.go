package ingest

import (
	"database/sql"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/db2/sqx"
	"github.com/hcnet/go/services/aurora/internal/test"
	testDB "github.com/hcnet/go/services/aurora/internal/test/db"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestEmptySignature(t *testing.T) {
	ingestion := Ingestion{
		DB: &db.Session{
			DB: testDB.Aurora(t),
		},
	}
	ingestion.Start()

	envelope := xdr.TransactionEnvelope{}
	resultPair := xdr.TransactionResultPair{}
	meta := xdr.TransactionMeta{}

	xdr.SafeUnmarshalBase64("AAAAAMIK9djC7k75ziKOLJcvMAIBG7tnBuoeI34x+Pi6zqcZAAAAZAAZphYAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAAynnCTTyw53VVRLOWX6XKTva63IM1LslPNW01YB0hz/8AAAAAAAAAAlQL5AAAAAAAAAAAAh0hz/8AAABA8qkkeKaKfsbgInyIkzXJhqJE5/Ufxri2LdxmyKkgkT6I3sPmvrs5cPWQSzEQyhV750IW2ds97xTHqTpOfuZCAnhSuFUAAAAA", &envelope)
	xdr.SafeUnmarshalBase64("AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=", &resultPair.Result)
	xdr.SafeUnmarshalBase64("AAAAAAAAAAEAAAADAAAAAQAZphoAAAAAAAAAAMIK9djC7k75ziKOLJcvMAIBG7tnBuoeI34x+Pi6zqcZAAAAF0h255wAGaYWAAAAAQAAAAMAAAAAAAAAAAAAAAADBQUFAAAAAwAAAAAtkqVYLPLYhqNMmQLPc+T9eTWp8LIE8eFlR5K4wNJKTQAAAAMAAAAAynnCTTyw53VVRLOWX6XKTva63IM1LslPNW01YB0hz/8AAAADAAAAAuOwxEKY/BwUmvv0yJlvuSQnrkHkZJuTTKSVmRt4UrhVAAAAAwAAAAAAAAAAAAAAAwAZphYAAAAAAAAAAMp5wk08sOd1VUSzll+lyk72utyDNS7JTzVtNWAdIc//AAAAF0h26AAAGaYWAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAQAZphoAAAAAAAAAAMp5wk08sOd1VUSzll+lyk72utyDNS7JTzVtNWAdIc//AAAAGZyCzAAAGaYWAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA", &meta)

	transaction := &core.Transaction{
		TransactionHash: "1939a8de30981e4171e1aaeca54a058a7fb06684864facba0620ab8cc5076d4f",
		LedgerSequence:  1680922,
		Index:           1,
		Envelope:        envelope,
		Result:          resultPair,
		ResultMeta:      meta,
	}

	transactionFee := &core.TransactionFee{}

	ingestion.Transaction(true, 1, transaction, transactionFee)
	assert.Equal(t, 1, len(ingestion.builders[TransactionsTableName].rows))

	err := ingestion.Flush()
	assert.NoError(t, err)

	err = ingestion.Close()
	assert.NoError(t, err)
}

func TestFeeMax(t *testing.T) {
	tt := test.Start(t).Scenario("kahuna")
	defer tt.Finish()

	ingestion := Ingestion{
		DB: &db.Session{
			DB: testDB.Aurora(t),
		},
	}
	ingestion.Start()

	envelope := xdr.TransactionEnvelope{}
	resultPair := xdr.TransactionResultPair{}
	meta := xdr.TransactionMeta{}

	xdr.SafeUnmarshalBase64("AAAAAOLWsdzzeqJ5N2DeHgvzMc/mwBhceAIKwLHfM5J8zsK6AA9CQAAI+n0AAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAA4tax3PN6onk3YN4eC/Mxz+bAGFx4AgrAsd8zknzOwroAAAAAAAAAAACYloAAAAAAAAAAAXzOwroAAABAJ6nzO0f3/izUO3nB+CfN1sII66VBwGyIb6rP8VidFiYSxgY9fBviXA4FtPt0p1msOOI8NNA0alMQ95E2HOn8Dg==", &envelope)
	xdr.SafeUnmarshalBase64("AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=", &resultPair.Result)
	xdr.SafeUnmarshalBase64("AAAAAQAAAAIAAAADAAj6gwAAAAAAAAAA4tax3PN6onk3YN4eC/Mxz+bAGFx4AgrAsd8zknzOwroAAAAXSHbnnAAI+n0AAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAj6gwAAAAAAAAAA4tax3PN6onk3YN4eC/Mxz+bAGFx4AgrAsd8zknzOwroAAAAXSHbnnAAI+n0AAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAA==", &meta)

	hash := "a48debf9245e39d78c4fcb8e010b4f41dbee10823148e89269aab74ceeefcdad"

	transaction := &core.Transaction{
		TransactionHash: hash,
		LedgerSequence:  1,
		Index:           1,
		Envelope:        envelope,
		Result:          resultPair,
		ResultMeta:      meta,
	}

	transactionFee := &core.TransactionFee{}

	ingestion.Transaction(true, 1, transaction, transactionFee)
	assert.Equal(t, 1, len(ingestion.builders[TransactionsTableName].rows))

	err := ingestion.Flush()
	assert.NoError(t, err)

	err = ingestion.Close()
	assert.NoError(t, err)

	q := history.Q{Session: ingestion.DB}
	tx := history.Transaction{}
	err = q.TransactionByHash(&tx, hash)
	tt.Require.NoError(err)
	tt.Assert.Equal(int32(1000000), tx.MaxFee)
	tt.Assert.Equal(int32(100), tx.FeeCharged)

	_, err = tt.AuroraSession().ExecRaw(
		`UPDATE history_transactions SET fee_charged = NULL WHERE transaction_hash = ?`, hash,
	)
	tt.Require.NoError(err)

	tx = history.Transaction{}
	err = q.TransactionByHash(&tx, hash)
	tt.Require.NoError(err)
	tt.Assert.Equal(int32(1000000), tx.MaxFee)
	tt.Assert.Equal(int32(1000000), tx.FeeCharged)
}

func TestTimeBoundsMaxBig(t *testing.T) {
	ingestion := Ingestion{
		DB: &db.Session{
			DB: testDB.Aurora(t),
		},
	}
	ingestion.Start()

	ingestion.builders[TransactionsTableName].Values(
		125,
		"hash",
		"123",
		0,
		"abc",
		1,
		100,
		1,
		"",
		"",
		"",
		"",
		sqx.StringArray([]string{}),
		ingestion.formatTimeBounds(&xdr.TimeBounds{
			MinTime: xdr.TimePoint(0),
			MaxTime: xdr.TimePoint(9999999999999999999),
		}),
		"id",
		"111",
		time.Now().UTC(),
		time.Now().UTC(),
		true,
	)

	err := ingestion.Flush()
	assert.NoError(t, err)
}

func TestAssetIngest(t *testing.T) {
	//ingest kahuna and sample a single expected asset output

	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: true})
	tt.Require.NoError(s.Err)
	q := history.Q{Session: s.Ingestion.DB}

	expectedAsset := history.Asset{
		ID:     4,
		Type:   "credit_alphanum4",
		Code:   "USD",
		Issuer: "GB2QIYT2IAUFMRXKLSLLPRECC6OCOGJMADSPTRK7TGNT2SFR2YGWDARD",
	}

	actualAsset := history.Asset{}
	err := q.GetAssetByID(&actualAsset, 4)
	tt.Require.NoError(err)
	tt.Assert.Equal(expectedAsset, actualAsset)
}

func TestAssetStatsIngest(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("ingest_asset_stats")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: true})
	tt.Require.NoError(s.Err)
	q := history.Q{Session: s.Ingestion.DB}

	type AssetStatResult struct {
		Type        string `db:"asset_type"`
		Code        string `db:"asset_code"`
		Issuer      string `db:"asset_issuer"`
		Amount      int64  `db:"amount"`
		NumAccounts int32  `db:"num_accounts"`
		Flags       int8   `db:"flags"`
		Toml        string `db:"toml"`
	}
	assetStats := []AssetStatResult{}
	err := q.Select(
		&assetStats,
		sq.
			Select(
				"hist.asset_type",
				"hist.asset_code",
				"hist.asset_issuer",
				"stats.amount",
				"stats.num_accounts",
				"stats.flags",
				"stats.toml",
			).
			From("history_assets hist").
			Join("asset_stats stats ON hist.id = stats.id").
			OrderBy("hist.asset_code ASC", "hist.asset_issuer ASC"),
	)
	tt.Require.NoError(err)
	tt.Assert.Equal(3, len(assetStats))

	tt.Assert.Equal(AssetStatResult{
		Type:        "credit_alphanum4",
		Code:        "BTC",
		Issuer:      "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
		Amount:      1009876000,
		NumAccounts: 1,
		Flags:       1,
		Toml:        "https://test.com/.well-known/hcnet.toml",
	}, assetStats[0])

	tt.Assert.Equal(AssetStatResult{
		Type:        "credit_alphanum4",
		Code:        "SCOT",
		Issuer:      "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
		Amount:      10000000000,
		NumAccounts: 1,
		Flags:       2,
		Toml:        "",
	}, assetStats[1])

	tt.Assert.Equal(AssetStatResult{
		Type:        "credit_alphanum4",
		Code:        "USD",
		Issuer:      "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
		Amount:      3000010434000,
		NumAccounts: 2,
		Flags:       1,
		Toml:        "https://test.com/.well-known/hcnet.toml",
	}, assetStats[2])
}

func TestAssetStatsDisabledIngest(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("ingest_asset_stats")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: false})
	tt.Require.NoError(s.Err)
	q := history.Q{Session: s.Ingestion.DB}

	type AssetStatResult struct {
		Type        string `db:"asset_type"`
		Code        string `db:"asset_code"`
		Issuer      string `db:"asset_issuer"`
		Amount      int64  `db:"amount"`
		NumAccounts int32  `db:"num_accounts"`
		Flags       int8   `db:"flags"`
		Toml        string `db:"toml"`
	}
	assetStats := []AssetStatResult{}
	err := q.Select(
		&assetStats,
		sq.
			Select(
				"hist.asset_type",
				"hist.asset_code",
				"hist.asset_issuer",
				"stats.amount",
				"stats.num_accounts",
				"stats.flags",
				"stats.toml",
			).
			From("history_assets hist").
			Join("asset_stats stats ON hist.id = stats.id").
			OrderBy("hist.asset_code ASC", "hist.asset_issuer ASC"),
	)
	tt.Require.NoError(err)
	tt.Assert.Equal(0, len(assetStats))
}

func TestIngestFailedTransactionsEnabled(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("failed_transactions")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: false, IngestFailedTransactions: true})
	tt.Require.NoError(s.Err)
	q := history.Q{Session: s.Ingestion.DB}

	tx := history.Transaction{}
	err := q.TransactionByHash(&tx, "aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf")
	tt.Require.NoError(err)
	tt.Assert.False(*tx.Successful)

	tx = history.Transaction{}
	err = q.TransactionByHash(&tx, "56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1")
	tt.Require.NoError(err)
	tt.Assert.True(*tx.Successful)
}

func TestIngestFailedTransactionsDisabled(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("failed_transactions")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: false, IngestFailedTransactions: false})
	tt.Require.NoError(s.Err)
	q := history.Q{Session: s.Ingestion.DB}

	tx := history.Transaction{}
	err := q.TransactionByHash(&tx, "aa168f12124b7c196c0adaee7c73a64d37f99428cacb59a91ff389626845e7cf")
	tt.Require.Error(err)
	tt.Require.Equal(err, sql.ErrNoRows)

	tx = history.Transaction{}
	err = q.TransactionByHash(&tx, "56e3216045d579bea40f2d35a09406de3a894ecb5be70dbda5ec9c0427a0d5a1")
	tt.Require.NoError(err)
	tt.Assert.True(*tx.Successful)
}

func TestTradeIngestTimestamp(t *testing.T) {
	//ingest trade scenario and verify that the trade timestamp
	//matches the appropriate ledger's timestamp
	tt := test.Start(t).ScenarioWithoutAurora("trades")
	defer tt.Finish()
	s := ingest(tt, Config{EnableAssetStats: false})
	q := history.Q{Session: s.Ingestion.DB}

	var ledgers []history.Ledger
	err := q.Ledgers().Select(&ledgers)
	tt.Require.NoError(err)

	var trades []history.Trade
	err = q.Trades().Select(&trades)
	tt.Require.NoError(err)

	tt.Require.Equal(trades[len(trades)-1].LedgerCloseTime, ledgers[len(ledgers)-1].ClosedAt)
}
