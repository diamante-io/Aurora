package federation

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/clients/hcnettoml"
	"github.com/hcnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
)

func TestLookupByAddress(t *testing.T) {
	hmock := httptest.NewClient()
	tomlmock := &hcnettoml.MockClient{}
	c := &Client{HcNetTOML: tomlmock, HTTP: hmock}

	// happy path - string integer
	tomlmock.On("GetHcNetToml", "hcnet.org").Return(&hcnettoml.Response{
		FederationServer: "https://hcnet.org/federation",
	}, nil)
	hmock.On("GET", "https://hcnet.org/federation").
		ReturnJSON(http.StatusOK, map[string]string{
			"hcnet_address": "scott*hcnet.org",
			"account_id":      "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C",
			"memo_type":       "id",
			"memo":            "123",
		})
	resp, err := c.LookupByAddress("scott*hcnet.org")

	if assert.NoError(t, err) {
		assert.Equal(t, "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C", resp.AccountID)
		assert.Equal(t, "id", resp.MemoType)
		assert.Equal(t, "123", resp.Memo.String())
	}

	// happy path - integer
	tomlmock.On("GetHcNetToml", "hcnet.org").Return(&hcnettoml.Response{
		FederationServer: "https://hcnet.org/federation",
	}, nil)
	hmock.On("GET", "https://hcnet.org/federation").
		ReturnJSON(http.StatusOK, map[string]interface{}{
			"hcnet_address": "scott*hcnet.org",
			"account_id":      "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C",
			"memo_type":       "id",
			"memo":            123,
		})
	resp, err = c.LookupByAddress("scott*hcnet.org")

	if assert.NoError(t, err) {
		assert.Equal(t, "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C", resp.AccountID)
		assert.Equal(t, "id", resp.MemoType)
		assert.Equal(t, "123", resp.Memo.String())
	}

	// happy path - string
	tomlmock.On("GetHcNetToml", "hcnet.org").Return(&hcnettoml.Response{
		FederationServer: "https://hcnet.org/federation",
	}, nil)
	hmock.On("GET", "https://hcnet.org/federation").
		ReturnJSON(http.StatusOK, map[string]interface{}{
			"hcnet_address": "scott*hcnet.org",
			"account_id":      "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C",
			"memo_type":       "text",
			"memo":            "testing",
		})
	resp, err = c.LookupByAddress("scott*hcnet.org")

	if assert.NoError(t, err) {
		assert.Equal(t, "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C", resp.AccountID)
		assert.Equal(t, "text", resp.MemoType)
		assert.Equal(t, "testing", resp.Memo.String())
	}

	// response exceeds limit
	tomlmock.On("GetHcNetToml", "toobig.org").Return(&hcnettoml.Response{
		FederationServer: "https://toobig.org/federation",
	}, nil)
	hmock.On("GET", "https://toobig.org/federation").
		ReturnJSON(http.StatusOK, map[string]string{
			"hcnet_address": strings.Repeat("0", FederationResponseMaxSize) + "*hcnet.org",
			"account_id":      "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C",
			"memo_type":       "id",
			"memo":            "123",
		})
	_, err = c.LookupByAddress("response*toobig.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "federation response exceeds")
	}

	// failed toml resolution
	tomlmock.On("GetHcNetToml", "missing.org").Return(
		(*hcnettoml.Response)(nil),
		errors.New("toml failed"),
	)
	_, err = c.LookupByAddress("scott*missing.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "toml failed")
	}

	// 404 federation response
	tomlmock.On("GetHcNetToml", "404.org").Return(&hcnettoml.Response{
		FederationServer: "https://404.org/federation",
	}, nil)
	hmock.On("GET", "https://404.org/federation").ReturnNotFound()
	_, err = c.LookupByAddress("scott*404.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed with (404)")
	}

	// connection error on federation response
	tomlmock.On("GetHcNetToml", "error.org").Return(&hcnettoml.Response{
		FederationServer: "https://error.org/federation",
	}, nil)
	hmock.On("GET", "https://error.org/federation").ReturnError("kaboom!")
	_, err = c.LookupByAddress("scott*error.org")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "kaboom!")
	}
}

func TestLookupByID(t *testing.T) {
	auroraMock := &hc.MockClient{}
	client := &Client{Aurora: auroraMock}

	auroraMock.On("HomeDomainForAccount", "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C").
		Return("", errors.New("homedomain not set"))

	// an account without a homedomain set fails
	_, err := client.LookupByAccountID("GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C")
	assert.Error(t, err)
	assert.Equal(t, "get homedomain failed: homedomain not set", err.Error())
}

func TestForwardRequest(t *testing.T) {
	hmock := httptest.NewClient()
	tomlmock := &hcnettoml.MockClient{}
	c := &Client{HcNetTOML: tomlmock, HTTP: hmock}

	// happy path - string integer
	tomlmock.On("GetHcNetToml", "hcnet.org").Return(&hcnettoml.Response{
		FederationServer: "https://hcnet.org/federation",
	}, nil)
	hmock.On("GET", "https://hcnet.org/federation").
		ReturnJSON(http.StatusOK, map[string]string{
			"account_id": "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C",
			"memo_type":  "id",
			"memo":       "123",
		})
	fields := url.Values{}
	fields.Add("federation_type", "bank_account")
	fields.Add("swift", "BOPBPHMM")
	fields.Add("acct", "2382376")
	resp, err := c.ForwardRequest("hcnet.org", fields)

	if assert.NoError(t, err) {
		assert.Equal(t, "GASTNVNLHVR3NFO3QACMHCJT3JUSIV4NBXDHDO4VTPDTNN65W3B2766C", resp.AccountID)
		assert.Equal(t, "id", resp.MemoType)
		assert.Equal(t, "123", resp.Memo.String())
	}
}

func Test_url(t *testing.T) {
	c := &Client{}

	// forward requests
	qstr := url.Values{}
	qstr.Add("type", "forward")
	qstr.Add("federation_type", "bank_account")
	qstr.Add("swift", "BOPBPHMM")
	qstr.Add("acct", "2382376")
	furl := c.url("https://hcnet.org/federation", qstr)
	assert.Equal(t, "https://hcnet.org/federation?acct=2382376&federation_type=bank_account&swift=BOPBPHMM&type=forward", furl)

	// regression: ensure that query is properly URI encoded
	qstr = url.Values{}
	qstr.Add("type", "q")
	qstr.Add("q", "scott+receiver1@hcnet.org*hcnet.org")
	furl = c.url("", qstr)
	assert.Equal(t, "?q=scott%2Breceiver1%40hcnet.org%2Ahcnet.org&type=q", furl)
}
