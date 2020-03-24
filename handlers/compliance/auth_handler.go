package compliance

import (
	"encoding/json"
	"net/http"

	complianceProtocol "github.com/hcnet/go/protocols/compliance"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
)

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authRequest := &complianceProtocol.AuthRequest{}
	authRequest.Populate(r)

	// Validate request
	err := authRequest.Validate()
	if err != nil {
		h.writeJSON(w, ErrorResponse{
			Code:    "invalid_request",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	authData, err := authRequest.Data()
	if err != nil {
		h.writeJSON(w, ErrorResponse{
			Code:    "invalid_data",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	err = authRequest.VerifySignature(authData.Sender)
	if err != nil {
		h.writeJSON(w, ErrorResponse{
			Code:    "invalid_signature",
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// Create response
	response := &complianceProtocol.AuthResponse{}

	// Sanctions check
	err = h.Strategy.SanctionsCheck(authData, response)
	if err != nil {
		h.writeError(w, err)
		return
	}

	// User info
	err = h.Strategy.GetUserData(authData, response)
	if err != nil {
		h.writeError(w, err)
		return
	}

	// If transaction allowed, persist it for future reference
	if response.TxStatus == complianceProtocol.AuthStatusOk && response.InfoStatus == complianceProtocol.AuthStatusOk {
		err = h.PersistTransaction(authData)
		if err != nil {
			h.writeError(w, err)
			return
		}
	}

	h.writeJSON(w, response, http.StatusOK)
}

/////////////////////////////////////////////////////////////
// Everything below copied from handlers/federation. We should probably move it
// to some `common` package.

// ErrorResponse represents the JSON response sent to a client when the request
// triggered an error.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *AuthHandler) writeJSON(
	w http.ResponseWriter,
	obj interface{},
	status int,
) {
	json, err := json.Marshal(obj)

	if err != nil {
		h.writeError(w, errors.Wrap(err, "response marshal"))
		return
	}

	if status == 0 {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(json)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, err error) {
	log.Error(err)
	http.Error(w, "An internal error occurred", http.StatusInternalServerError)
}
