package compliance

import (
	"encoding/json"
)

// Marshal marshals Attachment
func (r *AuthResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
