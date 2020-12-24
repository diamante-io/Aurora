package aurora

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/diamnet/go/services/aurora/internal/errors"
	"github.com/diamnet/go/services/aurora/internal/hchi"
	"github.com/diamnet/go/services/aurora/internal/httpx"
	"github.com/diamnet/go/services/aurora/internal/render"
	hProblem "github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/support/log"
	"github.com/diamnet/go/support/render/problem"
)

// appContextMiddleware adds the "app" context into every request, so that subsequence appContextMiddleware
// or handlers can retrieve a aurora.App instance
func appContextMiddleware(app *App) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withAppContext(r.Context(), app)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requestCacheHeadersMiddleware adds caching headers to each response.
func requestCacheHeadersMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Before changing this read Stack Overflow answer about staled request
		// in older versions of Chrome:
		// https://stackoverflow.com/questions/27513994/chrome-stalls-when-making-multiple-requests-to-same-resource
		w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")
		h.ServeHTTP(w, r)
	})
}

func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = hchi.WithChiRequestID(ctx)
		ctx, cancel := httpx.RequestContext(ctx, w, r)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const (
	clientNameHeader    = "X-Client-Name"
	clientVersionHeader = "X-Client-Version"
	appNameHeader       = "X-App-Name"
	appVersionHeader    = "X-App-Version"
)

// loggerMiddleware logs http requests and resposnes to the logging subsytem of aurora.
func loggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		logger := log.WithField("req", middleware.GetReqID(ctx))
		ctx = log.Set(ctx, logger)

		// Checking `Accept` header from user request because if the streaming connection
		// is reset before sending the first event no Content-Type header is sent in a response.
		acceptHeader := r.Header.Get("Accept")
		streaming := strings.Contains(acceptHeader, render.MimeEventStream)

		logStartOfRequest(ctx, r, streaming)
		then := time.Now()

		h.ServeHTTP(mw, r.WithContext(ctx))

		duration := time.Since(then)
		logEndOfRequest(ctx, r, duration, mw, streaming)
	})
}

// getClientData gets client data (name or version) from header or GET parameter
// (useful when not possible to set headers, like in EventStream).
func getClientData(r *http.Request, headerName string) string {
	value := r.Header.Get(headerName)
	if value != "" {
		return value
	}

	value = r.URL.Query().Get(headerName)
	if value == "" {
		value = "undefined"
	}

	return value
}

func logStartOfRequest(ctx context.Context, r *http.Request, streaming bool) {
	log.Ctx(ctx).WithFields(log.F{
		"client_name":    getClientData(r, clientNameHeader),
		"client_version": getClientData(r, clientVersionHeader),
		"app_name":       getClientData(r, appNameHeader),
		"app_version":    getClientData(r, appVersionHeader),
		"forwarded_ip":   firstXForwardedFor(r),
		"host":           r.Host,
		"ip":             remoteAddrIP(r),
		"ip_port":        r.RemoteAddr,
		"method":         r.Method,
		"path":           r.URL.String(),
		"streaming":      streaming,
	}).Info("Starting request")
}

func logEndOfRequest(ctx context.Context, r *http.Request, duration time.Duration, mw middleware.WrapResponseWriter, streaming bool) {
	routePattern := chi.RouteContext(r.Context()).RoutePattern()
	// Can be empty when request did not reached the final route (ex. blocked by
	// a middleware). More info: https://github.com/go-chi/chi/issues/270
	if routePattern == "" {
		routePattern = "undefined"
	}

	log.Ctx(ctx).WithFields(log.F{
		"bytes":          mw.BytesWritten(),
		"client_name":    getClientData(r, clientNameHeader),
		"client_version": getClientData(r, clientVersionHeader),
		"app_name":       getClientData(r, appNameHeader),
		"app_version":    getClientData(r, appVersionHeader),
		"duration":       duration.Seconds(),
		"forwarded_ip":   firstXForwardedFor(r),
		"host":           r.Host,
		"ip":             remoteAddrIP(r),
		"ip_port":        r.RemoteAddr,
		"method":         r.Method,
		"path":           r.URL.String(),
		"route":          routePattern,
		"status":         mw.Status(),
		"streaming":      streaming,
	}).Info("Finished request")
}

func firstXForwardedFor(r *http.Request) string {
	return strings.TrimSpace(strings.SplitN(r.Header.Get("X-Forwarded-For"), ",", 2)[0])
}

func (w *web) RateLimitMiddleware(next http.Handler) http.Handler {
	if w.rateLimiter == nil {
		return next
	}
	return w.rateLimiter.RateLimit(next)
}

// recoverMiddleware helps the server recover from panics. It ensures that
// no request can fully bring down the aurora server, and it also logs the
// panics to the logging subsystem.
func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rec := recover(); rec != nil {
				err := errors.FromPanic(rec)
				errors.ReportToSentry(err, r)
				problem.Render(ctx, w, err)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

// requestMetricsMiddleware records success and failures using a meter, and times every request
func requestMetricsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := AppFromContext(r.Context())
		mw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		app.web.requestTimer.Time(func() {
			h.ServeHTTP(mw.(http.ResponseWriter), r)
		})

		if 200 <= mw.Status() && mw.Status() < 400 {
			// a success is in [200, 400)
			app.web.successMeter.Mark(1)
		} else if 400 <= mw.Status() && mw.Status() < 600 {
			// a success is in [400, 600)
			app.web.failureMeter.Mark(1)
		}
	})
}

// acceptOnlyJSON inspects the accept header of the request and responds with
// an error if the content type is not JSON
func acceptOnlyJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := render.Negotiate(r)
		if contentType != render.MimeHal && contentType != render.MimeJSON {
			problem.Render(r.Context(), w, hProblem.NotAcceptable)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// requiresExperimentalIngestion is a middleware which enables a handler
// if the experimental ingestion system is enabled and initialized
func requiresExperimentalIngestion(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := AppFromContext(ctx)
		if !app.config.EnableExperimentalIngestion {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		lastIngestedLedger, err := app.HistoryQ().GetLastLedgerExpIngestNonBlocking()
		if err != nil {
			problem.Render(r.Context(), w, err)
			return
		}

		// expingest has not finished processing any ledger so no data.
		if lastIngestedLedger == 0 {
			problem.Render(r.Context(), w, hProblem.StillIngesting)
			return
		}

		h.ServeHTTP(w, r)
	})
}
