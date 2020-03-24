package handlers

import (
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	callback "github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
)

// HandlerReceive implements /receive endpoint
func (rh *RequestHandler) HandlerReceive(w http.ResponseWriter, r *http.Request) {
	request := &callback.ReceiveRequest{}
	err := helpers.FromRequest(r, request)
	if err != nil {
		log.Error(err.Error())
		helpers.Write(w, helpers.InvalidParameterError)
		return
	}

	err = helpers.Validate(request)
	if err != nil {
		switch err := err.(type) {
		case *helpers.ErrorResponse:
			helpers.Write(w, err)
		default:
			log.Error(err)
			helpers.Write(w, helpers.InternalServerError)
		}
		return
	}

	authorizedTransaction, err := rh.Database.GetAuthorizedTransactionByMemo(request.Memo)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error getting authorizedTransaction")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	if authorizedTransaction == nil {
		log.WithFields(log.Fields{"memo": request.Memo}).Warn("authorizedTransaction not found")
		helpers.Write(w, callback.TransactionNotFoundError)
		return
	}

	response := callback.ReceiveResponse{Data: authorizedTransaction.Data}
	helpers.Write(w, &response)
}
