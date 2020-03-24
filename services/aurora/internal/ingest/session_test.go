package ingest

import (
	"testing"

	protocolEffects "github.com/hcnet/go/protocols/aurora/effects"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func Test_ingestSignerEffects(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("set_options")
	defer tt.Finish()

	s := ingest(tt, Config{EnableAssetStats: false})
	tt.Require.NoError(s.Err)

	q := &history.Q{Session: tt.AuroraSession()}

	// Regression: https://github.com/hcnet/aurora/issues/390 doesn't produce a signer effect when
	// inflation has changed
	var effects []history.Effect
	err := q.Effects().ForLedger(3).Select(&effects)
	tt.Require.NoError(err)

	if tt.Assert.Len(effects, 1) {
		tt.Assert.NotEqual(history.EffectSignerUpdated, effects[0].Type)
	}
}

func Test_ingestOperationEffects(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("set_options")
	defer tt.Finish()

	s := ingest(tt, Config{EnableAssetStats: false})
	tt.Require.NoError(s.Err)

	q := &history.Q{Session: tt.AuroraSession()}
	var effects []history.Effect

	// ensure inflation destination change is correctly recorded
	err := q.Effects().ForLedger(3).Select(&effects)
	tt.Require.NoError(err)

	if tt.Assert.Len(effects, 1) {
		tt.Assert.Equal(history.EffectAccountInflationDestinationUpdated, effects[0].Type)
	}

	// HACK(scott): switch to kahuna recipe mid-stream.  We need to integrate our test scenario loader to be compatible with go subtests/
	tt.ScenarioWithoutAurora("kahuna")
	s = ingest(tt, Config{EnableAssetStats: false})
	tt.Require.NoError(s.Err)
	pq, err := db2.NewPageQuery("", true, "asc", 200)
	tt.Require.NoError(err)

	// ensure payments get the payment effects
	err = q.Effects().ForLedger(15).Page(pq).Select(&effects)
	tt.Require.NoError(err)

	if tt.Assert.Len(effects, 2) {
		tt.Assert.Equal(history.EffectAccountCredited, effects[0].Type)
		tt.Assert.Equal(history.EffectAccountDebited, effects[1].Type)
	}

	// ensure path payments get the payment effects
	err = q.Effects().ForLedger(20).Page(pq).Select(&effects)
	tt.Require.NoError(err)

	if tt.Assert.Len(effects, 4) {
		tt.Assert.Equal(history.EffectAccountCredited, effects[0].Type)
		tt.Assert.Equal(history.EffectAccountDebited, effects[1].Type)
		tt.Assert.Equal(history.EffectTrade, effects[2].Type)
		tt.Assert.Equal(history.EffectTrade, effects[3].Type)
	}

	err = q.Effects().ForOperation(81604382721).Page(pq).Select(&effects)
	tt.Require.NoError(err)

	var ad protocolEffects.AccountDebited
	err = effects[1].UnmarshalDetails(&ad)
	tt.Require.NoError(err)
	tt.Assert.Equal("100.0000000", ad.Amount)
}

func Test_ingestBumpSeq(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()

	s := ingest(tt, Config{EnableAssetStats: false})
	tt.Require.NoError(s.Err)

	q := &history.Q{Session: tt.AuroraSession()}

	//ensure bumpseq operations
	ops, _, err := q.Operations().ForAccount("GCQZP3IU7XU6EJ63JZXKCQOYT2RNXN3HB5CNHENNUEUHSMA4VUJJJSEN").Fetch()
	tt.Require.NoError(err)
	if tt.Assert.Len(ops, 5) {
		//first is create account, and then bump sequences
		tt.Assert.Equal(xdr.OperationTypeCreateAccount, ops[0].Type)
		for i := 1; i < 5; i++ {
			tt.Assert.Equal(xdr.OperationTypeBumpSequence, ops[i].Type)
		}
	}

	//ensure bumpseq effect
	var effects []history.Effect
	err = q.Effects().OfType(history.EffectSequenceBumped).Select(&effects)
	tt.Require.NoError(err)

	//sample a bumpseq effect
	if tt.Assert.Len(effects, 1) {
		testEffect := effects[0]
		details := struct {
			NewSq int64 `json:"new_seq"`
		}{}
		err = testEffect.UnmarshalDetails(&details)
		tt.Assert.Equal(int64(300000000000), details.NewSq)
	}
}
