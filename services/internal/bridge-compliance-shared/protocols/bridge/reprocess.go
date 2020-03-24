package bridge

import (
	"encoding/json"
	"net/http"
)

// ReprocessRequest represents request made to /reprocess endpoint of bridge server
type ReprocessRequest struct {
	OperationID string `form:"operation_id" valid:"required"`
	// Force is required for reprocessing successful payments. Please use with caution!
	Force bool `form:"force" valid:"-"`
}

func (r ReprocessRequest) Validate(params ...interface{}) error {
	// No custom validations
	return nil
}

// ReprocessResponse represents a response returned by /reprocess endpoint
type ReprocessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (r ReprocessResponse) HTTPStatus() int {
	if r.Status == "ok" {
		return http.StatusOK
	} else {
		return http.StatusBadRequest
	}
}

func (r ReprocessResponse) Marshal() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
