package hcnettoml

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hcnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientURL(t *testing.T) {
	//HACK:  we're testing an internal method rather than setting up a http client
	//mock.

	c := &Client{UseHTTP: false}
	assert.Equal(t, "https://hcnet.org/.well-known/hcnet.toml", c.url("hcnet.org"))

	c = &Client{UseHTTP: true}
	assert.Equal(t, "http://hcnet.org/.well-known/hcnet.toml", c.url("hcnet.org"))
}

func TestClient(t *testing.T) {
	h := httptest.NewClient()
	c := &Client{HTTP: h}

	// happy path
	h.
		On("GET", "https://hcnet.org/.well-known/hcnet.toml").
		ReturnString(http.StatusOK,
			`FEDERATION_SERVER="https://localhost/federation"`,
		)
	stoml, err := c.GetHcNetToml("hcnet.org")
	require.NoError(t, err)
	assert.Equal(t, "https://localhost/federation", stoml.FederationServer)

	// hcnet.toml exceeds limit
	h.
		On("GET", "https://toobig.org/.well-known/hcnet.toml").
		ReturnString(http.StatusOK,
			`FEDERATION_SERVER="https://localhost/federation`+strings.Repeat("0", HcNetTomlMaxSize)+`"`,
		)
	stoml, err = c.GetHcNetToml("toobig.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "hcnet.toml response exceeds")
	}

	// not found
	h.
		On("GET", "https://missing.org/.well-known/hcnet.toml").
		ReturnNotFound()
	stoml, err = c.GetHcNetToml("missing.org")
	assert.EqualError(t, err, "http request failed with non-200 status code")

	// invalid toml
	h.
		On("GET", "https://json.org/.well-known/hcnet.toml").
		ReturnJSON(http.StatusOK, map[string]string{"hello": "world"})
	stoml, err = c.GetHcNetToml("json.org")

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "toml decode failed")
	}
}
