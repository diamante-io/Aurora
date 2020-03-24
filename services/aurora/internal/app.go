package aurora

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/hcnet/go/clients/hcnetcore"
	"github.com/hcnet/go/exp/orderbook"
	auroraContext "github.com/hcnet/go/services/aurora/internal/context"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/expingest"
	"github.com/hcnet/go/services/aurora/internal/ingest"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/logmetrics"
	"github.com/hcnet/go/services/aurora/internal/operationfeestats"
	"github.com/hcnet/go/services/aurora/internal/paths"
	"github.com/hcnet/go/services/aurora/internal/reap"
	"github.com/hcnet/go/services/aurora/internal/txsub"
	"github.com/hcnet/go/support/app"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
	"github.com/throttled/throttled"
	"golang.org/x/net/http2"
	graceful "gopkg.in/tylerb/graceful.v1"
)

// App represents the root of the state of a aurora instance.
type App struct {
	config                       Config
	web                          *web
	historyQ                     *history.Q
	coreQ                        *core.Q
	ctx                          context.Context
	cancel                       func()
	redis                        *redis.Pool
	coreVersion                  string
	auroraVersion               string
	currentProtocolVersion       int32
	coreSupportedProtocolVersion int32
	submitter                    *txsub.System
	paths                        paths.Finder
	ingester                     *ingest.System
	expingester                  *expingest.System
	reaper                       *reap.System
	ticks                        *time.Ticker

	// metrics
	metrics                  metrics.Registry
	historyLatestLedgerGauge metrics.Gauge
	historyElderLedgerGauge  metrics.Gauge
	auroraConnGauge         metrics.Gauge
	coreLatestLedgerGauge    metrics.Gauge
	coreConnGauge            metrics.Gauge
	goroutineGauge           metrics.Gauge
}

// NewApp constructs an new App instance from the provided config.
func NewApp(config Config) *App {
	a := &App{
		config:         config,
		auroraVersion: app.Version(),
		ticks:          time.NewTicker(1 * time.Second),
	}

	a.init()
	return a
}

// Serve starts the aurora web server, binding it to a socket, setting up
// the shutdown signals.
func (a *App) Serve() {
	http.Handle("/", a.web.router)

	addr := fmt.Sprintf(":%d", a.config.Port)

	srv := &graceful.Server{
		Timeout: 10 * time.Second,

		Server: &http.Server{
			Addr:              addr,
			Handler:           http.DefaultServeMux,
			ReadHeaderTimeout: 5 * time.Second,
		},

		ShutdownInitiated: func() {
			log.Info("received signal, gracefully stopping")
			a.Close()
		},
	}

	http2.ConfigureServer(srv.Server, nil)

	log.Infof("Starting aurora on %s (ingest: %v)", addr, a.config.Ingest)

	go a.run()

	if a.expingester != nil {
		go a.expingester.Run()
	}

	var err error
	if a.config.TLSCert != "" {
		err = srv.ListenAndServeTLS(a.config.TLSCert, a.config.TLSKey)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		log.Panic(err)
	}

	a.CloseDB()

	log.Info("stopped")
}

// Close cancels the app. It does not close DB connections - use App.CloseDB().
func (a *App) Close() {
	a.cancel()
	if a.expingester != nil {
		a.expingester.Shutdown()
	}
	a.ticks.Stop()
}

// CloseDB closes DB connections. When using during web server shut down make
// sure all requests are first properly finished to avoid "sql: database is
// closed" errors.
func (a *App) CloseDB() {
	a.historyQ.Session.DB.Close()
	a.coreQ.Session.DB.Close()
}

// HistoryQ returns a helper object for performing sql queries against the
// history portion of aurora's database.
func (a *App) HistoryQ() *history.Q {
	return a.historyQ
}

// AuroraSession returns a new session that loads data from the aurora
// database. The returned session is bound to `ctx`.
func (a *App) AuroraSession(ctx context.Context) *db.Session {
	return &db.Session{DB: a.historyQ.Session.DB, Ctx: ctx}
}

// CoreSession returns a new session that loads data from the hcnet core
// database. The returned session is bound to `ctx`.
func (a *App) CoreSession(ctx context.Context) *db.Session {
	return &db.Session{DB: a.coreQ.Session.DB, Ctx: ctx}
}

// CoreQ returns a helper object for performing sql queries aginst the
// hcnet core database.
func (a *App) CoreQ() *core.Q {
	return a.coreQ
}

// IsHistoryStale returns true if the latest history ledger is more than
// `StaleThreshold` ledgers behind the latest core ledger
func (a *App) IsHistoryStale() bool {
	if a.config.StaleThreshold == 0 {
		return false
	}

	ls := ledger.CurrentState()
	return (ls.CoreLatest - ls.HistoryLatest) > int32(a.config.StaleThreshold)
}

// UpdateLedgerState triggers a refresh of several metrics gauges, such as open
// db connections and ledger state
func (a *App) UpdateLedgerState() {
	var next ledger.State

	logErr := func(err error, msg string) {
		log.WithStack(err).WithField("err", err.Error()).Error(msg)
	}

	err := a.CoreQ().LatestLedger(&next.CoreLatest)
	if err != nil {
		logErr(err, "failed to load the latest known ledger state from core DB")
		return
	}

	err = a.HistoryQ().LatestLedger(&next.HistoryLatest)
	if err != nil {
		logErr(err, "failed to load the latest known ledger state from history DB")
		return
	}

	err = a.HistoryQ().ElderLedger(&next.HistoryElder)
	if err != nil {
		logErr(err, "failed to load the oldest known ledger state from history DB")
		return
	}

	next.ExpHistoryLatest, err = a.HistoryQ().GetLastLedgerExpIngestNonBlocking()
	if err != nil {
		logErr(err, "failed to load the oldest known exp ledger state from history DB")
		return
	}

	ledger.SetState(next)
}

// UpdateOperationFeeStatsState triggers a refresh of several operation fee metrics.
func (a *App) UpdateOperationFeeStatsState() {
	var (
		next          operationfeestats.State
		latest        history.LatestLedger
		feeStats      history.FeeStats
		capacityStats history.LedgerCapacityUsageStats
	)

	logErr := func(err error, msg string) {
		// If DB is empty ignore the error
		if errors.Cause(err) == sql.ErrNoRows {
			return
		}

		log.WithStack(err).WithField("err", err.Error()).Error(msg)
	}

	cur := operationfeestats.CurrentState()

	err := a.HistoryQ().LatestLedgerBaseFeeAndSequence(&latest)
	if err != nil {
		logErr(err, "failed to load the latest known ledger's base fee and sequence number")
		return
	}

	// finish early if no new ledgers
	if cur.LastLedger == int64(latest.Sequence) {
		return
	}

	next.LastBaseFee = int64(latest.BaseFee)
	next.LastLedger = int64(latest.Sequence)

	err = a.HistoryQ().OperationFeeStats(latest.Sequence, &feeStats)
	if err != nil {
		logErr(err, "failed to load operation fee stats")
		return
	}

	err = a.HistoryQ().LedgerCapacityUsageStats(latest.Sequence, &capacityStats)
	if err != nil {
		logErr(err, "failed to load ledger capacity usage stats")
		return
	}

	next.LedgerCapacityUsage = capacityStats.CapacityUsage.String

	// if no transactions in last 5 ledgers, return
	// latest ledger's base fee for all
	if !feeStats.Mode.Valid && !feeStats.Min.Valid {
		next.FeeMin = next.LastBaseFee
		next.FeeMode = next.LastBaseFee
		next.FeeP10 = next.LastBaseFee
		next.FeeP20 = next.LastBaseFee
		next.FeeP30 = next.LastBaseFee
		next.FeeP40 = next.LastBaseFee
		next.FeeP50 = next.LastBaseFee
		next.FeeP60 = next.LastBaseFee
		next.FeeP70 = next.LastBaseFee
		next.FeeP80 = next.LastBaseFee
		next.FeeP90 = next.LastBaseFee
		next.FeeP95 = next.LastBaseFee
		next.FeeP99 = next.LastBaseFee
	} else {
		next.FeeMin = feeStats.Min.Int64
		next.FeeMode = feeStats.Mode.Int64
		next.FeeP10 = feeStats.P10.Int64
		next.FeeP20 = feeStats.P20.Int64
		next.FeeP30 = feeStats.P30.Int64
		next.FeeP40 = feeStats.P40.Int64
		next.FeeP50 = feeStats.P50.Int64
		next.FeeP60 = feeStats.P60.Int64
		next.FeeP70 = feeStats.P70.Int64
		next.FeeP80 = feeStats.P80.Int64
		next.FeeP90 = feeStats.P90.Int64
		next.FeeP95 = feeStats.P95.Int64
		next.FeeP99 = feeStats.P99.Int64
	}

	operationfeestats.SetState(next)
}

// UpdateHcNetCoreInfo updates the value of coreVersion,
// currentProtocolVersion, and coreSupportedProtocolVersion from the HcNet
// core API.
func (a *App) UpdateHcNetCoreInfo() {
	if a.config.HcNetCoreURL == "" {
		return
	}

	core := &hcnetcore.Client{
		URL: a.config.HcNetCoreURL,
	}

	resp, err := core.Info(context.Background())
	if err != nil {
		log.Warnf("could not load hcnet-core info: %s", err)
		return
	}

	// Check if NetworkPassphrase is different, if so exit Aurora as it can break the
	// state of the application.
	if resp.Info.Network != a.config.NetworkPassphrase {
		log.Errorf(
			"Network passphrase of hcnet-core (%s) does not match Aurora configuration (%s). Exiting...",
			resp.Info.Network,
			a.config.NetworkPassphrase,
		)
		os.Exit(1)
	}

	a.coreVersion = resp.Info.Build
	a.currentProtocolVersion = int32(resp.Info.Ledger.Version)
	a.coreSupportedProtocolVersion = int32(resp.Info.ProtocolVersion)
}

// UpdateMetrics triggers a refresh of several metrics gauges, such as open
// db connections and ledger state
func (a *App) UpdateMetrics() {
	a.goroutineGauge.Update(int64(runtime.NumGoroutine()))
	ls := ledger.CurrentState()
	a.historyLatestLedgerGauge.Update(int64(ls.HistoryLatest))
	a.historyElderLedgerGauge.Update(int64(ls.HistoryElder))
	a.coreLatestLedgerGauge.Update(int64(ls.CoreLatest))

	a.auroraConnGauge.Update(int64(a.historyQ.Session.DB.Stats().OpenConnections))
	a.coreConnGauge.Update(int64(a.coreQ.Session.DB.Stats().OpenConnections))
}

// DeleteUnretainedHistory forwards to the app's reaper.  See
// `reap.DeleteUnretainedHistory` for details
func (a *App) DeleteUnretainedHistory() error {
	return a.reaper.DeleteUnretainedHistory()
}

// Tick triggers aurora to update all of it's background processes such as
// transaction submission, metrics, ingestion and reaping.
func (a *App) Tick() {
	var wg sync.WaitGroup
	log.Debug("ticking app")
	// update ledger state, operation fee state, and hcnet-core info in parallel
	wg.Add(3)
	go func() { a.UpdateLedgerState(); wg.Done() }()
	go func() { a.UpdateOperationFeeStatsState(); wg.Done() }()
	go func() { a.UpdateHcNetCoreInfo(); wg.Done() }()
	wg.Wait()

	if a.ingester != nil {
		go a.ingester.Tick()
	}

	wg.Add(2)
	go func() { a.reaper.Tick(); wg.Done() }()
	go func() { a.submitter.Tick(a.ctx); wg.Done() }()
	wg.Wait()

	// finally, update metrics
	a.UpdateMetrics()
	log.Debug("finished ticking app")
}

// Init initializes app, using the config to populate db connections and
// whatnot.
func (a *App) init() {
	// app-context
	a.ctx, a.cancel = context.WithCancel(context.Background())

	// log
	log.DefaultLogger.Logger.Level = a.config.LogLevel
	log.DefaultLogger.Logger.Hooks.Add(logmetrics.DefaultMetrics)

	// sentry
	initSentry(a)

	// loggly
	initLogglyLog(a)

	// hcnetCoreInfo
	a.UpdateHcNetCoreInfo()

	// aurora-db and core-db
	mustInitAuroraDB(a)
	mustInitCoreDB(a)

	// ingester
	initIngester(a)

	var orderBookGraph *orderbook.OrderBookGraph
	if a.config.EnableExperimentalIngestion {
		orderBookGraph = orderbook.NewOrderBookGraph()
		// expingester
		initExpIngester(a, orderBookGraph)
	}

	// txsub
	initSubmissionSystem(a)

	// path-finder
	initPathFinder(a, orderBookGraph)

	// reaper
	a.reaper = reap.New(a.config.HistoryRetentionCount, a.AuroraSession(context.Background()))

	// web.init
	a.web = mustInitWeb(a.ctx, a.historyQ, a.coreQ, a.config.SSEUpdateFrequency, a.config.StaleThreshold, a.config.IngestFailedTransactions)

	// web.rate-limiter
	a.web.rateLimiter = maybeInitWebRateLimiter(a.config.RateQuota)

	// web.middleware
	// Note that we passed in `a` here for putting the whole App in the context.
	// This parameter will be removed soon.
	a.web.mustInstallMiddlewares(a, a.config.ConnectionTimeout)

	// web.actions
	a.web.mustInstallActions(
		a.config.EnableAssetStats,
		a.config.FriendbotURL,
	)

	// metrics and log.metrics
	a.metrics = metrics.NewRegistry()
	for level, meter := range *logmetrics.DefaultMetrics {
		a.metrics.Register(fmt.Sprintf("logging.%s", level), meter)
	}

	// db-metrics
	initDbMetrics(a)

	// web.metrics
	initWebMetrics(a)

	// txsub.metrics
	initTxSubMetrics(a)

	// ingester.metrics
	initIngesterMetrics(a)

	// redis
	initRedis(a)
}

// run is the function that runs in the background that triggers Tick each
// second
func (a *App) run() {
	for {
		select {
		case <-a.ticks.C:
			a.Tick()
		case <-a.ctx.Done():
			log.Info("finished background ticker")
			return
		}
	}
}

// withAppContext create a context on from the App type.
func withAppContext(ctx context.Context, a *App) context.Context {
	return context.WithValue(ctx, &auroraContext.AppContextKey, a)
}

// GetRateLimiter returns the HTTPRateLimiter of the App.
func (a *App) GetRateLimiter() *throttled.HTTPRateLimiter {
	return a.web.rateLimiter
}

// AppFromContext returns the set app, if one has been set, from the
// provided context returns nil if no app has been set.
func AppFromContext(ctx context.Context) *App {
	if ctx == nil {
		return nil
	}

	val, _ := ctx.Value(&auroraContext.AppContextKey).(*App)
	return val
}
