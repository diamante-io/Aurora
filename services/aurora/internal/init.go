package aurora

import (
	"context"
	"net/http"
	"net/url"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/gomodule/redigo/redis"
	metrics "github.com/rcrowley/go-metrics"
	ingestio "github.com/diamnet/go/exp/ingest/io"
	"github.com/diamnet/go/exp/orderbook"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/expingest"
	"github.com/diamnet/go/services/aurora/internal/ingest"
	"github.com/diamnet/go/services/aurora/internal/simplepath"
	"github.com/diamnet/go/services/aurora/internal/txsub"
	results "github.com/diamnet/go/services/aurora/internal/txsub/results/db"
	"github.com/diamnet/go/services/aurora/internal/txsub/sequence"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/log"
)

func mustInitAuroraDB(app *App) {
	session, err := db.Open("postgres", app.config.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot open Aurora DB: %v", err)
	}

	session.DB.SetMaxIdleConns(app.config.AuroraDBMaxIdleConnections)
	session.DB.SetMaxOpenConns(app.config.AuroraDBMaxOpenConnections)
	app.historyQ = &history.Q{session}
}

func mustInitCoreDB(app *App) {
	session, err := db.Open("postgres", app.config.DiamNetCoreDatabaseURL)
	if err != nil {
		log.Fatalf("cannot open Core DB: %v", err)
	}

	session.DB.SetMaxIdleConns(app.config.CoreDBMaxIdleConnections)
	session.DB.SetMaxOpenConns(app.config.CoreDBMaxOpenConnections)
	app.coreQ = &core.Q{session}
}

func initIngester(app *App) {
	if !app.config.Ingest {
		return
	}

	if app.config.NetworkPassphrase == "" {
		log.Fatal("Cannot start ingestion without network passphrase. Please confirm connectivity with diamnet-core.")
	}

	app.ingester = ingest.New(
		app.config.NetworkPassphrase,
		app.config.DiamNetCoreURL,
		app.CoreSession(context.Background()),
		app.AuroraSession(context.Background()),
		ingest.Config{
			EnableAssetStats:         app.config.EnableAssetStats,
			IngestFailedTransactions: app.config.IngestFailedTransactions,
			CursorName:               app.config.CursorName,
		},
	)

	app.ingester.SkipCursorUpdate = app.config.SkipCursorUpdate
	app.ingester.HistoryRetentionCount = app.config.HistoryRetentionCount
}

func initExpIngester(app *App, orderBookGraph *orderbook.OrderBookGraph) {
	var tempSet ingestio.TempSet = &ingestio.MemoryTempSet{}
	switch app.config.IngestStateReaderTempSet {
	case "postgres":
		tempSet = &ingestio.PostgresTempSet{
			Session: app.AuroraSession(context.Background()),
		}
	}

	var err error
	app.expingester, err = expingest.NewSystem(expingest.Config{
		CoreSession:    app.CoreSession(context.Background()),
		HistorySession: app.AuroraSession(context.Background()),
		// TODO:
		// Use the first archive for now. We don't have a mechanism to
		// use multiple archives at the same time currently.
		HistoryArchiveURL: app.config.HistoryArchiveURLs[0],
		DiamNetCoreURL:    app.config.DiamNetCoreURL,
		OrderBookGraph:    orderBookGraph,
		TempSet:           tempSet,
	})
	if err != nil {
		log.Panic(err)
	}
}

func initPathFinder(app *App, orderBookGraph *orderbook.OrderBookGraph) {
	if app.config.EnableExperimentalIngestion {
		app.paths = simplepath.NewInMemoryFinder(orderBookGraph)
	} else {
		app.paths = &simplepath.Finder{app.CoreQ()}
	}
}

// initSentry initialized the default sentry client with the configured DSN
func initSentry(app *App) {
	if app.config.SentryDSN == "" {
		return
	}

	log.WithField("dsn", app.config.SentryDSN).Info("Initializing sentry")
	err := raven.SetDSN(app.config.SentryDSN)
	if err != nil {
		log.Panic(err)
	}
}

// initLogglyLog attaches a loggly hook to our logging system.
func initLogglyLog(app *App) {
	if app.config.LogglyToken == "" {
		return
	}

	log.WithFields(log.F{
		"token": app.config.LogglyToken,
		"tag":   app.config.LogglyTag,
	}).Info("Initializing loggly hook")

	hook := log.NewLogglyHook(app.config.LogglyToken, app.config.LogglyTag)
	log.DefaultLogger.Logger.Hooks.Add(hook)

	go func() {
		<-app.ctx.Done()
		hook.Flush()
	}()
}

func initDbMetrics(app *App) {
	app.historyLatestLedgerGauge = metrics.NewGauge()
	app.historyElderLedgerGauge = metrics.NewGauge()
	app.coreLatestLedgerGauge = metrics.NewGauge()
	app.auroraConnGauge = metrics.NewGauge()
	app.coreConnGauge = metrics.NewGauge()
	app.goroutineGauge = metrics.NewGauge()
	app.metrics.Register("history.latest_ledger", app.historyLatestLedgerGauge)
	app.metrics.Register("history.elder_ledger", app.historyElderLedgerGauge)
	app.metrics.Register("diamnet_core.latest_ledger", app.coreLatestLedgerGauge)
	app.metrics.Register("history.open_connections", app.auroraConnGauge)
	app.metrics.Register("diamnet_core.open_connections", app.coreConnGauge)
	app.metrics.Register("goroutines", app.goroutineGauge)
}

func initIngesterMetrics(app *App) {
	if app.ingester == nil {
		return
	}
	app.metrics.Register("ingester.ingest_ledger",
		app.ingester.Metrics.IngestLedgerTimer)
	app.metrics.Register("ingester.clear_ledger",
		app.ingester.Metrics.ClearLedgerTimer)
}

func initTxSubMetrics(app *App) {
	app.submitter.Init()
	app.metrics.Register("txsub.buffered", app.submitter.Metrics.BufferedSubmissionsGauge)
	app.metrics.Register("txsub.open", app.submitter.Metrics.OpenSubmissionsGauge)
	app.metrics.Register("txsub.succeeded", app.submitter.Metrics.SuccessfulSubmissionsMeter)
	app.metrics.Register("txsub.failed", app.submitter.Metrics.FailedSubmissionsMeter)
	app.metrics.Register("txsub.total", app.submitter.Metrics.SubmissionTimer)
}

// initWebMetrics registers the metrics for the web server into the provided
// app's metrics registry.
func initWebMetrics(app *App) {
	app.metrics.Register("requests.total", app.web.requestTimer)
	app.metrics.Register("requests.succeeded", app.web.successMeter)
	app.metrics.Register("requests.failed", app.web.failureMeter)
}

func initRedis(app *App) {
	if app.config.RedisURL == "" {
		return
	}

	redisURL, err := url.Parse(app.config.RedisURL)
	if err != nil {
		log.Fatal(err)
	}

	app.redis = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        dialRedis(redisURL),
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// test the connection
	c := app.redis.Get()
	defer c.Close()

	_, err = c.Do("PING")
	if err != nil {
		log.Fatal(err)
	}
}

func dialRedis(redisURL *url.URL) func() (redis.Conn, error) {
	return func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", redisURL.Host)
		if err != nil {
			return nil, err
		}

		if redisURL.User == nil {
			return c, err
		}

		if pass, ok := redisURL.User.Password(); ok {
			if _, err = c.Do("AUTH", pass); err != nil {
				c.Close()
				return nil, err
			}
		}

		return c, err
	}
}

func initSubmissionSystem(app *App) {
	cq := &core.Q{Session: app.CoreSession(context.Background())}

	app.submitter = &txsub.System{
		Pending:         txsub.NewDefaultSubmissionList(),
		Submitter:       txsub.NewDefaultSubmitter(http.DefaultClient, app.config.DiamNetCoreURL),
		SubmissionQueue: sequence.NewManager(),
		Results: &results.DB{
			Core:    cq,
			History: &history.Q{Session: app.AuroraSession(context.Background())},
		},
		Sequences:         cq.SequenceProvider(),
		NetworkPassphrase: app.config.NetworkPassphrase,
	}
}
