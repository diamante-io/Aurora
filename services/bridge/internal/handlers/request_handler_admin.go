package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"github.com/diamnet/go/protocols/compliance"
	"github.com/diamnet/go/services/bridge/internal/db"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols/bridge"
	callback "github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
	"github.com/diamnet/go/support/errors"
)

// AdminReceivedPayment implements /admin/received-payments/{id} endpoint
func (rh *RequestHandler) AdminReceivedPayment(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	payment, err := rh.Database.GetReceivedPaymentByID(id)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error getting ReceivedPayments")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	if payment == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	paymentResponse, err := rh.Aurora.OperationDetail(payment.OperationID)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error getting operation from Aurora")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	bridgePayment, err := rh.PaymentListener.ConvertToBridgePayment(paymentResponse)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error converting operation to bridge payment type")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	var authData *compliance.AuthData
	if bridgePayment.MemoType == "hash" && rh.Config.Compliance != "" {
		authData, err = rh.getComplianceData(bridgePayment.Memo)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error loading compliance data")
			helpers.Write(w, helpers.InternalServerError)
			return
		}
	}

	response := struct {
		Payment   *db.ReceivedPayment    `json:"payment"`
		Operation bridge.PaymentResponse `json:"operation"`
		AuthData  *compliance.AuthData   `json:"auth_data"`
	}{payment, bridgePayment, authData}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "payments": payment}).Error("Error encoding ReceivedPayment")
		helpers.Write(w, helpers.InternalServerError)
		return
	}
}

func (rh *RequestHandler) getComplianceData(memo string) (*compliance.AuthData, error) {
	complianceRequestURL := rh.Config.Compliance + "/receive"
	complianceRequestBody := url.Values{"memo": {string(memo)}}

	log.WithFields(log.Fields{"url": complianceRequestURL, "body": complianceRequestBody}).Info("Sending request to compliance server")
	resp, err := rh.Client.PostForm(complianceRequestURL, complianceRequestBody)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending request to compliance server")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading compliance server response")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"status": resp.StatusCode, "body": string(body)}).Error("Error response from compliance server")
		return nil, errors.New("Error response from compliance server")
	}

	var receiveResponse callback.ReceiveResponse
	err = json.Unmarshal([]byte(body), &receiveResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot unmarshal receiveResponse")
	}

	var authData compliance.AuthData
	err = json.Unmarshal([]byte(receiveResponse.Data), &authData)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot unmarshal authData")
	}

	return &authData, nil
}

// AdminReceivedPayments implements /admin/received-payments endpoint
func (rh *RequestHandler) AdminReceivedPayments(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit := 10

	payments, err := rh.Database.GetReceivedPayments(uint64(page), uint64(limit))
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error loading ReceivedPayments")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(payments)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "payments": payments}).Error("Error encoding ReceivedPayments")
		helpers.Write(w, helpers.InternalServerError)
		return
	}
}

// AdminReceivedPayments implements /admin/sent-transactions endpoint
func (rh *RequestHandler) AdminSentTransactions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit := 10

	transactions, err := rh.Database.GetSentTransactions(uint64(page), uint64(limit))
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error loading SentTransactions")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(transactions)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "transactions": transactions}).Error("Error encoding SentTransactions")
		helpers.Write(w, helpers.InternalServerError)
		return
	}
}
