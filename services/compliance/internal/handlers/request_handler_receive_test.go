package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/hcnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"

	"github.com/hcnet/go/services/compliance/internal/config"
	"github.com/hcnet/go/services/compliance/internal/db"
	"github.com/hcnet/go/services/compliance/internal/mocks"
	callback "github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
)

func TestRequestHandlerReceive(t *testing.T) {
	var rhconfig = &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
	}

	var mockHTTPClient = new(mocks.MockHTTPClient)
	var mockDatabase = new(mocks.MockDatabase)
	var mockFederationResolver = new(mocks.MockFederationResolver)
	var mockSignerVerifier = new(mocks.MockSignerVerifier)
	var mockHcNettomlResolver = new(mocks.MockHcNettomlResolver)

	requestHandler := RequestHandler{
		Config:                  rhconfig,
		Client:                  mockHTTPClient,
		Database:                mockDatabase,
		FederationResolver:      mockFederationResolver,
		SignatureSignerVerifier: mockSignerVerifier,
		HcNetTomlResolver:     mockHcNettomlResolver,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerReceive))
	defer testServer.Close()

	// When memo not found, returns error
	memo := "907ba78b4545338d3539683e63ecb51cf51c10adc9dabd86e92bd52339f298b9"
	params := url.Values{"memo": {memo}}

	mockDatabase.On("GetAuthorizedTransactionByMemo", memo).Return(nil, nil).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 404, statusCode)
	errString, err := callback.TransactionNotFoundError.Marshal()
	assert.Nil(t, err)
	assert.Equal(t, errString, []byte(responseString))

	// When memo is found, return preimage
	memo = "bcc649cfdb8cc557053da67df7e7fcb740dcf7f721cebe1f2082597ad0d5e7d8"
	params = url.Values{"memo": {memo}}

	authorizedTransaction := db.AuthorizedTransaction{
		Memo: memo,
		Data: "hello world",
	}

	mockDatabase.On("GetAuthorizedTransactionByMemo", memo).Return(
		&authorizedTransaction,
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, "{\n  \"data\": \"hello world\"\n}", responseString)

}
