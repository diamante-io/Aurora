package aurora

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
)

type HTTPT struct {
	Assert *Assertions
	App    *App
	RH     test.RequestHelper
	*test.T
}

// StartHTTPTest is a helper function to setup a new test that will make http
// requests. Pair it with a deferred call to FinishHTTPTest.
func StartHTTPTest(t *testing.T, scenario string) *HTTPT {
	ret := &HTTPT{T: test.Start(t).Scenario(scenario)}
	ret.App = NewTestApp()
	ret.RH = test.NewRequestHelper(ret.App.web.router)
	ret.Assert = &Assertions{ret.T.Assert}
	ret.App.UpdateLedgerState()

	return ret
}

// Get delegates to the test's request helper
func (ht *HTTPT) Get(
	path string,
	fn ...func(*http.Request),
) *httptest.ResponseRecorder {
	return ht.RH.Get(path, fn...)
}

// GetWithParams delegates to the test's request helper and encodes along with the query params
func (ht *HTTPT) GetWithParams(
	path string,
	queryParams url.Values,
	fn ...func(*http.Request),
) *httptest.ResponseRecorder {
	return ht.RH.Get(path+"?"+queryParams.Encode(), fn...)
}

// Finish closes the test app and finishes the test
func (ht *HTTPT) Finish() {
	ht.T.Finish()
	ht.App.Close()
}

// Post delegates to the test's request helper
func (ht *HTTPT) Post(
	path string,
	form url.Values,
	mods ...func(*http.Request),
) *httptest.ResponseRecorder {
	return ht.RH.Post(path, form, mods...)
}

// ReapHistory causes the test server to run `DeleteUnretainedHistory`, after
// setting the retention count to the provided number.
func (ht *HTTPT) ReapHistory(retention uint) {
	ht.App.reaper.RetentionCount = retention
	err := ht.App.DeleteUnretainedHistory()
	ht.Require.NoError(err)
	ht.App.UpdateLedgerState()
}
