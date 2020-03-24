package test

import (
	"io"

	"encoding/json"

	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/operationfeestats"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/render/hal"
)

// CoreSession returns a db.Session instance pointing at the hcnet core test database
func (t *T) CoreSession() *db.Session {
	return &db.Session{
		DB:  t.CoreDB,
		Ctx: t.Ctx,
	}
}

// Finish finishes the test, logging any accumulated aurora logs to the logs
// output
func (t *T) Finish() {
	RestoreLogger()
	// Reset cached ledger state
	ledger.SetState(ledger.State{})
	operationfeestats.ResetState()

	if t.LogBuffer.Len() > 0 {
		t.T.Log("\n" + t.LogBuffer.String())
	}
}

// AuroraSession returns a db.Session instance pointing at the aurora test
// database
func (t *T) AuroraSession() *db.Session {
	return &db.Session{
		DB:  t.AuroraDB,
		Ctx: t.Ctx,
	}
}

// Scenario loads the named sql scenario into the database
func (t *T) Scenario(name string) *T {
	LoadScenario(name)
	t.UpdateLedgerState()
	return t
}

// ScenarioWithoutAurora loads the named sql scenario into the database
func (t *T) ScenarioWithoutAurora(name string) *T {
	LoadScenarioWithoutAurora(name)
	t.UpdateLedgerState()
	return t
}

// UnmarshalPage populates dest with the records contained in the json-encoded page in r
func (t *T) UnmarshalPage(r io.Reader, dest interface{}) hal.Links {
	var env struct {
		Embedded struct {
			Records json.RawMessage `json:"records"`
		} `json:"_embedded"`
		Links struct {
			Self hal.Link `json:"self"`
			Next hal.Link `json:"next"`
			Prev hal.Link `json:"prev"`
		} `json:"_links"`
	}

	err := json.NewDecoder(r).Decode(&env)
	t.Require.NoError(err, "failed to decode page")

	err = json.Unmarshal(env.Embedded.Records, dest)
	t.Require.NoError(err, "failed to decode records")

	return env.Links
}

// UnmarshalNext extracts and returns the next link
func (t *T) UnmarshalNext(r io.Reader) string {
	var env struct {
		Links struct {
			Next struct {
				Href string `json:"href"`
			} `json:"next"`
		} `json:"_links"`
	}

	err := json.NewDecoder(r).Decode(&env)
	t.Require.NoError(err, "failed to decode page")
	return env.Links.Next.Href
}

// UnmarshalExtras extracts and returns extras content
func (t *T) UnmarshalExtras(r io.Reader) map[string]string {
	var resp struct {
		Extras map[string]string `json:"extras"`
	}

	err := json.NewDecoder(r).Decode(&resp)
	t.Require.NoError(err, "failed to decode page")

	return resp.Extras
}

// UpdateLedgerState updates the cached ledger state (or panicing on failure).
func (t *T) UpdateLedgerState() {
	var next ledger.State

	err := t.CoreSession().GetRaw(&next, `
		SELECT
			COALESCE(MAX(ledgerseq), 0) as core_latest
		FROM ledgerheaders
	`)

	if err != nil {
		panic(err)
	}

	err = t.AuroraSession().GetRaw(&next, `
			SELECT
				COALESCE(MIN(sequence), 0) as history_elder,
				COALESCE(MAX(sequence), 0) as history_latest
			FROM history_ledgers
		`)

	if err != nil {
		panic(err)
	}

	ledger.SetState(next)
}
