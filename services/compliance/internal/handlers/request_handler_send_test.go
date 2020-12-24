package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/diamnet/go/clients/diamnettoml"
	"github.com/diamnet/go/protocols/compliance"
	"github.com/diamnet/go/protocols/federation"
	"github.com/diamnet/go/support/http/httptest"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/diamnet/go/services/compliance/internal/config"
	"github.com/diamnet/go/services/compliance/internal/db"
	"github.com/diamnet/go/services/compliance/internal/mocks"
	"github.com/diamnet/go/services/compliance/internal/test"
	"github.com/diamnet/go/txnbuild"
)

func TestRequestHandlerSendInvalidParams(t *testing.T) {
	var rhconfig = &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
		Callbacks: config.Callbacks{
			FetchInfo: "http://fetch_info",
		},
	}

	var mockHTTPClient = new(mocks.MockHTTPClient)
	var mockDatabase = new(mocks.MockDatabase)
	var mockFederationResolver = new(mocks.MockFederationResolver)
	var mockSignerVerifier = new(mocks.MockSignerVerifier)
	var mockDiamNettomlResolver = new(mocks.MockDiamNettomlResolver)
	var mockNonceGenerator = new(mocks.MockNonceGenerator)

	requestHandler := RequestHandler{
		Config:                  rhconfig,
		Client:                  mockHTTPClient,
		Database:                mockDatabase,
		FederationResolver:      mockFederationResolver,
		SignatureSignerVerifier: mockSignerVerifier,
		DiamNetTomlResolver:     mockDiamNettomlResolver,
		NonceGenerator:          mockNonceGenerator,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerSend))
	defer testServer.Close()

	// When id parameter is missing, return error
	params := url.Values{
		// "id":           {"id"},
		"source":      {"GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
		"sender":      {"alice*diamnet.org"}, // GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD
		"destination": {"bob*diamnet.org"},   // GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE
		"amount":      {"20"},
	}

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected := test.StringToJSONMap(`{
			  "code": "missing_parameter",
			  "message": "Required parameter is missing.",
			  "data": {
			    "name": "ID"
			  }
			}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when source parameter is missing
	params = url.Values{
		"id": {"id"},
		// "source":      {"bad"},
		"sender":      {"alice*diamnet.org"}, // GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD
		"destination": {"bob*diamnet.org"},   // GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE
		"amount":      {"20"},
	}

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
			  "code": "missing_parameter",
			  "message": "Required parameter is missing.",
			  "data": {
			    "name": "Source"
			  }
			}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When source param is invalid, return error
	params = url.Values{
		"id":           {"id"},
		"source":       {"bad"},
		"sender":       {"alice*diamnet.org"}, // GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD
		"destination":  {"bob*diamnet.org"},   // GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE
		"amount":       {"20"},
		"asset_code":   {"USD"},
		"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		"extra_memo":   {"hello world"},
	}

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
			  "code": "invalid_parameter",
			  "message": "Invalid parameter.",
			  "data": {
			    "name": "Source"
			  }
			}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))
}

func TestRequestHandlerSendValidParams(t *testing.T) {
	var rhconfig = &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
		Callbacks: config.Callbacks{
			FetchInfo: "http://fetch_info",
		},
	}

	var mockHTTPClient = new(mocks.MockHTTPClient)
	var mockDatabase = new(mocks.MockDatabase)
	var mockFederationResolver = new(mocks.MockFederationResolver)
	var mockSignerVerifier = new(mocks.MockSignerVerifier)
	var mockDiamNettomlResolver = new(mocks.MockDiamNettomlResolver)
	var mockNonceGenerator = new(mocks.MockNonceGenerator)

	requestHandler := RequestHandler{
		Config:                  rhconfig,
		Client:                  mockHTTPClient,
		Database:                mockDatabase,
		FederationResolver:      mockFederationResolver,
		SignatureSignerVerifier: mockSignerVerifier,
		DiamNetTomlResolver:     mockDiamNettomlResolver,
		NonceGenerator:          mockNonceGenerator,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerSend))
	defer testServer.Close()

	// When params are valid, it returns SendResponse when success (payment)
	params := url.Values{
		"id":           {"id"},
		"source":       {"GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
		"sender":       {"alice*diamnet.org"}, // GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD
		"destination":  {"bob*diamnet.org"},   // GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE
		"amount":       {"20"},
		"asset_code":   {"USD"},
		"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		"extra_memo":   {"hello world"},
	}

	mockDatabase.Mock.On("GetAuthData", "id").Return(nil, nil).Once()
	mockDatabase.On("InsertAuthData", mock.AnythingOfType("*db.AuthData")).Run(func(args mock.Arguments) {
		entity, ok := args.Get(0).(*db.AuthData)
		assert.True(t, ok, "Invalid conversion")
		assert.Equal(t, "id", entity.RequestID)
		assert.Equal(t, "diamnet.org", entity.Domain)
	}).Return(nil).Once()

	senderInfo := compliance.SenderInfo{FirstName: "John", LastName: "Doe"}
	senderInfoMap, err := senderInfo.Map()
	require.NoError(t, err)

	authServer := "https://acme.com/auth"

	mockFederationResolver.On(
		"LookupByAddress",
		"bob*diamnet.org",
	).Return(&federation.NameResponse{
		AccountID: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		MemoType:  "text",
		Memo:      federation.Memo{"bob"},
	}, nil).Once()

	mockDiamNettomlResolver.On(
		"GetDiamNetToml",
		"diamnet.org",
	).Return(&diamnettoml.Response{AuthServer: authServer}, nil).Once()

	attachment := compliance.Attachment{
		Nonce: "nonce",
		Transaction: compliance.Transaction{
			Route:      "bob",
			Note:       "",
			SenderInfo: senderInfoMap,
			Extra:      "hello world",
		},
	}

	attachmentJSON, err := attachment.Marshal()
	require.NoError(t, err)
	attachHash, err := attachment.Hash()
	require.NoError(t, err)

	txnOp := &txnbuild.Payment{
		Destination:   "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		Amount:        "20",
		Asset:         txnbuild.CreditAsset{Code: "USD", Issuer: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
	}

	tx := txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD", Sequence: int64(-1)},
		Operations:    []txnbuild.Operation{txnOp},
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       rhconfig.NetworkPassphrase,
		Memo:          txnbuild.MemoHash(attachHash),
	}

	err = tx.Build()
	require.NoError(t, err)

	err = tx.Sign()
	require.NoError(t, err)

	txeB64, err := tx.Base64()
	require.NoError(t, err)

	var txXDR xdr.TransactionEnvelope
	err = xdr.SafeUnmarshalBase64(txeB64, &txXDR)
	require.NoError(t, err)
	txB64, err := xdr.MarshalBase64(txXDR.Tx)
	require.NoError(t, err)

	authData := compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}

	authDataJSON, err := authData.Marshal()
	require.NoError(t, err)

	authRequest := compliance.AuthRequest{
		DataJSON:  string(authDataJSON),
		Signature: "8IGqaeusNOnoFzUYzAlW0sxXXyIOx8fxvEToUHGyTJ9bxdoXJPj6yopa3kEjkVdMIiO5fuzVluk/KTxlbROGDg==",
	}

	authResponse := compliance.AuthResponse{
		InfoStatus: compliance.AuthStatusOk,
		TxStatus:   compliance.AuthStatusOk,
	}

	authResponseJSON, err := authResponse.Marshal()
	require.NoError(t, err)

	mockHTTPClient.On(
		"PostForm",
		rhconfig.Callbacks.FetchInfo,
		url.Values{"address": {"alice*diamnet.org"}},
	).Return(
		mocks.BuildHTTPResponse(200, "{\"first_name\": \"John\", \"last_name\": \"Doe\"}"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		authServer,
		authRequest.ToURLValues(),
	).Return(
		mocks.BuildHTTPResponse(200, string(authResponseJSON)),
		nil,
	).Once()

	mockSignerVerifier.On(
		"Sign",
		rhconfig.Keys.SigningSeed,
		[]byte(authRequest.DataJSON),
	).Return(authRequest.Signature, nil).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected := test.StringToJSONMap(`{
				  "auth_response": {
				    "info_status": "ok",
				    "tx_status": "ok"
				  },
				  "transaction_xdr": "` + txB64 + `"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When params are valid, it returns SendResponse when success (path payment)
	params["send_max"] = []string{"100"}
	params["send_asset_code"] = []string{"USD"}
	params["send_asset_issuer"] = []string{"GBDOSO3K4JTGSWJSIHXAOFIBMAABVM3YK3FI6VJPKIHHM56XAFIUCGD6"}

	// Native
	params["path[0][asset_code]"] = []string{""}
	params["path[0][asset_issuer]"] = []string{""}
	// Credit
	params["path[1][asset_code]"] = []string{"EUR"}
	params["path[1][asset_issuer]"] = []string{"GAF3PBFQLH57KPECN4GRGHU5NUZ3XXKYYWLOTBIRJMBYHPUBWANIUCZU"}

	mockDatabase.Mock.On("GetAuthData", "id").Return(nil, nil).Once()
	mockDatabase.On("InsertAuthData", mock.AnythingOfType("*db.AuthData")).Run(func(args mock.Arguments) {
		entity, ok := args.Get(0).(*db.AuthData)
		assert.True(t, ok, "Invalid conversion")
		assert.Equal(t, "id", entity.RequestID)
		assert.Equal(t, "diamnet.org", entity.Domain)
	}).Return(nil).Once()

	senderInfo = compliance.SenderInfo{FirstName: "John", LastName: "Doe"}
	senderInfoMap, err = senderInfo.Map()
	require.NoError(t, err)

	authServer = "https://acme.com/auth"

	mockFederationResolver.On(
		"LookupByAddress",
		"bob*diamnet.org",
	).Return(&federation.NameResponse{
		AccountID: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		MemoType:  "text",
		Memo:      federation.Memo{"bob"},
	}, nil).Once()

	mockDiamNettomlResolver.On(
		"GetDiamNetToml",
		"diamnet.org",
	).Return(&diamnettoml.Response{AuthServer: authServer}, nil).Once()

	attachment = compliance.Attachment{
		Nonce: "nonce",
		Transaction: compliance.Transaction{
			Route:      "bob",
			Note:       "",
			SenderInfo: senderInfoMap,
			Extra:      "hello world",
		},
	}

	attachmentJSON, err = attachment.Marshal()
	require.NoError(t, err)
	attachHash, err = attachment.Hash()
	require.NoError(t, err)

	pathPaymentOp := &txnbuild.PathPayment{
		SendAsset:     txnbuild.CreditAsset{Code: "USD", Issuer: "GBDOSO3K4JTGSWJSIHXAOFIBMAABVM3YK3FI6VJPKIHHM56XAFIUCGD6"},
		SendMax:       "100",
		Destination:   "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		DestAmount:    "20",
		DestAsset:     txnbuild.CreditAsset{Code: "USD", Issuer: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
		Path:          []txnbuild.Asset{txnbuild.NativeAsset{}, txnbuild.CreditAsset{Code: "EUR", Issuer: "GAF3PBFQLH57KPECN4GRGHU5NUZ3XXKYYWLOTBIRJMBYHPUBWANIUCZU"}},
	}

	tx = txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD", Sequence: int64(-1)},
		Operations:    []txnbuild.Operation{pathPaymentOp},
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       rhconfig.NetworkPassphrase,
		Memo:          txnbuild.MemoHash(attachHash),
	}

	err = tx.Build()
	require.NoError(t, err)

	err = tx.Sign()
	require.NoError(t, err)

	txeB64, err = tx.Base64()
	require.NoError(t, err)

	err = xdr.SafeUnmarshalBase64(txeB64, &txXDR)
	require.NoError(t, err)
	txB64, err = xdr.MarshalBase64(txXDR.Tx)
	require.NoError(t, err)

	authData = compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}

	authDataJSON, err = authData.Marshal()
	require.NoError(t, err)

	authRequest = compliance.AuthRequest{
		DataJSON:  string(authDataJSON),
		Signature: "8IGqaeusNOnoFzUYzAlW0sxXXyIOx8fxvEToUHGyTJ9bxdoXJPj6yopa3kEjkVdMIiO5fuzVluk/KTxlbROGDg==",
	}

	authResponse = compliance.AuthResponse{
		InfoStatus: compliance.AuthStatusOk,
		TxStatus:   compliance.AuthStatusOk,
	}

	authResponseJSON, err = authResponse.Marshal()
	require.NoError(t, err)

	mockHTTPClient.On(
		"PostForm",
		rhconfig.Callbacks.FetchInfo,
		url.Values{"address": {"alice*diamnet.org"}},
	).Return(
		mocks.BuildHTTPResponse(200, "{\"first_name\": \"John\", \"last_name\": \"Doe\"}"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		authServer,
		authRequest.ToURLValues(),
	).Return(
		mocks.BuildHTTPResponse(200, string(authResponseJSON)),
		nil,
	).Once()

	mockSignerVerifier.On(
		"Sign",
		rhconfig.Keys.SigningSeed,
		[]byte(authRequest.DataJSON),
	).Return(authRequest.Signature, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
				  "auth_response": {
				    "info_status": "ok",
				    "tx_status": "ok"
				  },
				  "transaction_xdr": "` + txB64 + `"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When params are valid, it returns SendResponse when success (Forward request)
	params = url.Values{
		"id":           {"id"},
		"source":       {"GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
		"sender":       {"alice*diamnet.org"}, // GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD
		"destination":  {"bob*diamnet.org"},   // GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE
		"amount":       {"20"},
		"asset_code":   {"USD"},
		"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		"extra_memo":   {"hello world"},
	}

	params["forward_destination[domain]"] = []string{"diamnet.org"}
	params["forward_destination[fields][federation_type]"] = []string{"bank_account"}
	params["forward_destination[fields][swift]"] = []string{"BOPBPHMM"}
	params["forward_destination[fields][acct]"] = []string{"2382376"}

	mockDatabase.Mock.On("GetAuthData", "id").Return(nil, nil).Once()
	mockDatabase.On("InsertAuthData", mock.AnythingOfType("*db.AuthData")).Run(func(args mock.Arguments) {
		entity, ok := args.Get(0).(*db.AuthData)
		assert.True(t, ok, "Invalid conversion")
		assert.Equal(t, "id", entity.RequestID)
		assert.Equal(t, "diamnet.org", entity.Domain)
	}).Return(nil).Once()

	senderInfo = compliance.SenderInfo{FirstName: "John", LastName: "Doe"}
	senderInfoMap, err = senderInfo.Map()
	require.NoError(t, err)

	authServer = "https://acme.com/auth"

	mockFederationResolver.On(
		"ForwardRequest",
		"diamnet.org",
		url.Values{
			"federation_type": []string{"bank_account"},
			"swift":           []string{"BOPBPHMM"},
			"acct":            []string{"2382376"},
		},
	).Return(&federation.NameResponse{
		AccountID: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		MemoType:  "text",
		Memo:      federation.Memo{"bob"},
	}, nil).Once()

	mockDiamNettomlResolver.On(
		"GetDiamNetToml",
		"diamnet.org",
	).Return(&diamnettoml.Response{AuthServer: authServer}, nil).Once()

	attachment = compliance.Attachment{
		Nonce: "nonce",
		Transaction: compliance.Transaction{
			Route:      "bob",
			Note:       "",
			SenderInfo: senderInfoMap,
			Extra:      "hello world",
		},
	}

	attachmentJSON, err = attachment.Marshal()
	require.NoError(t, err)
	attachHash, err = attachment.Hash()
	require.NoError(t, err)

	txnOp = &txnbuild.Payment{
		Destination:   "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE",
		Amount:        "20",
		Asset:         txnbuild.CreditAsset{Code: "USD", Issuer: "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
	}

	tx = txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: "GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD", Sequence: int64(-1)},
		Operations:    []txnbuild.Operation{txnOp},
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       rhconfig.NetworkPassphrase,
		Memo:          txnbuild.MemoHash(attachHash),
	}

	err = tx.Build()
	require.NoError(t, err)

	err = tx.Sign()
	require.NoError(t, err)

	txeB64, err = tx.Base64()
	require.NoError(t, err)

	err = xdr.SafeUnmarshalBase64(txeB64, &txXDR)
	require.NoError(t, err)
	txB64, err = xdr.MarshalBase64(txXDR.Tx)
	require.NoError(t, err)

	authData = compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}

	authDataJSON, err = authData.Marshal()
	require.NoError(t, err)

	authRequest = compliance.AuthRequest{
		DataJSON:  string(authDataJSON),
		Signature: "8IGqaeusNOnoFzUYzAlW0sxXXyIOx8fxvEToUHGyTJ9bxdoXJPj6yopa3kEjkVdMIiO5fuzVluk/KTxlbROGDg==",
	}

	authResponse = compliance.AuthResponse{
		InfoStatus: compliance.AuthStatusOk,
		TxStatus:   compliance.AuthStatusOk,
	}

	authResponseJSON, err = authResponse.Marshal()
	require.NoError(t, err)

	mockHTTPClient.On(
		"PostForm",
		rhconfig.Callbacks.FetchInfo,
		url.Values{"address": {"alice*diamnet.org"}},
	).Return(
		mocks.BuildHTTPResponse(200, "{\"first_name\": \"John\", \"last_name\": \"Doe\"}"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		authServer,
		authRequest.ToURLValues(),
	).Return(
		mocks.BuildHTTPResponse(200, string(authResponseJSON)),
		nil,
	).Once()

	mockSignerVerifier.On(
		"Sign",
		rhconfig.Keys.SigningSeed,
		[]byte(authRequest.DataJSON),
	).Return(authRequest.Signature, nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
				  "auth_response": {
				    "info_status": "ok",
				    "tx_status": "ok"
				  },
				  "transaction_xdr": "` + txB64 + `"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

}
