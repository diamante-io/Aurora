package httpx

import (
	"context"
)

var DefaultHost = ""

func Host(ctx context.Context) string {
	r := RequestFromContext(ctx)

	if r == nil {
		return DefaultHost
	}

	if r.Host == "" {
		return DefaultHost
	}

	return r.Host
}
