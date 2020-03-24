package httpx

import (
	"context"
	"net/url"
)

// BaseURL returns the "base" url for this request, defined as a url containing
// the Host and Scheme portions of the request uri.
func BaseURL(ctx context.Context) *url.URL {
	r := RequestFromContext(ctx)
	if r == nil {
		return nil
	}

	var scheme string
	switch {
	case r.Header.Get("X-Forwarded-Proto") != "":
		scheme = r.Header.Get("X-Forwarded-Proto")
	case r.TLS != nil:
		scheme = "https"
	default:
		scheme = "http"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   r.Host,
	}
}
