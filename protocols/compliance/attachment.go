package compliance

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"time"
)

// GenerateNonce generates a nonce and assigns it to `Nonce` field.
// It does not have to be crypto random. We just want two attachments
// always have a different hashes.
func (attachment *Attachment) GenerateNonce() {
	attachment.Nonce = strconv.FormatInt(time.Now().UnixNano(), 10)
}

// Hash returns sha-256 hash of the JSON marshalled attachment.
func (attachment *Attachment) Hash() ([32]byte, error) {
	marshalled, err := attachment.Marshal()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(marshalled), nil
}

// Marshal marshals Attachment
func (attachment *Attachment) Marshal() ([]byte, error) {
	return json.Marshal(attachment)
}
