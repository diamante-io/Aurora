package ingest

import (
	"testing"

	"github.com/hcnet/go/network"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestIngest_Kahuna1(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()

	s := ingest(tt, Config{EnableAssetStats: false})

	tt.Require.NoError(s.Err)
	tt.Assert.Equal(61, s.Ingested)

	// Test that re-importing fails
	s.Err = nil
	s.Run()
	tt.Require.Error(s.Err, "Reimport didn't fail as expected")

	// Test that re-importing fails with allowing clear succeeds
	s.Err = nil
	s.ClearExisting = true
	s.Run()
	tt.Require.NoError(s.Err, "Couldn't re-import, even with clear allowed")
}

func TestIngest_Kahuna2(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna-2")
	defer tt.Finish()

	s := ingest(tt, Config{EnableAssetStats: false})

	tt.Require.NoError(s.Err)
	tt.Assert.Equal(5, s.Ingested)

	// ensure that the onetime signer is gone
	q := core.Q{Session: tt.CoreSession()}
	var signers []core.Signer

	err := q.SignersByAddress(&signers, "GD6NTRJW5Z6NCWH4USWMNEYF77RUR2MTO6NP4KEDVJATTCUXDRO3YIFS")
	tt.Require.NoError(err)

	tt.Assert.Len(signers, 1)
}

func TestTick(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("base")
	defer tt.Finish()
	sys := sys(tt, Config{EnableAssetStats: false, CursorName: "HORIZON"})

	// ingest by tick
	s := sys.Tick()
	tt.Require.NoError(s.Err)
	tt.Require.Nil(sys.current)

	tt.UpdateLedgerState()
	s = sys.Tick()
	tt.Require.NotNil(s)
	tt.Require.NoError(s.Err)
}

func ingest(tt *test.T, c Config) *Session {
	sys := sys(tt, c)
	s := NewSession(sys)
	s.Cursor = NewCursor(1, ledger.CurrentState().CoreLatest, sys)
	s.Run()

	return s
}

func sys(tt *test.T, c Config) *System {
	return New(
		network.TestNetworkPassphrase,
		"",
		tt.CoreSession(),
		tt.AuroraSession(),
		c,
	)
}
