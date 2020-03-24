package http

import (
	"context"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/hcnet/go/support/http/mutil"
	"github.com/hcnet/go/support/log"
)

// LoggingMiddleware is a middleware that logs requests to the logger.
func LoggingMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		mw := mutil.WrapWriter(w)
		ctx := log.PushContext(r.Context(), func(l *log.Entry) *log.Entry {
			return l.WithFields(log.F{
				"req": middleware.GetReqID(r.Context()),
			})
		})

		logStartOfRequest(ctx, r)

		then := time.Now()
		next.ServeHTTP(mw, r)
		duration := time.Since(then)

		logEndOfRequest(ctx, r, duration, mw)
	})
}

// logStartOfRequest emits the logline that reports that an http request is
// beginning processing.
func logStartOfRequest(
	ctx context.Context,
	r *stdhttp.Request,
) {
	log.Ctx(ctx).WithFields(log.F{
		"subsys": "http",
		"path":   r.URL.String(),
		"method": r.Method,
		"ip":     r.RemoteAddr,
		"host":   r.Host,
	}).Info("starting request")
}

// logEndOfRequest emits the logline for the end of the request
func logEndOfRequest(
	ctx context.Context,
	r *stdhttp.Request,
	duration time.Duration,
	mw mutil.WriterProxy,
) {
	log.Ctx(ctx).WithFields(log.F{
		"subsys":   "http",
		"path":     r.URL.String(),
		"method":   r.Method,
		"status":   mw.Status(),
		"bytes":    mw.BytesWritten(),
		"duration": duration,
	}).Info("finished request")
}
