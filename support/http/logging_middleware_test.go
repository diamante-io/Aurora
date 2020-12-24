package http

import (
	stdhttp "net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/diamnet/go/support/http/httptest"
	"github.com/diamnet/go/support/log"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMiddleware(t *testing.T) {
	done := log.DefaultLogger.StartTest(log.InfoLevel)
	mux := chi.NewMux()

	mux.Use(middleware.RequestID)
	mux.Use(LoggingMiddleware)

	mux.Get("/", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	}))
	mux.Handle("/not_found", stdhttp.NotFoundHandler())

	src := httptest.NewServer(t, mux)
	src.GET("/").Expect().Status(stdhttp.StatusOK)
	src.GET("/not_found").Expect().Status(stdhttp.StatusNotFound)

	// get the log buffer and ensure it has both the start and end log lines for
	// each request
	logged := done()
	if assert.Len(t, logged, 4, "unexpected log line count") {
		assert.Equal(t, "starting request", logged[0].Message)
		assert.Equal(t, "starting request", logged[2].Message)
		assert.Equal(t, "finished request", logged[1].Message)
		assert.Equal(t, "finished request", logged[3].Message)
	}

	for _, line := range logged {
		assert.Equal(t, "http", line.Data["subsys"])
		assert.Equal(t, "GET", line.Data["method"])
		assert.NotEmpty(t, line.Data["req"])
		assert.NotEmpty(t, line.Data["path"])
	}
}
