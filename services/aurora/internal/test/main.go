// Package test contains simple test helpers that should not
// have any dependencies on aurora's packages.  think constants,
// custom matchers, generic helpers etc.
package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	tdb "github.com/diamnet/go/services/aurora/internal/test/db"
	"github.com/diamnet/go/support/log"
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
	T         *testing.T
	Assert    *assert.Assertions
	Require   *require.Assertions
	Ctx       context.Context
	AuroraDB *sqlx.DB
	CoreDB    *sqlx.DB
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

// OverrideLogger calls StartTest on default logger. This is used
// by the testing system so that we can collect output from logs during test
// runs.  Panics if the logger is already overridden.
func OverrideLogger() {
	if endLogTest != nil {
		panic("logger already overridden")
	}

	endLogTest = log.StartTest(log.DebugLevel)
}

// RestoreLogger restores the default aurora logger after it is overridden
// using a call to `OverrideLogger`.  Panics if the default logger is not
// presently overridden.
func RestoreLogger() []logrus.Entry {
	if endLogTest == nil {
		panic("logger not overridden, cannot restore")
	}

	entries := endLogTest()
	endLogTest = nil
	return entries
}

// Start initializes a new test helper object and conceptually "starts" a new
// test
func Start(t *testing.T) *T {
	result := &T{}

	result.T = t

	OverrideLogger()

	result.Ctx = log.Set(context.Background(), log.DefaultLogger)
	result.AuroraDB = Database(t)
	result.CoreDB = DiamnetCoreDatabase(t)
	result.Assert = assert.New(t)
	result.Require = require.New(t)

	return result
}

// DiamnetCoreDatabase returns a connection to the diamnet core test database
//
// DEPRECATED:  use `DiamnetCore()` from test/db package
func DiamnetCoreDatabase(t *testing.T) *sqlx.DB {
	return tdb.DiamnetCore(t)
}

// DiamnetCoreDatabaseURL returns the database connection the url any test
// use when connecting to the diamnet-core database
//
// DEPRECATED:  use `DiamnetCoreURL()` from test/db package
func DiamnetCoreDatabaseURL() string {
	return tdb.DiamnetCoreURL()
}

var endLogTest func() []logrus.Entry = nil
