package aurora

import (
	"log"
	"time"

	"github.com/diamnet/throttled"

	"github.com/diamnet/go/network"
	"github.com/diamnet/go/services/aurora/internal/test"
	supportLog "github.com/diamnet/go/support/log"
)

func NewTestApp() *App {
	app, err := NewApp(NewTestConfig())
	if err != nil {
		log.Fatal("cannot create app", err)
	}
	return app
}

func NewTestConfig() Config {
	return Config{
		DatabaseURL:            test.DatabaseURL(),
		DiamnetCoreDatabaseURL: test.DiamnetCoreDatabaseURL(),
		RateQuota: &throttled.RateQuota{
			MaxRate:  throttled.PerHour(1000),
			MaxBurst: 100,
		},
		ConnectionTimeout: 55 * time.Second, // Default
		LogLevel:          supportLog.InfoLevel,
		NetworkPassphrase: network.TestNetworkPassphrase,
	}
}

func NewRequestHelper(app *App) test.RequestHelper {
	return test.NewRequestHelper(app.webServer.Router.Mux)
}
