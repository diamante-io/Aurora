package handlers

import (
	"net/http"
	"testing"

	"github.com/goji/httpauth"
	"github.com/hcnet/go/services/compliance/internal/config"
	"github.com/hcnet/go/services/compliance/internal/mocks"
	"github.com/hcnet/go/support/http/httptest"
)

func TestRequestHandlerTxStatus(t *testing.T) {
	txStatusAuth := config.TxStatusAuth{
		Username: "username",
		Password: "password",
	}

	rhconfig := &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
		TxStatusAuth: &txStatusAuth,
	}

	mockHTTPClient := new(mocks.MockHTTPClient)
	mockDatabase := new(mocks.MockDatabase)
	mockFederationResolver := new(mocks.MockFederationResolver)
	mockSignerVerifier := new(mocks.MockSignerVerifier)
	mockHcNettomlResolver := new(mocks.MockHcNettomlResolver)

	requestHandler := RequestHandler{
		Config:                  rhconfig,
		Client:                  mockHTTPClient,
		Database:                mockDatabase,
		FederationResolver:      mockFederationResolver,
		SignatureSignerVerifier: mockSignerVerifier,
		HcNetTomlResolver:     mockHcNettomlResolver,
	}
	testServer := httptest.NewServer(t, httpauth.SimpleBasicAuth(rhconfig.TxStatusAuth.Username,
		rhconfig.TxStatusAuth.Password)(http.HandlerFunc(requestHandler.HandlerTxStatus)))
	// testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerTxStatus))
	defer testServer.Close()

	// It returns unauthorized when no auth
	testServer.GET("/tx_status").
		WithQuery("id", "123").
		Expect().
		Status(http.StatusUnauthorized)

	// It returns unauthorized when bad auth
	testServer.GET("/tx_status").
		WithBasicAuth("username", "wrong_password").
		Expect().
		Status(http.StatusUnauthorized)

		// it returns bad request when no parameter
	testServer.GET("/tx_status").
		WithBasicAuth("username", "password").
		Expect().
		Status(http.StatusBadRequest)

		// it returns unknown when no tx_status endpoint in config
	testServer.GET("/tx_status").
		WithBasicAuth("username", "password").
		WithQuery("id", "123").
		Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status":"unknown"}` + "\n")

	// it returns unknown when valid endpoint returns bad request
	rhconfig.Callbacks = config.Callbacks{
		TxStatus: "http://tx_status",
	}
	txid := "abc123"

	mockHTTPClient.On(
		"Get",
		"http://tx_status?id="+txid,
	).Return(
		mocks.BuildHTTPResponse(400, "badrequest"),
		nil,
	).Once()

	testServer.GET("/tx_status").
		WithBasicAuth("username", "password").
		WithQuery("id", txid).
		Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status":"unknown"}` + "\n")

	// it returns unknown when valid endpoint returns empty data
	mockHTTPClient.On(
		"Get",
		"http://tx_status?id="+txid,
	).Return(
		mocks.BuildHTTPResponse(200, "{}"),
		nil,
	).Once()

	testServer.GET("/tx_status").
		WithBasicAuth("username", "password").
		WithQuery("id", txid).
		Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status":"unknown"}` + "\n")

	// it returns response from valid endpoint with data
	mockHTTPClient.On(
		"Get",
		"http://tx_status?id="+txid,
	).Return(
		mocks.BuildHTTPResponse(200, `{"status":"delivered","msg":"cash paid"}`),
		nil,
	).Once()

	testServer.GET("/tx_status").
		WithBasicAuth("username", "password").
		WithQuery("id", txid).
		Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status":"delivered","msg":"cash paid"}` + "\n")
}
