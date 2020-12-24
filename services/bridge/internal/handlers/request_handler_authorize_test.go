package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/http/httptest"
	"github.com/diamnet/go/txnbuild"
	"github.com/stretchr/testify/assert"

	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/bridge/internal/config"
	"github.com/diamnet/go/services/bridge/internal/mocks"
	"github.com/diamnet/go/services/bridge/internal/test"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols/bridge"
)

func TestRequestHandlerAuthorize(t *testing.T) {

	mockTS := new(mocks.MockTransactionSubmitter)

	config := config.Config{
		Assets: []protocols.Asset{
			{Code: "USD", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
			{Code: "EUR", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
		},
		Accounts: config.Accounts{
			IssuingAccountID: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR",
			// GBQXA3ABGQGTCLEVZIUTDRWWJOQD5LSAEDZAG7GMOGD2HBLWONGUVO4I
			AuthorizingSeed: "SC37TBSIAYKIDQ6GTGLT2HSORLIHZQHBXVFI5P5K4Q5TSHRTRBK3UNWG",
		},
	}

	requestHandler := RequestHandler{Config: &config, TransactionSubmitter: mockTS}
	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.Authorize))
	defer testServer.Close()

	// Invalid Account ID
	accountID := "GD3YBOYIUVLU"
	assetCode := "USD"

	statusCode, response := mocks.GetResponse(testServer, url.Values{"account_id": {accountID}, "asset_code": {assetCode}})
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected := test.StringToJSONMap(`{
				  "code": "invalid_parameter",
				  "message": "Invalid parameter.",
				  "data": {
				    "name": "AccountID"
				  }
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// no asset code
	accountID = "GCOGCYU77DLEVYCXDQM7F32M5PCKES6VU3Z5GURF6U6OA5LFOVTRYPOX"
	assetCode = ""

	statusCode, response = mocks.GetResponse(testServer, url.Values{"account_id": {accountID}, "asset_code": {assetCode}})
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
				  "code": "missing_parameter",
				  "message": "Required parameter is missing.",
				  "data": {
				    "name": "AssetCode"
				  }
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// Valid parameters
	accountID = "GCOGCYU77DLEVYCXDQM7F32M5PCKES6VU3Z5GURF6U6OA5LFOVTRYPOX"
	assetCode = "USD"

	allowTrustOp := bridge.AllowTrustOperationBody{
		Source:    &config.Accounts.IssuingAccountID,
		Authorize: true,
		Trustor:   accountID,
		AssetCode: assetCode,
	}

	operationBuilder := allowTrustOp.Build()

	// tx failure
	mockTS.On(
		"SubmitTransaction",
		(*string)(nil),
		config.Accounts.AuthorizingSeed,
		[]txnbuild.Operation{operationBuilder},
		nil,
	).Return(hProtocol.TransactionSuccess{},
		errors.New("Error sending transaction"),
	).Once()

	statusCode, response = mocks.GetResponse(testServer, url.Values{"account_id": {accountID}, "asset_code": {assetCode}})
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 500, statusCode)
	expected = test.StringToJSONMap(`{
					  "code": "internal_server_error",
					  "message": "Internal Server Error, please try again."
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	mockTS.AssertExpectations(t)

	var ledger int32
	ledger = 100
	expectedSubmitResponse := hProtocol.TransactionSuccess{
		Ledger: ledger,
	}

	mockTS.On(
		"SubmitTransaction",
		(*string)(nil),
		config.Accounts.AuthorizingSeed,
		[]txnbuild.Operation{operationBuilder},
		nil,
	).Return(expectedSubmitResponse, nil).Once()
	statusCode, response = mocks.GetResponse(testServer, url.Values{"account_id": {accountID}, "asset_code": {assetCode}})
	var actualSubmitTransactionResponse hProtocol.TransactionSuccess
	json.Unmarshal(response, &actualSubmitTransactionResponse)

	assert.Equal(t, 200, statusCode)
	assert.Equal(t, expectedSubmitResponse, actualSubmitTransactionResponse)
	mockTS.AssertExpectations(t)
}
