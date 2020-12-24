package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols/bridge"
)

// Authorize implements /reprocess endpoint
func (rh *RequestHandler) Reprocess(w http.ResponseWriter, r *http.Request) {
	request := &bridge.ReprocessRequest{}
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

	operation, err := rh.Aurora.OperationDetail(request.OperationID)
	if err != nil {
		helpers.Write(w, &bridge.ReprocessResponse{Status: "error", Message: err.Error()})
		return
	}

	bridgePayment, err := rh.PaymentListener.ConvertToBridgePayment(operation)
	if err != nil {
		helpers.Write(w, &bridge.ReprocessResponse{Status: "error", Message: err.Error()})
		return
	}

	err = rh.PaymentListener.ReprocessPayment(bridgePayment, request.Force)

	if err != nil {
		helpers.Write(w, &bridge.ReprocessResponse{Status: "error", Message: err.Error()})
		return
	}

	helpers.Write(w, &bridge.ReprocessResponse{Status: "ok"})
}
