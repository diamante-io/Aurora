package bridge

import (
	"encoding/base64"

	shared "github.com/diamnet/go/services/internal/bridge-compliance-shared"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/txnbuild"
)

// ManageDataOperationBody represents manage_data operation
type ManageDataOperationBody struct {
	Source *string
	Name   string
	Data   string
}

// Build returns a txnbuild.Operation
func (op ManageDataOperationBody) Build() txnbuild.Operation {

	// This is validated in Validate()
	data, _ := base64.StdEncoding.DecodeString(op.Data)

	txnOp := txnbuild.ManageData{
		Name:  op.Name,
		Value: data,
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op ManageDataOperationBody) Validate() error {
	if len(op.Name) > 64 {
		return helpers.NewInvalidParameterError("name", "Name must be less than or equal 64 characters.")
	}

	data, err := base64.StdEncoding.DecodeString(op.Data)
	if err != nil || len(data) > 64 {
		return helpers.NewInvalidParameterError("data", "Not a valid base64 string.")
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
