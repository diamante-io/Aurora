package listener

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/protocols/compliance"
	hProtocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/protocols/aurora/base"
	"github.com/hcnet/go/protocols/aurora/effects"
	"github.com/hcnet/go/protocols/aurora/operations"
	"github.com/hcnet/go/services/bridge/internal/config"
	"github.com/hcnet/go/services/bridge/internal/db"
	"github.com/hcnet/go/services/bridge/internal/mocks"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
	callback "github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
	"github.com/hcnet/go/strkey"
	"github.com/hcnet/go/support/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var plconfig = &config.Config{
	Assets: []protocols.Asset{
		{Code: "USD", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
		{Code: "EUR", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
	},
	Accounts: config.Accounts{
		IssuingAccountID:   "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB",
		ReceivingAccountID: "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB",
	},
	Callbacks: config.Callbacks{
		Receive: "http://receive_callback",
	},
}

var mockAurora = new(hc.MockClient)
var mockDatabase = new(mocks.MockDatabase)
var mockHTTPClient = new(mocks.MockHTTPClient)

func ensurePaymentStatus(t *testing.T, operation operations.Operation, status string) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		payment := args.Get(0).(*db.ReceivedPayment)
		assert.Equal(t, operation.GetID(), payment.OperationID)
		assert.Equal(t, mocks.PredefinedTime, payment.ProcessedAt)
		assert.Equal(t, operation.PagingToken(), payment.PagingToken)
		assert.Equal(t, status, payment.Status)
		assert.Equal(t, operation.GetTransactionHash(), payment.TransactionID)
	}
}

func setDefaultPaymentOperation() operations.Payment {
	baseOp := operations.Base{
		ID:                    "1",
		PT:                    "2",
		TransactionSuccessful: true,
		Type:                  "payment",
		TransactionHash:       "ad71fc31bfae25b0bd14add4cc5306661edf84cdd73f1353d2906363899167e1",
	}

	paymentAsset := base.Asset{
		Type: "native",
	}

	paymentOp := operations.Payment{
		Asset:  paymentAsset,
		Base:   baseOp,
		From:   "ABC",
		To:     "XYZ",
		Amount: "100",
	}

	return paymentOp
}

func resetConfig() *config.Config {
	return &config.Config{
		Assets: []protocols.Asset{
			{Code: "USD", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
			{Code: "EUR", Issuer: "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"},
		},
		Accounts: config.Accounts{
			IssuingAccountID:   "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB",
			ReceivingAccountID: "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB",
		},
		Callbacks: config.Callbacks{
			Receive: "http://receive_callback",
		},
	}

}

func TestPaymentListener(t *testing.T) {

	paymentListener, err := NewPaymentListener(
		plconfig,
		mockDatabase,
		mockAurora,
		mocks.Now,
	)
	require.NoError(t, err)
	paymentListener.client = mockHTTPClient

	paymentOp := setDefaultPaymentOperation()

	mocks.PredefinedTime = time.Now()

	// When operation exists it should save the status
	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(&db.ReceivedPayment{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when operation is not a payment it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "create_account"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Not a payment operation")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when operation is not permitted it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GDNXBMIJLLLXZYKZBHXJ45WQ4AJQBRVT776YKGQTDBHTSPMNAFO3OZOS"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Operation type not permitted")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when asset issuer is not allowed it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GC4WWLMUGZJMRVJM7JUVVZBY3LJ5HL4RKIPADEGKEMLAAJEDRONUGYG7"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Asset not allowed")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when asset code is not allowed it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = "GBP"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Asset not allowed")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when payment is XLM (no XLM asset in config) it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.From = "GBL27BKG2JSDU6KQ5YJKCDWTVIU24VTG4PLB63SF4K2DBZS5XZMWRPVU"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = ""
	paymentOp.Asset.Issuer = ""
	paymentOp.Asset.Type = "native"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Asset not allowed")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)

	// when payment is XLM (XLM asset in config) it should save the status
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = ""
	paymentOp.Asset.Issuer = ""
	paymentOp.Asset.Type = "native"
	plconfig.Assets[1].Code = "XLM"
	plconfig.Assets[1].Issuer = ""

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Success")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{MemoType: "book", Memo: "testing"}, nil).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// when unable to load transaction memo it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "unable to get transaction details: Connection error")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, errors.New("Connection error")).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// when receive callback returns error it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Error response from receive callback")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{MemoType: "text", Memo: "testing"}, nil).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(503, "ok"),
		nil,
	).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// when receive callback returns success it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.Base.TransactionHash = "abc123"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Amount = "100"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Success")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{MemoType: "text", Memo: "testing"}, nil).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)

		assert.Equal(t, paymentOp.From, req.PostFormValue("from"))
		assert.Equal(t, paymentOp.Amount, req.PostFormValue("amount"))
		assert.Equal(t, paymentOp.Asset.Code, req.PostFormValue("asset_code"))
		assert.Equal(t, paymentOp.Asset.Issuer, req.PostFormValue("asset_issuer"))
		assert.Equal(t, paymentOp.Base.TransactionHash, req.PostFormValue("transaction_id"))
		assert.Equal(t, "text", req.PostFormValue("memo_type"))
		assert.Equal(t, "testing", req.PostFormValue("memo"))
	}).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// when receive callback returns success(account merge) it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	accountMergeOp := operations.AccountMerge{
		Base:    paymentOp.Base,
		Into:    "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB",
		Account: "GBL27BKG2JSDU6KQ5YJKCDWTVIU24VTG4PLB63SF4K2DBZS5XZMWRPVU",
	}
	accountMergeOp.Base.Type = "account_merge"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, accountMergeOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, accountMergeOp, "Success")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{MemoType: "text", Memo: "testing"}, nil).Once()

	effectsBase := effects.Base{Type: "account_credited"}
	acEffect := effects.AccountCredited{Base: effectsBase, Amount: "100"}
	embedded := struct {
		Records []effects.Effect
	}{
		[]effects.Effect{acEffect},
	}
	effectsResponse := effects.EffectsPage{
		Embedded: embedded,
	}

	mockAurora.On("Effects", mock.AnythingOfType("auroraclient.EffectRequest")).Return(effectsResponse, nil).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)

		assert.Equal(t, accountMergeOp.Account, req.PostFormValue("from"))
		assert.Equal(t, paymentOp.Amount, req.PostFormValue("amount"))
		assert.Equal(t, "XLM", req.PostFormValue("asset_code"))
		assert.Equal(t, "", req.PostFormValue("asset_issuer"))
		assert.Equal(t, "text", req.PostFormValue("memo_type"))
		assert.Equal(t, "testing", req.PostFormValue("memo"))
	}).Once()

	paymentListener.onPayment(accountMergeOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// when receive callback returns success(no memo) it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Amount = "100"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Success")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{}, nil).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)

	// receive callback returns success and compliance server is connected it should save the status
	plconfig = resetConfig()
	paymentOp = setDefaultPaymentOperation()
	paymentOp.Base.Type = "payment"
	paymentOp.To = "GATKP6ZQM5CSLECPMTAC5226PE367QALCPM6AFHTSULPPZMT62OOPMQB"
	paymentOp.Amount = "100"
	paymentOp.Asset.Code = "USD"
	paymentOp.Asset.Issuer = "GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR"

	mockDatabase.On("InsertReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Processing...")).Return(nil).Once()

	mockDatabase.On("UpdateReceivedPayment", mock.AnythingOfType("*db.ReceivedPayment")).
		Run(ensurePaymentStatus(t, paymentOp, "Success")).Return(nil).Once()

	mockDatabase.On("GetReceivedPaymentByOperationID", "1").Return(nil, nil).Once()

	mockAurora.On("TransactionDetail", mock.AnythingOfType("string")).Return(hProtocol.Transaction{MemoType: "hash", Memo: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"}, nil).Once()

	attachment := compliance.Attachment{
		Transaction: compliance.Transaction{
			Route: "jed*hcnet.org",
		},
	}

	attachmentString, _ := json.Marshal(attachment)

	auth := compliance.AuthData{
		AttachmentJSON: string(attachmentString),
	}

	authString, _ := json.Marshal(auth)

	response := callback.ReceiveResponse{
		Data: string(authString),
	}

	responseString, _ := json.Marshal(response)

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://compliance/receive"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, string(responseString)),
		nil,
	).Once()

	mockHTTPClient.On(
		"Do",
		mock.MatchedBy(func(req *http.Request) bool {
			return req.URL.String() == "http://receive_callback"
		}),
	).Return(
		mocks.BuildHTTPResponse(200, "ok"),
		nil,
	).Once()

	paymentListener.onPayment(paymentOp)
	assert.Nil(t, err)
	mockDatabase.AssertExpectations(t)
	mockAurora.AssertExpectations(t)
}

func TestPostForm_MACKey(t *testing.T) {
	validKey := "SABLR5HOI2IUOYB27TR4TO7HWDJIGSRJTT4UUTXXZOFVVPGQKJ5ME43J"
	rawkey, err := strkey.Decode(strkey.VersionByteSeed, validKey)
	require.NoError(t, err)

	handler := http.NewServeMux()
	handler.HandleFunc("/no_mac", func(w http.ResponseWriter, req *http.Request) {
		assert.Empty(t, req.Header.Get("X_PAYLOAD_MAC"), "unexpected MAC present")
	})
	handler.HandleFunc("/mac", func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		macer := hmac.New(sha256.New, rawkey)
		macer.Write(body)
		rawExpected := macer.Sum(nil)
		encExpected := base64.StdEncoding.EncodeToString(rawExpected)

		assert.Equal(t, encExpected, req.Header.Get("X_PAYLOAD_MAC"), "MAC is wrong")
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	cfg := &config.Config{}
	pl, err := NewPaymentListener(cfg, nil, nil, nil)
	require.NoError(t, err)

	// no mac if the key is not set
	_, err = pl.postForm(srv.URL+"/no_mac", url.Values{"foo": []string{"base"}})
	require.NoError(t, err)

	// generates a valid mac if a key is set.
	cfg.MACKey = validKey
	_, err = pl.postForm(srv.URL+"/mac", url.Values{"foo": []string{"base"}})
	require.NoError(t, err)

	// errors is the key is invalid
	cfg.MACKey = "broken"
	_, err = pl.postForm(srv.URL+"/mac", url.Values{"foo": []string{"base"}})

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid MAC key")
	}
}
