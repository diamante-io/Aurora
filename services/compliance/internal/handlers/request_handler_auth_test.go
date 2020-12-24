package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/diamnet/go/clients/diamnettoml"
	"github.com/diamnet/go/protocols/compliance"
	"github.com/diamnet/go/support/http/httptest"
	"github.com/diamnet/go/txnbuild"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/diamnet/go/services/compliance/internal/config"
	"github.com/diamnet/go/services/compliance/internal/db"
	"github.com/diamnet/go/services/compliance/internal/mocks"
	"github.com/diamnet/go/services/compliance/internal/test"
)

func TestRequestHandlerAuthInvalidParams(t *testing.T) {
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

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerAuth))
	defer testServer.Close()

	// auth request (no sanctions check) When data param is missing, return error
	statusCode, response := mocks.GetResponse(testServer, url.Values{})
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected := test.StringToJSONMap(`{
					  "code": "invalid_parameter",
					  "message": "Invalid parameter."
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// auth request (no sanctions check) When data is invalid, return error
	params := url.Values{
		"data": {"hello world"},
		"sig":  {"bad sig"},
	}

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
					  "code": "invalid_parameter",
					  "message": "Invalid parameter."
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When signature is invalid
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	attachment := compliance.Attachment{}
	attachHash, err := attachment.Hash()
	require.NoError(t, err)
	attachmentJSON, err := attachment.Marshal()
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

	params = url.Values{
		"data": {string(authDataJSON)},
		"sig":  {"ACamNqa0dF8gf97URhFVKWSD7fmvZKc5At+8dCLM5ySR0HsHySF3G2WuwYP2nKjeqjKmu3U9Z3+u1P10w1KBCA=="},
	}

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(errors.New("Verify error")).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
  "code": "invalid_parameter",
  "message": "Invalid parameter.",
  "data": {
    "name": "sig"
  }
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

	// When sender's diamnet.toml does not contain signing key
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{}, nil).Once()

	attachHash = sha256.Sum256([]byte("{}"))
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
		AttachmentJSON: "{}",
	}

	authDataJSON, err = authData.Marshal()
	require.NoError(t, err)

	params = url.Values{
		"data": {string(authDataJSON)},
		"sig":  {"ACamNqa0dF8gf97URhFVKWSD7fmvZKc5At+8dCLM5ySR0HsHySF3G2WuwYP2nKjeqjKmu3U9Z3+u1P10w1KBCA=="},
	}

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
		  "code": "invalid_parameter",
		  "message": "Invalid parameter.",
		  "data": {
		    "name": "data.sender"
		  }
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString, "more_info"))

}

func TestRequestHandlerAuthValidParams(t *testing.T) {
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

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerAuth))
	defer testServer.Close()

	// When all params are valid
	attachment := compliance.Attachment{}
	attachHash, err := attachment.Hash()
	require.NoError(t, err)
	attachHashB64 := base64.StdEncoding.EncodeToString(attachHash[:])
	attachmentJSON, err := attachment.Marshal()
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

	txHash, err := tx.Hash()
	require.NoError(t, err)
	txHashHex := hex.EncodeToString(txHash[:])

	authData := compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}
	authDataJSON, err := authData.Marshal()
	require.NoError(t, err)

	params := url.Values{
		"data": {string(authDataJSON)},
		"sig":  {"ACamNqa0dF8gf97URhFVKWSD7fmvZKc5At+8dCLM5ySR0HsHySF3G2WuwYP2nKjeqjKmu3U9Z3+u1P10w1KBCA=="},
	}

	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	// It returns AuthResponse
	authorizedTransaction := &db.AuthorizedTransaction{
		TransactionID:  txHashHex,
		Memo:           attachHashB64,
		TransactionXdr: txB64,
		Data:           params["data"][0],
	}

	mockDatabase.On(
		"InsertAuthorizedTransaction",
		mock.AnythingOfType("*db.AuthorizedTransaction"),
	).Run(func(args mock.Arguments) {
		value := args.Get(0).(*db.AuthorizedTransaction)
		assert.Equal(t, authorizedTransaction.TransactionID, value.TransactionID)
		assert.Equal(t, authorizedTransaction.Memo, value.Memo)
		assert.Equal(t, authorizedTransaction.TransactionXdr, value.TransactionXdr)
		assert.WithinDuration(t, time.Now(), value.AuthorizedAt, 2*time.Second)
		assert.Equal(t, authorizedTransaction.Data, value.Data)
	}).Return(nil).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected := test.StringToJSONMap(`{
  "info_status": "ok",
  "tx_status": "ok"
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

}

func TestRequestHandlerAuthSanctionsCheck(t *testing.T) {
	var rhconfig = &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
		Callbacks: config.Callbacks{
			Sanctions: "http://sanctions",
			AskUser:   "http://ask_user",
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

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerAuth))
	defer testServer.Close()
	senderInfo := compliance.SenderInfo{FirstName: "John", LastName: "Doe"}
	senderInfoMap, err := senderInfo.Map()
	require.NoError(t, err)

	attachment := compliance.Attachment{
		Transaction: compliance.Transaction{
			Route:      "bob*acme.com",
			Note:       "Happy birthday",
			SenderInfo: senderInfoMap,
			Extra:      "extra",
		},
	}

	attachHash, err := attachment.Hash()
	require.NoError(t, err)
	attachHashB64 := base64.StdEncoding.EncodeToString(attachHash[:])

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

	txHash, err := tx.Hash()
	require.NoError(t, err)
	txHashHex := hex.EncodeToString(txHash[:])

	attachmentJSON, err := attachment.Marshal()
	require.NoError(t, err)

	senderInfoJSON, err := json.Marshal(attachment.Transaction.SenderInfo)
	require.NoError(t, err)

	// When all params are valid (NeedInfo = `false`)
	authData := compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}

	authDataJSON, err := authData.Marshal()
	require.NoError(t, err)

	params := url.Values{
		"data": {string(authDataJSON)},
		"sig":  {"Q2cQVOn/A+aOxrLLeUPwHmBm3LMvlfXN8tDHo4Oi6SxWWueMTDfRkC4XvRX4emLij+Npo7/GfrZ82CnT5yB5Dg=="},
	}

	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	// when sanctions server returns forbidden it returns tx_status `denied`

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{"sender": {string(senderInfoJSON)}},
	).Return(
		mocks.BuildHTTPResponse(403, "forbidden"),
		nil,
	).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 403, statusCode)
	expected := test.StringToJSONMap(`{
		  "info_status": "ok",
		  "tx_status": "denied"
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when sanctions server returns bad request it returns tx_status `error`
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{"sender": {string(senderInfoJSON)}},
	).Return(
		mocks.BuildHTTPResponse(400, "{\"error\": \"Invalid name\"}"),
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
  "info_status": "ok",
  "tx_status": "error",
  "error": "Invalid name"
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when sanctions server returns accepted it returns tx_status `pending`
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{"sender": {string(senderInfoJSON)}},
	).Return(
		mocks.BuildHTTPResponse(202, "pending"),
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 202, statusCode)
	expected = test.StringToJSONMap(`{
  "info_status": "ok",
  "tx_status": "pending",
  "pending": 600
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when sanctions server returns ok it returns tx_status `ok` and persists transaction
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{"sender": {string(senderInfoJSON)}},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	authorizedTransaction := &db.AuthorizedTransaction{
		TransactionID:  txHashHex,
		Memo:           attachHashB64,
		TransactionXdr: txB64,
		Data:           params["data"][0],
	}

	mockDatabase.On(
		"InsertAuthorizedTransaction",
		mock.AnythingOfType("*db.AuthorizedTransaction"),
	).Run(func(args mock.Arguments) {
		value := args.Get(0).(*db.AuthorizedTransaction)
		assert.Equal(t, authorizedTransaction.TransactionID, value.TransactionID)
		assert.Equal(t, authorizedTransaction.Memo, value.Memo)
		assert.Equal(t, authorizedTransaction.TransactionXdr, value.TransactionXdr)
		assert.WithinDuration(t, time.Now(), value.AuthorizedAt, 2*time.Second)
		assert.Equal(t, authorizedTransaction.Data, value.Data)
	}).Return(nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
  "info_status": "ok",
  "tx_status": "ok"
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))
}

func TestRequestHandlerAuthSanctionsCheckNeedInfo(t *testing.T) {
	var rhconfig = &config.Config{
		NetworkPassphrase: "Test SDF Network ; September 2015",
		Keys: config.Keys{
			// GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB
			SigningSeed: "SDWTLFPALQSP225BSMX7HPZ7ZEAYSUYNDLJ5QI3YGVBNRUIIELWH3XUV",
		},
		Callbacks: config.Callbacks{
			Sanctions: "http://sanctions",
			AskUser:   "http://ask_user",
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

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.HandlerAuth))
	defer testServer.Close()
	senderInfo := compliance.SenderInfo{FirstName: "John", LastName: "Doe"}
	senderInfoMap, err := senderInfo.Map()
	require.NoError(t, err)

	attachment := compliance.Attachment{
		Transaction: compliance.Transaction{
			Route:      "bob*acme.com",
			Note:       "Happy birthday",
			SenderInfo: senderInfoMap,
			Extra:      "extra",
		},
	}

	attachHash, err := attachment.Hash()
	require.NoError(t, err)
	attachHashB64 := base64.StdEncoding.EncodeToString(attachHash[:])

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

	txHash, err := tx.Hash()
	require.NoError(t, err)
	txHashHex := hex.EncodeToString(txHash[:])

	attachmentJSON, err := attachment.Marshal()
	require.NoError(t, err)

	senderInfoJSON, err := json.Marshal(attachment.Transaction.SenderInfo)
	require.NoError(t, err)

	// When all params are valid (NeedInfo = `true`)
	authData := compliance.AuthData{
		Sender:         "alice*diamnet.org",
		NeedInfo:       true,
		Tx:             txB64,
		AttachmentJSON: string(attachmentJSON),
	}

	authDataJSON, err := authData.Marshal()
	require.NoError(t, err)

	params := url.Values{
		"data": {string(authDataJSON)},
		"sig":  {"Q2cQVOn/A+aOxrLLeUPwHmBm3LMvlfXN8tDHo4Oi6SxWWueMTDfRkC4XvRX4emLij+Npo7/GfrZ82CnT5yB5Dg=="},
	}

	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	// Make sanctions checks successful (tested in the previous test case)
	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	// when ask_user server returns forbidden it returns info_status `denied`
	mockHTTPClient.On(
		"PostForm",
		"http://ask_user",
		url.Values{
			"sender":       {string(senderInfoJSON)},
			"note":         {attachment.Transaction.Note},
			"amount":       {"20.0000000"},
			"asset_code":   {"USD"},
			"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		},
	).Return(
		mocks.BuildHTTPResponse(403, "forbidden"),
		nil,
	).Once()

	statusCode, response := mocks.GetResponse(testServer, params)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 403, statusCode)
	expected := test.StringToJSONMap(`{
					"info_status": "denied",
					"tx_status": "ok"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when ask_user server returns bad request it returns info_status `error`
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://ask_user",
		url.Values{
			"sender":       {string(senderInfoJSON)},
			"note":         {attachment.Transaction.Note},
			"amount":       {"20.0000000"},
			"asset_code":   {"USD"},
			"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		},
	).Return(
		mocks.BuildHTTPResponse(400, "{\"error\": \"Invalid name\"}"),
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 400, statusCode)
	expected = test.StringToJSONMap(`{
					"info_status": "error",
					"tx_status": "ok",
					"error": "Invalid name"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when ask_user server returns pending it returns info_status `pending`
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://ask_user",
		url.Values{
			"sender":       {string(senderInfoJSON)},
			"note":         {attachment.Transaction.Note},
			"amount":       {"20.0000000"},
			"asset_code":   {"USD"},
			"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		},
	).Return(
		mocks.BuildHTTPResponse(202, "{\"pending\": 300}"),
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 202, statusCode)
	expected = test.StringToJSONMap(`{
					"info_status": "pending",
					"tx_status": "ok",
					"pending": 300
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when ask_user server returns pending but invalid response body it returns info_status `pending` (600 seconds)
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://ask_user",
		url.Values{
			"sender":       {string(senderInfoJSON)},
			"note":         {attachment.Transaction.Note},
			"amount":       {"20.0000000"},
			"asset_code":   {"USD"},
			"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		},
	).Return(
		mocks.BuildHTTPResponse(202, "pending"),
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 202, statusCode)
	expected = test.StringToJSONMap(`{
				"info_status": "pending",
				"tx_status": "ok",
				"pending": 600
			}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when ask_user server returns ok it returns info_status `ok` and DestInfo and persists transaction
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://ask_user",
		url.Values{
			"sender":       {string(senderInfoJSON)},
			"note":         {attachment.Transaction.Note},
			"amount":       {"20.0000000"},
			"asset_code":   {"USD"},
			"asset_issuer": {"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://fetch_info",
		url.Values{"address": {"bob*acme.com"}},
	).Return(
		mocks.BuildHTTPResponse(200, "user data"),
		nil,
	).Once()

	authorizedTransaction := &db.AuthorizedTransaction{
		TransactionID:  txHashHex,
		Memo:           attachHashB64,
		TransactionXdr: txB64,
		Data:           params["data"][0],
	}

	mockDatabase.On(
		"InsertAuthorizedTransaction",
		mock.AnythingOfType("*db.AuthorizedTransaction"),
	).Run(func(args mock.Arguments) {
		value := args.Get(0).(*db.AuthorizedTransaction)
		assert.Equal(t, authorizedTransaction.TransactionID, value.TransactionID)
		assert.Equal(t, authorizedTransaction.Memo, value.Memo)
		assert.Equal(t, authorizedTransaction.TransactionXdr, value.TransactionXdr)
		assert.WithinDuration(t, time.Now(), value.AuthorizedAt, 2*time.Second)
		assert.Equal(t, authorizedTransaction.Data, value.Data)
	}).Return(nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
				"info_status": "ok",
				"tx_status": "ok",
				"dest_info": "user data"
			}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When no callbacks.ask_user server
	rhconfig.Callbacks.AskUser = ""

	// when FI allowed it returns info_status = `ok` and DestInfo and persists transaction
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockDatabase.On(
		"GetAllowedFIByDomain",
		"diamnet.org", // sender = `alice*diamnet.org`
	).Return(
		&db.AllowedFI{}, // It just returns existing record
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://fetch_info",
		url.Values{"address": {"bob*acme.com"}},
	).Return(
		mocks.BuildHTTPResponse(200, "user data"),
		nil,
	).Once()

	authorizedTransaction = &db.AuthorizedTransaction{
		TransactionID:  txHashHex,
		Memo:           attachHashB64,
		TransactionXdr: txB64,
		Data:           params["data"][0],
	}

	mockDatabase.On(
		"InsertAuthorizedTransaction",
		mock.AnythingOfType("*db.AuthorizedTransaction"),
	).Run(func(args mock.Arguments) {
		value := args.Get(0).(*db.AuthorizedTransaction)
		assert.Equal(t, authorizedTransaction.TransactionID, value.TransactionID)
		assert.Equal(t, authorizedTransaction.Memo, value.Memo)
		assert.Equal(t, authorizedTransaction.TransactionXdr, value.TransactionXdr)
		assert.WithinDuration(t, time.Now(), value.AuthorizedAt, 2*time.Second)
		assert.Equal(t, authorizedTransaction.Data, value.Data)
	}).Return(nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
						"info_status": "ok",
						"tx_status": "ok",
						"dest_info": "user data"
					}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when FI not allowed but User is allowed it returns info_status = `ok` and DestInfo and persists transaction
	mockDatabase.On(
		"GetAllowedFIByDomain",
		"diamnet.org", // sender = `alice*diamnet.org`
	).Return(
		nil,
		nil,
	).Once()

	mockDatabase.On(
		"GetAllowedUserByDomainAndUserID",
		"diamnet.org", // sender = `alice*diamnet.org`
		"alice",
	).Return(
		&db.AllowedUser{},
		nil,
	).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://fetch_info",
		url.Values{"address": {"bob*acme.com"}},
	).Return(
		mocks.BuildHTTPResponse(200, "user data"),
		nil,
	).Once()

	authorizedTransaction = &db.AuthorizedTransaction{
		TransactionID:  txHashHex,
		Memo:           attachHashB64,
		TransactionXdr: txB64,
		Data:           params["data"][0],
	}

	mockDatabase.On(
		"InsertAuthorizedTransaction",
		mock.AnythingOfType("*db.AuthorizedTransaction"),
	).Run(func(args mock.Arguments) {
		value := args.Get(0).(*db.AuthorizedTransaction)
		assert.Equal(t, authorizedTransaction.TransactionID, value.TransactionID)
		assert.Equal(t, authorizedTransaction.Memo, value.Memo)
		assert.Equal(t, authorizedTransaction.TransactionXdr, value.TransactionXdr)
		assert.WithinDuration(t, time.Now(), value.AuthorizedAt, 2*time.Second)
		assert.Equal(t, authorizedTransaction.Data, value.Data)
	}).Return(nil).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	expected = test.StringToJSONMap(`{
					"info_status": "ok",
					"tx_status": "ok",
					"dest_info": "user data"
				}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// when neither FI nor User is allowed it returns info_status = `denied`
	mockDiamNettomlResolver.On(
		"GetDiamNetTomlByAddress",
		"alice*diamnet.org",
	).Return(&diamnettoml.Response{
		SigningKey: "GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
	}, nil).Once()

	mockSignerVerifier.On(
		"Verify",
		"GBYJZW5XFAI6XV73H5SAIUYK6XZI4CGGVBUBO3ANA2SV7KKDAXTV6AEB",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil).Once()

	mockHTTPClient.On(
		"PostForm",
		"http://sanctions",
		url.Values{
			"sender": {string(senderInfoJSON)},
		},
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	mockDatabase.On(
		"GetAllowedFIByDomain",
		"diamnet.org", // sender = `alice*diamnet.org`
	).Return(
		nil,
		nil,
	).Once()

	mockDatabase.On(
		"GetAllowedUserByDomainAndUserID",
		"diamnet.org", // sender = `alice*diamnet.org`
		"alice",
	).Return(
		nil,
		nil,
	).Once()

	statusCode, response = mocks.GetResponse(testServer, params)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 403, statusCode)
	expected = test.StringToJSONMap(`{
  "info_status": "denied",
  "tx_status": "ok"
}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

}
