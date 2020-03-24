package http

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
)

// NewMux returns a new server mux configured with the common defaults used across all
// hcnet services.
func NewMux(behindProxy bool) *chi.Mux {
	mux := chi.NewMux()

	if behindProxy {
		mux.Use(middleware.RealIP)
	}

	mux.Use(middleware.RequestID)
	mux.Use(middleware.Recoverer)
	mux.Use(LoggingMiddleware)

	return mux
}

// NewAPIMux returns a new server mux configured with the common defaults used for a web API in
// hcnet.
func NewAPIMux(behindProxy bool) *chi.Mux {
	mux := NewMux(behindProxy)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	})

	mux.Use(c.Handler)
	return mux
}
