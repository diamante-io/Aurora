package compliance

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/asaskevich/govalidator"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// Marshal marshals Attachment
func (d *AuthData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// Attachment returns attachment from the the object.
func (d AuthData) Attachment() (attachment Attachment, err error) {
	err = json.Unmarshal([]byte(d.AttachmentJSON), &attachment)
	return
}

// Validate checks if fields are of required form:
//  * `Sender` field is valid address
//  * `Tx` is valid and it's memo_hash equals sha256 hash of attachment preimage
//  * `Attachment` is valid JSON
func (d AuthData) Validate() error {
	valid, err := govalidator.ValidateStruct(d)

	if !valid {
		return err
	}

	// Check if Tx is a valid transaction
	var tx xdr.Transaction
	err = xdr.SafeUnmarshalBase64(d.Tx, &tx)
	if err != nil {
		return errors.Wrap(err, "Tx is invalid")
	}

	if tx.Memo.Hash == nil {
		return errors.New("Memo.Hash is nil")
	}

	// Check if Memo.Hash is sha256 hash of attachment preimage
	attachmentPreimageHashBytes := d.AttachmentPreimageHash()
	memoBytes := [32]byte(*tx.Memo.Hash)
	if attachmentPreimageHashBytes != memoBytes {
		return errors.New("Attachment preimage hash does not equal Memo.Hash in Tx")
	}

	return nil
}

// AttachmentPreimageHash returns sha-256 hash of memo preimage.
func (d AuthData) AttachmentPreimageHash() [32]byte {
	return sha256.Sum256([]byte(d.AttachmentJSON))
}
