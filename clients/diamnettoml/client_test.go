package diamnettoml

import (
	"strings"
	"testing"

	"net/http"

	"github.com/diamnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientURL(t *testing.T) {
	//HACK:  we're testing an internal method rather than setting up a http client
	//mock.

	c := &Client{UseHTTP: false}
	assert.Equal(t, "https://diamnet.org/.well-known/diamnet.toml", c.url("diamnet.org"))

	c = &Client{UseHTTP: true}
	assert.Equal(t, "http://diamnet.org/.well-known/diamnet.toml", c.url("diamnet.org"))
}

func TestClient(t *testing.T) {
	h := httptest.NewClient()
	c := &Client{HTTP: h}

	// happy path
	h.
		On("GET", "https://diamnet.org/.well-known/diamnet.toml").
		ReturnString(http.StatusOK,
			`FEDERATION_SERVER="https://localhost/federation"`,
		)
	stoml, err := c.GetDiamnetToml("diamnet.org")
	require.NoError(t, err)
	assert.Equal(t, "https://localhost/federation", stoml.FederationServer)

	// diamnet.toml exceeds limit
	h.
		On("GET", "https://toobig.org/.well-known/diamnet.toml").
		ReturnString(http.StatusOK,
			`FEDERATION_SERVER="https://localhost/federation`+strings.Repeat("0", DiamnetTomlMaxSize)+`"`,
		)
	stoml, err = c.GetDiamnetToml("toobig.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "diamnet.toml response exceeds")
	}

	// not found
	h.
		On("GET", "https://missing.org/.well-known/diamnet.toml").
		ReturnNotFound()
	stoml, err = c.GetDiamnetToml("missing.org")
	assert.EqualError(t, err, "http request failed with non-200 status code")

	// invalid toml
	h.
		On("GET", "https://json.org/.well-known/diamnet.toml").
		ReturnJSON(http.StatusOK, map[string]string{"hello": "world"})
	stoml, err = c.GetDiamnetToml("json.org")

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "toml decode failed")
	}
}
