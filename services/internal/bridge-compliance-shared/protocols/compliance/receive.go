package compliance

import (
	"encoding/json"

	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
)

// ReceiveRequest represents request sent to /receive endpoint of compliance server
type ReceiveRequest struct {
	Memo string `form:"memo" valid:"required"`
}

// Validate is additional validation method to validate special fields.
func (request *ReceiveRequest) Validate(params ...interface{}) error {
	return nil
}

// ReceiveResponse represents response returned by /receive endpoint
type ReceiveResponse struct {
	helpers.SuccessResponse
	// The AuthData hash of this memo.
	Data string `json:"data"`
}

// Marshal marshals ReceiveResponse
func (response *ReceiveResponse) Marshal() ([]byte, error) {
	return json.MarshalIndent(response, "", "  ")
}
