package httpx

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientContext(t *testing.T) {
	// returns the default client
	assert.Equal(t, defaultClient, ClientFromContext(context.Background()))

	// returns a set client
	c := &http.Client{}
	ctx := ClientContext(context.Background(), c)
	assert.Equal(t, c, ClientFromContext(ctx))

	assert.Panics(t, func() {
		ClientContext(context.Background(), nil)
	})
}
