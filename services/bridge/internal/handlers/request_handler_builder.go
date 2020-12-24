package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	hc "github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/keypair"
	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols/bridge"
	"github.com/diamnet/go/txnbuild"
)

// Builder implements /builder endpoint
func (rh *RequestHandler) Builder(w http.ResponseWriter, r *http.Request) {
	var request bridge.BuilderRequest
	var sequenceNumber uint64

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error decoding request")
		helpers.Write(w, helpers.NewInvalidParameterError("", "Request body is not a valid JSON"))
		return
	}

	err = request.Process()
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

	if request.SequenceNumber == "" {
		accountRequest := hc.AccountRequest{AccountID: request.Source}
		var accountResponse protocol.Account
		accountResponse, err = rh.Aurora.AccountDetail(accountRequest)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Error when loading account")
			helpers.Write(w, helpers.InternalServerError)
			return
		}
		sequenceNumber, err = strconv.ParseUint(accountResponse.Sequence, 10, 64)
	} else {
		sequenceNumber, err = strconv.ParseUint(request.SequenceNumber, 10, 64)
		if err == nil {
			// decrement sequence number when it is provided because txnbuild will autoincrement
			// to do: remove in txnbuild v2.*
			sequenceNumber = sequenceNumber - 1
		}
	}

	if err != nil {
		errorResponse := helpers.NewInvalidParameterError("sequence_number", "Sequence number must be a number")
		helpers.Write(w, errorResponse)
		return
	}

	var txOps []txnbuild.Operation
	for _, operation := range request.Operations {
		txOps = append(txOps, operation.Body.Build())
	}

	tx := txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: request.Source, Sequence: int64(sequenceNumber)},
		Operations:    txOps,
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       rh.Config.NetworkPassphrase,
	}

	err = tx.Build()
	if err != nil {
		log.WithFields(log.Fields{"err": err, "request": request}).Error("TransactionBuilder returned error")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	for _, s := range request.Signers {
		kp, err := keypair.Parse(s)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "request": request}).Error("Error converting signers to keypairs")
			helpers.Write(w, helpers.InternalServerError)
			return
		}

		err = tx.Sign(kp.(*keypair.Full))
		if err != nil {
			log.WithFields(log.Fields{"err": err, "request": request}).Error("Error signing transaction")
			helpers.Write(w, helpers.InternalServerError)
			return
		}
	}

	txeBase64, err := tx.Base64()
	if err != nil {
		log.WithFields(log.Fields{"err": err, "request": request}).Error("Error encoding transaction envelope")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	helpers.Write(w, &bridge.BuilderResponse{TransactionEnvelope: txeBase64})
}
