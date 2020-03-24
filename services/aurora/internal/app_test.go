package aurora

import (
	"net/http"
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestNewApp(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()

	config := NewTestConfig()
	config.SentryDSN = "Not a url"

	tt.Assert.Panics(func() {
		NewApp(config).Close()
	})
}

func TestGenericHTTPFeatures(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	// CORS
	w := ht.Get("/")
	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.Empty(w.HeaderMap.Get("Access-Control-Allow-Origin"))
	}

	w = ht.Get("/", func(r *http.Request) {
		r.Header.Set("Origin", "somewhere.com")
	})

	if ht.Assert.Equal(200, w.Code) {
		ht.Assert.Equal(
			"somewhere.com",
			w.HeaderMap.Get("Access-Control-Allow-Origin"),
		)
	}

	// Trailing slash is stripped
	w = ht.Get("/ledgers")
	ht.Assert.Equal(200, w.Code)
	w = ht.Get("/ledgers/")
	ht.Assert.Equal(200, w.Code)
}

func TestMetrics(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	hl := ht.App.historyLatestLedgerGauge
	he := ht.App.historyElderLedgerGauge
	cl := ht.App.coreLatestLedgerGauge

	ht.Require.EqualValues(0, hl.Value())
	ht.Require.EqualValues(0, he.Value())
	ht.Require.EqualValues(0, cl.Value())

	ht.App.UpdateLedgerState()
	ht.App.UpdateMetrics()

	ht.Require.EqualValues(3, hl.Value())
	ht.Require.EqualValues(1, he.Value())
	ht.Require.EqualValues(3, cl.Value())
}
