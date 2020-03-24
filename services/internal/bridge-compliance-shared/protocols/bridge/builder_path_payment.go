package bridge

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hcnet/go/amount"
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
	"github.com/hcnet/go/txnbuild"
)

// PathPaymentOperationBody represents path_payment operation
type PathPaymentOperationBody struct {
	Source *string

	SendMax   string          `json:"send_max"`
	SendAsset protocols.Asset `json:"send_asset"`

	Destination       string
	DestinationAmount string          `json:"destination_amount"`
	DestinationAsset  protocols.Asset `json:"destination_asset"`

	Path []protocols.Asset
}

// ToValuesSpecial converts special values from http.Request to struct
func (op *PathPaymentOperationBody) FromRequestSpecial(r *http.Request, destination interface{}) error {
	var path []protocols.Asset

	for i := 0; i < 5; i++ {
		codeFieldName := fmt.Sprintf(protocols.PathCodeField, i)
		issuerFieldName := fmt.Sprintf(protocols.PathIssuerField, i)

		// If the element does not exist in PostForm break the loop
		if _, exists := r.PostForm[codeFieldName]; !exists {
			break
		}

		code := r.PostFormValue(codeFieldName)
		issuer := r.PostFormValue(issuerFieldName)

		if code == "" && issuer == "" {
			path = append(path, protocols.Asset{})
		} else {
			path = append(path, protocols.Asset{code, issuer})
		}
	}

	op.Path = path
	return nil
}

// ToValuesSpecial adds special values (not easily convertable) to given url.Values
func (op PathPaymentOperationBody) ToValuesSpecial(values url.Values) {
	for i, asset := range op.Path {
		values.Set(fmt.Sprintf(protocols.PathCodeField, i), asset.Code)
		values.Set(fmt.Sprintf(protocols.PathIssuerField, i), asset.Issuer)
	}
}

// Build returns a txnbuild.Operation
func (op PathPaymentOperationBody) Build() txnbuild.Operation {

	var paths []txnbuild.Asset

	for _, asset := range op.Path {
		paths = append(paths, asset.ToBaseAsset())
	}

	txnOp := txnbuild.PathPayment{
		Destination: op.Destination,
		DestAmount:  op.DestinationAmount,
		DestAsset:   op.DestinationAsset.ToBaseAsset(),
		SendAsset:   op.SendAsset.ToBaseAsset(),
		SendMax:     op.SendMax,
		Path:        paths,
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op PathPaymentOperationBody) Validate() error {
	if !shared.IsValidAccountID(op.Destination) {
		return helpers.NewInvalidParameterError("destination", "Destination must be a public key (starting with `G`).")
	}

	_, err := amount.Parse(op.SendMax)
	if err != nil {
		return helpers.NewInvalidParameterError("send_max", "Not a valid amount.")
	}

	_, err = amount.Parse(op.DestinationAmount)
	if err != nil {
		return helpers.NewInvalidParameterError("destination_amount", "Not a valid amount.")
	}

	err = op.SendAsset.Validate()
	if err != nil {
		return helpers.NewInvalidParameterError("send_asset", err.Error())
	}

	err = op.DestinationAsset.Validate()
	if err != nil {
		return helpers.NewInvalidParameterError("destination_asset", err.Error())
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	for i, asset := range op.Path {
		err := asset.Validate()
		if err != nil {
			return helpers.NewInvalidParameterError(fmt.Sprintf("path[%d]", i), err.Error())
		}
	}

	return nil
}
