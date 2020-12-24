package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	hc "github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/protocols/federation"
	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/bridge/internal/config"
	"github.com/diamnet/go/services/bridge/internal/mocks"
	"github.com/diamnet/go/services/bridge/internal/test"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/http/httptest"
	"github.com/diamnet/go/txnbuild"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var phconfig = &config.Config{
	NetworkPassphrase: "Test SDF Network ; September 2015",
	Compliance:        "http://compliance",
	Assets: []protocols.Asset{
		{Code: "USD", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
		{Code: "EUR", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
	},
	Accounts: config.Accounts{
		// GAHA6GRCLCCN7XE2NEEUDSIVOFBOQ6GLSYXVLYCJXJKLPMDR5XB5XZZJ
		BaseSeed: "SBKKWO3ZVDDEHDJILGHPHCJCFD2GNUAYIUDMRAS326HLUEQ7ZFXWIGQK",
	},
}
var mockAurora = new(hc.MockClient)
var mockDatabase = new(mocks.MockDatabase)
var mockHTTPClient = new(mocks.MockHTTPClient)
var mockTransactionSubmitter = new(mocks.MockTransactionSubmitter)
var mockFederationResolver = new(mocks.MockFederationResolver)
var mockDiamNettomlResolver = new(mocks.MockDiamNettomlResolver)

func TestRequestHandlerPaymentInvalidParameter(t *testing.T) {
	requestHandler := RequestHandler{
		Config:               phconfig,
		Client:               mockHTTPClient,
		Aurora:              mockAurora,
		Database:             mockDatabase,
		TransactionSubmitter: mockTransactionSubmitter,
		FederationResolver:   mockFederationResolver,
		DiamNetTomlResolver:  mockDiamNettomlResolver,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.Payment))
	defer testServer.Close()

	// When source is invalid it should return error
	params := url.Values{
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX43"},
		"destination": {"GBABZMS7MEDWKWSHOMUKAWGIOE5UA4XLVPUHRHVMUW2DUVEZXLH5OIET"},
		"amount":      {"20.0"},
	}

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected := test.StringToJSONMap(`{
		"code": "invalid_parameter",
		"message": "Invalid parameter.",
		"data": {
			"name": "Source"
		}
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When destination is invalid it should return error
	params = url.Values{
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination": {"GD3YBOYIUVLU"},
		"amount":      {"20.0"},
	}
	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
		"code": "invalid_parameter",
		"message": "Invalid parameter.",
		"data": {
			"name": "Destination"
		}
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When amount is invalid it should return error
	params = url.Values{
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination": {"GBABZMS7MEDWKWSHOMUKAWGIOE5UA4XLVPUHRHVMUW2DUVEZXLH5OIET"},
	}
	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
		"code": "missing_parameter",
		"message": "Required parameter is missing.",
		"data": {
			"name": "Amount"
		}
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When asset issuer is invalid it should return error
	params = url.Values{
		"source":       {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination":  {"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		"asset_code":   {"USD"},
		"asset_issuer": {"GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOX"},
		"amount":       {"100.0"},
	}

	mockFederationResolver.On(
		"LookupByAddress",
		"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632",
	).Return(
		&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
		"code": "invalid_parameter",
		"message": "Invalid parameter.",
		"data": {
			"name": "AssetIssuer"
		}
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When asset code is invalid it should return error
	params = url.Values{
		"source":       {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination":  {"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		"asset_code":   {"USD01234567890"},
		"asset_issuer": {"GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
		"amount":       {"100.0"},
	}

	mockFederationResolver.On(
		"LookupByAddress",
		"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632",
	).Return(
		&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		nil,
	).Once()

	// Should return an error
	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
  "code": "invalid_parameter",
		"message": "Invalid parameter.",
		"data": {
			"name": "AssetCode"
		}
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

}

func TestRequestHandlerPaymentErrorResponse(t *testing.T) {
	requestHandler := RequestHandler{
		Config:               phconfig,
		Client:               mockHTTPClient,
		Aurora:              mockAurora,
		Database:             mockDatabase,
		TransactionSubmitter: mockTransactionSubmitter,
		FederationResolver:   mockFederationResolver,
		DiamNetTomlResolver:  mockDiamNettomlResolver,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.Payment))
	defer testServer.Close()

	// When federation response is an error it should return error
	params := url.Values{
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination": {"bob*diamnet.org"},
		"amount":      {"20.0"},
	}

	mockFederationResolver.On(
		"LookupByAddress",
		"bob*diamnet.org",
	).Return(
		&federation.NameResponse{},
		errors.New("diamnet.toml response status code indicates error"),
	).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected := test.StringToJSONMap(`{
		"code": "cannot_resolve_destination",
		"message": "Cannot resolve federated DiamNet address."
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When using foward destination with federation error. It should return error
	params = url.Values{
		"source":                      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"forward_destination[domain]": {"diamnet.org"},
		"forward_destination[fields][federation_type]": {"bank_account"},
		"forward_destination[fields][swift]":           {"BOPBPHMM"},
		"forward_destination[fields][acct]":            {"2382376"},
		"amount":                                       {"20.0"},
	}

	mockFederationResolver.On(
		"ForwardRequest",
		"diamnet.org",
		url.Values{
			"federation_type": {"bank_account"},
			"swift":           {"BOPBPHMM"},
			"acct":            {"2382376"},
		},
	).Return(
		&federation.NameResponse{},
		errors.New("diamnet.toml response status code indicates error"),
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
		"code": "cannot_resolve_destination",
		"message": "Cannot resolve federated DiamNet address."
	}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

}

func TestRequestHandlerPaymentSuccessResponse(t *testing.T) {
	requestHandler := RequestHandler{
		Config:               phconfig,
		Client:               mockHTTPClient,
		Aurora:              mockAurora,
		Database:             mockDatabase,
		TransactionSubmitter: mockTransactionSubmitter,
		FederationResolver:   mockFederationResolver,
		DiamNetTomlResolver:  mockDiamNettomlResolver,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.Payment))
	defer testServer.Close()

	// When destination is public key it should return success
	params := url.Values{
		// GBKGH7QZVCZ2ZA5OUGZSTHFNXTBHL3MPCKSCBJUAQODGPMWP7OMMRKDW
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination": {"GAPCT362RATBUJ37RN2MOKQIZLHSJMO33MMCSRUXTTHIGVDYWOFG5HDS"},
		"amount":      {"20.0"},
	}

	accountRequest := hc.AccountRequest{AccountID: "GAPCT362RATBUJ37RN2MOKQIZLHSJMO33MMCSRUXTTHIGVDYWOFG5HDS"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	var ledger int32
	ledger = 1988728
	auroraResponse := hProtocol.TransactionSuccess{
		Hash:   "6a0049b44e0d0341bd52f131c74383e6ccd2b74b92c829c990994d24bbfcfa7a",
		Ledger: ledger,
	}

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42",
		mock.AnythingOfType("[]txnbuild.Operation"),
		nil,
	).Return(auroraResponse, nil).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))

	assert.Equal(t, 200, statusCode)
	expected := test.StringToJSONMap(`{
					  "hash": "6a0049b44e0d0341bd52f131c74383e6ccd2b74b92c829c990994d24bbfcfa7a",
						"ledger": 1988728
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

	// When destination is a diamnet address it should return success
	params = url.Values{
		"source":      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"destination": {"bob*diamnet.org"},
		"amount":      {"20.0"},
	}

	mockFederationResolver.On(
		"LookupByAddress",
		"bob*diamnet.org",
	).Return(
		&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		nil,
	).Once()

	accountRequest = hc.AccountRequest{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42",
		mock.AnythingOfType("[]txnbuild.Operation"),
		nil,
	).Return(auroraResponse, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
					  "hash": "6a0049b44e0d0341bd52f131c74383e6ccd2b74b92c829c990994d24bbfcfa7a",
						"ledger": 1988728
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

	// When federation response has memo it should return success
	mockFederationResolver.On(
		"LookupByAddress",
		"bob*diamnet.org",
	).Return(
		&federation.NameResponse{
			AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632",
			MemoType:  "text",
			Memo:      federation.Memo{Value: "125"},
		},
		nil,
	).Once()

	accountRequest = hc.AccountRequest{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	ledger = 1988728
	auroraResponse = hProtocol.TransactionSuccess{
		Hash:   "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
		Ledger: ledger,
	}

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42",
		mock.AnythingOfType("[]txnbuild.Operation"),
		txnbuild.MemoText("125"),
	).Return(auroraResponse, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))

	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
					  "hash": "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
						"ledger": 1988728
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

	// When using foward destination with memo. It should return success
	params = url.Values{
		"source":                      {"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42"},
		"forward_destination[domain]": {"diamnet.org"},
		"forward_destination[fields][federation_type]": {"bank_account"},
		"forward_destination[fields][swift]":           {"BOPBPHMM"},
		"forward_destination[fields][acct]":            {"2382376"},
		"amount":                                       {"20.0"},
	}

	mockFederationResolver.On(
		"ForwardRequest",
		"diamnet.org",
		url.Values{
			"federation_type": {"bank_account"},
			"swift":           {"BOPBPHMM"},
			"acct":            {"2382376"},
		},
	).Return(
		&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632",
			MemoType: "text",
			Memo:     federation.Memo{Value: "125"}},
		nil,
	).Once()

	accountRequest = hc.AccountRequest{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	ledger = 1988728
	auroraResponse = hProtocol.TransactionSuccess{
		Hash:   "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
		Ledger: ledger,
	}

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42",
		mock.AnythingOfType("[]txnbuild.Operation"),
		txnbuild.MemoText("125"),
	).Return(auroraResponse, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
					  "hash": "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
						"ledger": 1988728
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

	// When using foward destination without memo. It should return success
	mockFederationResolver.On(
		"ForwardRequest",
		"diamnet.org",
		url.Values{
			"federation_type": {"bank_account"},
			"swift":           {"BOPBPHMM"},
			"acct":            {"2382376"},
		},
	).Return(
		&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		nil,
	).Once()

	accountRequest = hc.AccountRequest{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	ledger = 1988728
	auroraResponse = hProtocol.TransactionSuccess{
		Hash:   "6a0049b44e0d0341bd52f131c74383e6ccd2b74b92c829c990994d24bbfcfa7a",
		Ledger: ledger,
	}

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SDRAS7XIQNX25UDCCX725R4EYGBFYGJE4HJ2A3DFCWJIHMRSMS7CXX42",
		mock.AnythingOfType("[]txnbuild.Operation"),
		nil,
	).Return(auroraResponse, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))

	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
					  "hash": "6a0049b44e0d0341bd52f131c74383e6ccd2b74b92c829c990994d24bbfcfa7a",
						"ledger": 1988728
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

	// When no source is specified, it should use the base seed
	params = url.Values{
		"destination":  {"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
		"amount":       {"20"},
		"asset_code":   {"USD"},
		"asset_issuer": {"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
	}

	// mockFederationResolver.On(
	// 	"LookupByAddress",
	// 	"GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632",
	// ).Return(
	// 	&federation.NameResponse{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"},
	// 	nil,
	// ).Once()

	accountRequest = hc.AccountRequest{AccountID: "GDSIKW43UA6JTOA47WVEBCZ4MYC74M3GNKNXTVDXFHXYYTNO5GGVN632"}
	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(hProtocol.Account{}, nil).Once()

	ledger = 1988728
	auroraResponse = hProtocol.TransactionSuccess{
		Hash:   "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
		Ledger: ledger,
	}

	mockTransactionSubmitter.On(
		"SubmitTransaction",
		mock.AnythingOfType("*string"),
		"SBKKWO3ZVDDEHDJILGHPHCJCFD2GNUAYIUDMRAS326HLUEQ7ZFXWIGQK",
		mock.AnythingOfType("[]txnbuild.Operation"),
		nil,
	).Return(auroraResponse, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
				  "hash": "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
				  "ledger": 1988728
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "envelope_xdr", "_links", "result_meta_xdr", "result_xdr"))

}
