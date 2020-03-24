package http

import (
	stdhttp "net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_setup(t *testing.T) {

	// test that using no handler panics
	assert.Panics(t, func() {
		setup(Config{
			Handler: nil,
		})
	})

	// test defaults
	srv := setup(Config{
		Handler: stdhttp.NotFoundHandler(),
	})

	assert.Equal(t, DefaultShutdownGracePeriod, srv.Timeout)
	assert.Equal(t, DefaultListenAddr, srv.Server.Addr)
}
