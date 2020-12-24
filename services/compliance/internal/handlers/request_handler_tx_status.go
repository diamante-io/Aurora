package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/diamnet/go/protocols/compliance"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
)

// HandlerTxStatus implements /tx_status endpoint
func (rh *RequestHandler) HandlerTxStatus(w http.ResponseWriter, r *http.Request) {
	txid := r.URL.Query().Get("id")
	if txid == "" {
		log.Info("unable to get query parameter")
		helpers.Write(w, helpers.NewMissingParameter("id"))
		return
	}
	response := compliance.TransactionStatusResponse{}

	if rh.Config.Callbacks.TxStatus == "" {
		response.Status = compliance.TransactionStatusUnknown
	} else {

		u, err := url.Parse(rh.Config.Callbacks.TxStatus)
		if err != nil {
			log.Error(err, "failed to parse tx status endpoint")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		q := u.Query()
		q.Set("id", txid)
		u.RawQuery = q.Encode()
		resp, err := rh.Client.Get(u.String())
		if err != nil {
			log.WithFields(log.Fields{
				"tx_status": u.String(),
				"err":       err,
			}).Error("Error sending request to tx_status server")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error reading tx_status server response")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		switch resp.StatusCode {
		case http.StatusOK:
			err := json.Unmarshal(body, &response)
			if err != nil {
				log.WithFields(log.Fields{
					"tx_status": rh.Config.Callbacks.TxStatus,
					"body":      string(body),
				}).Error("Unable to decode tx_status response")
				helpers.Write(w, helpers.InternalServerError)
				return
			}
			if response.Status == "" {
				response.Status = compliance.TransactionStatusUnknown
			}

		default:
			response.Status = compliance.TransactionStatusUnknown
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error("Error encoding tx status response")
		helpers.Write(w, helpers.InternalServerError)
		return
	}
}
