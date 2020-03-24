// Package test contains simple test helpers that should not
// have any dependencies on aurora's packages.  think constants,
// custom matchers, generic helpers etc.
package test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/hcnet/go/services/aurora/internal/logmetrics"
	tdb "github.com/hcnet/go/services/aurora/internal/test/db"
	"github.com/hcnet/go/support/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StaticMockServer is a test helper that records it's last request
type StaticMockServer struct {
	*httptest.Server
	LastRequest *http.Request
}

// T provides a common set of functionality for each test in aurora
type T struct {
	T          *testing.T
	Assert     *assert.Assertions
	Require    *require.Assertions
	Ctx        context.Context
	AuroraDB  *sqlx.DB
	CoreDB     *sqlx.DB
	Logger     *log.Entry
	LogMetrics *logmetrics.Metrics
	LogBuffer  *bytes.Buffer
}

// Context provides a context suitable for testing in tests that do not create
// a full App instance (in which case your tests should be using the app's
// context).  This context has a logger bound to it suitable for testing.
func Context() context.Context {
	return log.Set(context.Background(), testLogger)
}

// Database returns a connection to the aurora test database
//
// DEPRECATED:  use `Aurora()` from test/db package
func Database(t *testing.T) *sqlx.DB {
	return tdb.Aurora(t)
}

// DatabaseURL returns the database connection the url any test
// use when connecting to the history/aurora database
//
// DEPRECATED:  use `AuroraURL()` from test/db package
func DatabaseURL() string {
	return tdb.AuroraURL()
}

// LoadScenario populates the test databases with pre-created scenarios.  Each
// scenario is in the scenarios subfolder of this package and are a pair of
// sql files, one per database.
func LoadScenario(scenarioName string) {
	loadScenario(scenarioName, true)
}

// LoadScenarioWithoutAurora populates the test HcNet core database a with
// pre-created scenario.  Unlike `LoadScenario`, this
func LoadScenarioWithoutAurora(scenarioName string) {
	loadScenario(scenarioName, false)
}

// OverrideLogger sets the default logger used by aurora to `l`.  This is used
// by the testing system so that we can collect output from logs during test
// runs.  Panics if the logger is already overridden.
func OverrideLogger(l *log.Entry) {
	if oldDefault != nil {
		panic("logger already overridden")
	}

	oldDefault = log.DefaultLogger
	log.DefaultLogger = l
}

// RestoreLogger restores the default aurora logger after it is overridden
// using a call to `OverrideLogger`.  Panics if the default logger is not
// presently overridden.
func RestoreLogger() {
	if oldDefault == nil {
		panic("logger not overridden, cannot restore")
	}

	log.DefaultLogger = oldDefault
	oldDefault = nil
}

// Start initializes a new test helper object and conceptually "starts" a new
// test
func Start(t *testing.T) *T {
	result := &T{}

	result.T = t
	result.LogBuffer = new(bytes.Buffer)
	result.Logger, result.LogMetrics = logmetrics.New()
	result.Logger.Logger.Out = result.LogBuffer
	result.Logger.Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	result.Logger.Logger.Level = logrus.DebugLevel

	OverrideLogger(result.Logger)

	result.Ctx = log.Set(context.Background(), result.Logger)
	result.AuroraDB = Database(t)
	result.CoreDB = HcNetCoreDatabase(t)
	result.Assert = assert.New(t)
	result.Require = require.New(t)

	return result
}

// HcNetCoreDatabase returns a connection to the hcnet core test database
//
// DEPRECATED:  use `HcNetCore()` from test/db package
func HcNetCoreDatabase(t *testing.T) *sqlx.DB {
	return tdb.HcNetCore(t)
}

// HcNetCoreDatabaseURL returns the database connection the url any test
// use when connecting to the hcnet-core database
//
// DEPRECATED:  use `HcNetCoreURL()` from test/db package
func HcNetCoreDatabaseURL() string {
	return tdb.HcNetCoreURL()
}

var oldDefault *log.Entry = nil
