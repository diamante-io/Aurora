package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/bridge"
	"github.com/hcnet/go/txnbuild"
)

// Authorize implements /authorize endpoint
func (rh *RequestHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	request := &bridge.AuthorizeRequest{}
	err := helpers.FromRequest(r, request)
	if err != nil {
		log.Error(err.Error())
		helpers.Write(w, helpers.InvalidParameterError)
		return
	}

	err = helpers.Validate(request, rh.Config.Assets, rh.Config.Accounts.IssuingAccountID)
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

	var sourceAccount *string
	if rh.Config.Accounts.IssuingAccountID != "" {
		sourceAccount = &rh.Config.Accounts.IssuingAccountID
	}

	allowTrustOp := bridge.AllowTrustOperationBody{
		Source:    sourceAccount,
		Authorize: true,
		Trustor:   request.AccountID,
		AssetCode: request.AssetCode,
	}

	operationBuilder := allowTrustOp.Build()

	submitResponse, err := rh.TransactionSubmitter.SubmitTransaction(
		nil,
		rh.Config.Accounts.AuthorizingSeed,
		[]txnbuild.Operation{operationBuilder},
		nil,
	)

	jsonEncoder := json.NewEncoder(w)

	if err != nil {
		herr, isAuroraError := err.(*hc.Error)
		if !isAuroraError {
			log.WithFields(log.Fields{"err": err}).Error("Error submitting transaction")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		w.WriteHeader(herr.Problem.Status)
		err = jsonEncoder.Encode(herr.Problem)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error encoding response")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		return
	}

	err = jsonEncoder.Encode(submitResponse)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error encoding response")
		helpers.Write(w, helpers.InternalServerError)
		return
	}
}
