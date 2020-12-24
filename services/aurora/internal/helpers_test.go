package aurora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi"
	"github.com/diamnet/go/network"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/test"
	supportLog "github.com/diamnet/go/support/log"
	"github.com/throttled/throttled"
)

func NewTestApp() *App {
	return NewApp(NewTestConfig())
}

func NewTestConfig() Config {
	return Config{
		DatabaseURL:            test.DatabaseURL(),
		DiamNetCoreDatabaseURL: test.DiamNetCoreDatabaseURL(),
		RateQuota: &throttled.RateQuota{
			MaxRate:  throttled.PerHour(1000),
			MaxBurst: 100,
		},
		ConnectionTimeout:        55 * time.Second, // Default
		LogLevel:                 supportLog.InfoLevel,
		NetworkPassphrase:        network.TestNetworkPassphrase,
		IngestFailedTransactions: true,
	}
}

func NewRequestHelper(app *App) test.RequestHelper {
	return test.NewRequestHelper(app.web.router)
}

func ShouldBePageOf(actual interface{}, options ...interface{}) string {
	body := actual.(*bytes.Buffer)
	expected := options[0].(int)

	var result map[string]interface{}
	err := json.Unmarshal(body.Bytes(), &result)
	if err != nil {
		return fmt.Sprintf("Could not unmarshal json:\n%s\n", body.String())
	}

	embedded, ok := result["_embedded"]
	if !ok {
		return "No _embedded key in response"
	}

	records, ok := embedded.(map[string]interface{})["records"]
	if !ok {
		return "No records key in _embedded"
	}

	length := len(records.([]interface{}))
	if length != expected {
		return fmt.Sprintf("Expected %d records in page, got %d", expected, length)
	}

	return ""
}

func NewTestAction(ctx context.Context, path string) *Action {
	return &Action{
		App: NewTestApp(),
		Base: actions.Base{
			R: httptest.NewRequest("GET", path, nil).WithContext(context.WithValue(ctx, chi.RouteCtxKey, chi.NewRouteContext())),
		},
	}
}
