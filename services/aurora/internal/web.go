package aurora

import (
	"compress/flate"
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/rs/cors"
	"github.com/sebest/xff"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	hProblem "github.com/hcnet/go/services/aurora/internal/render/problem"
	"github.com/hcnet/go/services/aurora/internal/render/sse"
	"github.com/hcnet/go/services/aurora/internal/txsub/sequence"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/log"
	"github.com/hcnet/go/support/render/problem"
	"github.com/throttled/throttled"
)

const LRUCacheSize = 50000

// Web contains the http server related fields for aurora: the router,
// rate limiter, etc.
type web struct {
	appCtx             context.Context
	router             *chi.Mux
	rateLimiter        *throttled.HTTPRateLimiter
	sseUpdateFrequency time.Duration
	staleThreshold     uint
	ingestFailedTx     bool

	historyQ *history.Q
	coreQ    *core.Q

	requestTimer metrics.Timer
	failureMeter metrics.Meter
	successMeter metrics.Meter
}

func init() {
	// register problems
	problem.RegisterError(sql.ErrNoRows, problem.NotFound)
	problem.RegisterError(sequence.ErrNoMoreRoom, hProblem.ServerOverCapacity)
	problem.RegisterError(db2.ErrInvalidCursor, problem.BadRequest)
	problem.RegisterError(db2.ErrInvalidLimit, problem.BadRequest)
	problem.RegisterError(db2.ErrInvalidOrder, problem.BadRequest)
	problem.RegisterError(sse.ErrRateLimited, hProblem.RateLimitExceeded)
}

// mustInitWeb installed a new Web instance onto the provided app object.
func mustInitWeb(ctx context.Context, hq *history.Q, cq *core.Q, updateFreq time.Duration, threshold uint, ingestFailedTx bool) *web {
	if hq == nil {
		log.Fatal("missing history DB for installing the web instance")
	}
	if cq == nil {
		log.Fatal("missing core DB for installing the web instance")
	}

	return &web{
		appCtx:             ctx,
		router:             chi.NewRouter(),
		historyQ:           hq,
		coreQ:              cq,
		sseUpdateFrequency: updateFreq,
		staleThreshold:     threshold,
		ingestFailedTx:     ingestFailedTx,
		requestTimer:       metrics.NewTimer(),
		failureMeter:       metrics.NewMeter(),
		successMeter:       metrics.NewMeter(),
	}
}

// mustInstallMiddlewares installs the middleware stack used for aurora onto the
// provided app.
// Note that a request will go through the middlewares from top to bottom.
func (w *web) mustInstallMiddlewares(app *App, connTimeout time.Duration) {
	if w == nil {
		log.Fatal("missing web instance for installing middlewares")
	}

	r := w.router
	r.Use(chimiddleware.Timeout(connTimeout))
	r.Use(chimiddleware.StripSlashes)

	//TODO: remove this middleware
	r.Use(appContextMiddleware(app))

	r.Use(requestCacheHeadersMiddleware)
	r.Use(chimiddleware.RequestID)
	r.Use(contextMiddleware)
	r.Use(xff.Handler)
	r.Use(loggerMiddleware)
	r.Use(requestMetricsMiddleware)
	r.Use(recoverMiddleware)
	r.Use(chimiddleware.Compress(flate.DefaultCompression, "application/hal+json"))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
	})
	r.Use(c.Handler)

	r.Use(w.RateLimitMiddleware)
}

// mustInstallActions installs the routing configuration of aurora onto the
// provided app.  All route registration should be implemented here.
func (w *web) mustInstallActions(enableAssetStats bool, friendbotURL *url.URL) {
	if w == nil {
		log.Fatal("missing web instance for installing web actions")
	}

	r := w.router
	r.Get("/", RootAction{}.Handle)
	r.Get("/metrics", MetricsAction{}.Handle)

	// ledger actions
	r.Route("/ledgers", func(r chi.Router) {
		r.Get("/", LedgerIndexAction{}.Handle)
		r.Route("/{ledger_id}", func(r chi.Router) {
			r.Get("/", LedgerShowAction{}.Handle)
			r.Get("/transactions", w.streamIndexActionHandler(w.getTransactionPage, w.streamTransactions))
			r.Get("/operations", OperationIndexAction{}.Handle)
			r.Get("/payments", OperationIndexAction{OnlyPayments: true}.Handle)
			r.Get("/effects", EffectIndexAction{}.Handle)
		})
	})

	// account actions
	r.Route("/accounts", func(r chi.Router) {
		r.With(requiresExperimentalIngestion).
			Get("/", accountIndexActionHandler(w.getAccountPage))
		r.Route("/{account_id}", func(r chi.Router) {
			r.Get("/", w.streamShowActionHandler(w.getAccountInfo, true))
			r.Get("/transactions", w.streamIndexActionHandler(w.getTransactionPage, w.streamTransactions))
			r.Get("/operations", OperationIndexAction{}.Handle)
			r.Get("/payments", OperationIndexAction{OnlyPayments: true}.Handle)
			r.Get("/effects", EffectIndexAction{}.Handle)
			r.Get("/offers", OffersByAccountAction{}.Handle)
			r.Get("/trades", TradeIndexAction{}.Handle)
			r.Get("/data/{key}", DataShowAction{}.Handle)
		})
	})

	// transaction history actions
	r.Route("/transactions", func(r chi.Router) {
		r.Get("/", w.streamIndexActionHandler(w.getTransactionPage, w.streamTransactions))
		r.Route("/{tx_id}", func(r chi.Router) {
			r.Get("/", showActionHandler(w.getTransactionResource))
			r.Get("/operations", OperationIndexAction{}.Handle)
			r.Get("/payments", OperationIndexAction{OnlyPayments: true}.Handle)
			r.Get("/effects", EffectIndexAction{}.Handle)
		})
	})

	// operation actions
	r.Route("/operations", func(r chi.Router) {
		r.Get("/", OperationIndexAction{}.Handle)
		r.Get("/{id}", OperationShowAction{}.Handle)
		r.Get("/{op_id}/effects", EffectIndexAction{}.Handle)
	})

	// payment actions
	r.Get("/payments", OperationIndexAction{OnlyPayments: true}.Handle)

	// effect actions
	r.Get("/effects", EffectIndexAction{}.Handle)

	// trading related endpoints
	r.Get("/trades", TradeIndexAction{}.Handle)
	r.Get("/trade_aggregations", TradeAggregateIndexAction{}.Handle)
	r.Route("/offers", func(r chi.Router) {
		r.With(acceptOnlyJSON, requiresExperimentalIngestion).
			Get("/{id}", getOfferResource)
		r.Get("/{offer_id}/trades", TradeIndexAction{}.Handle)
	})
	r.Get("/order_book", OrderBookShowAction{}.Handle)

	// Transaction submission API
	r.Post("/transactions", TransactionCreateAction{}.Handle)
	r.Get("/paths", PathIndexAction{}.Handle)
	r.Get("/paths/strict-receive", PathIndexAction{}.Handle)
	r.Get("/paths/strict-send", FixedPathIndexAction{}.Handle)

	if enableAssetStats {
		// Asset related endpoints
		r.Get("/assets", AssetsAction{}.Handle)
	}

	// Network state related endpoints
	r.Get("/fee_stats", OperationFeeStatsAction{}.Handle)

	// friendbot
	if friendbotURL != nil {
		redirectFriendbot := func(w http.ResponseWriter, r *http.Request) {
			redirectURL := friendbotURL.String() + "?" + r.URL.RawQuery
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		}
		r.Post("/friendbot", redirectFriendbot)
		r.Get("/friendbot", redirectFriendbot)
	}

	r.NotFound(NotFoundAction{}.Handle)
}

func maybeInitWebRateLimiter(rateQuota *throttled.RateQuota) *throttled.HTTPRateLimiter {
	// Disabled
	if rateQuota == nil {
		return nil
	}

	rateLimiter, err := throttled.NewGCRARateLimiter(LRUCacheSize, *rateQuota)
	if err != nil {
		log.Fatalf("unable to create RateLimiter: %v", err)
	}

	return &throttled.HTTPRateLimiter{
		RateLimiter:   rateLimiter,
		DeniedHandler: &RateLimitExceededAction{Action{}},
		VaryBy:        VaryByRemoteIP{},
	}
}

type VaryByRemoteIP struct{}

func (v VaryByRemoteIP) Key(r *http.Request) string {
	return remoteAddrIP(r)
}

func remoteAddrIP(r *http.Request) string {
	// To support IPv6
	lastSemicolon := strings.LastIndex(r.RemoteAddr, ":")
	if lastSemicolon == -1 {
		return r.RemoteAddr
	} else {
		return r.RemoteAddr[0:lastSemicolon]
	}
}

// auroraSession returns a new session that loads data from the aurora
// database. The returned session is bound to `ctx`.
func (w *web) auroraSession(ctx context.Context) (*db.Session, error) {
	err := errorIfHistoryIsStale(w.isHistoryStale())
	if err != nil {
		return nil, err
	}

	return &db.Session{DB: w.historyQ.Session.DB, Ctx: ctx}, nil
}

// coreSession returns a new session that loads data from the hcnet core
// database. The returned session is bound to `ctx`.
func (w *web) coreSession(ctx context.Context) *db.Session {
	return &db.Session{DB: w.coreQ.Session.DB, Ctx: ctx}
}

// isHistoryStale returns true if the latest history ledger is more than
// `StaleThreshold` ledgers behind the latest core ledger
func (w *web) isHistoryStale() bool {
	if w.staleThreshold == 0 {
		return false
	}

	ls := ledger.CurrentState()
	return (ls.CoreLatest - ls.HistoryLatest) > int32(w.staleThreshold)
}

// errorIfHistoryIsStale returns a formatted error if isStale is true.
func errorIfHistoryIsStale(isStale bool) error {
	if !isStale {
		return nil
	}

	ls := ledger.CurrentState()
	err := hProblem.StaleHistory
	err.Extras = map[string]interface{}{
		"history_latest_ledger": ls.HistoryLatest,
		"core_latest_ledger":    ls.CoreLatest,
	}
	return err
}
