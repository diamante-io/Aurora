package compliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hcnet/go/protocols/compliance"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
)

// SendRequest represents request sent to /send endpoint of compliance server
type SendRequest struct {
	// Payment ID - used to resubmit auth request in case of `pending` response.
	ID string `form:"id" valid:"required"`
	// Source account ID
	Source string `form:"source" valid:"required,hcnet_accountid"`
	// Sender address (like alice*hcnet.org)
	Sender string `form:"sender" valid:"required,hcnet_address"`
	// Destination address (like bob*hcnet.org)
	Destination string `form:"destination" valid:"required,hcnet_address"`
	// ForwardDestination
	ForwardDestination *protocols.ForwardDestination `form:"forward_destination" valid:"-"`
	// Amount destination should receive
	Amount string `form:"amount" valid:"required,hcnet_amount"`
	// Code of the asset destination should receive
	AssetCode string `form:"asset_code" valid:"optional,hcnet_asset_code"`
	// Issuer of the asset destination should receive
	AssetIssuer string `form:"asset_issuer" valid:"optional,hcnet_accountid"`
	// Only for path_payment
	SendMax string `form:"send_max" valid:"optional,hcnet_amount"`
	// Only for path_payment
	SendAssetCode string `form:"send_asset_code" valid:"optional,hcnet_asset_code"`
	// Only for path_payment
	SendAssetIssuer string `form:"send_asset_issuer" valid:"optional,hcnet_accountid"`
	// path[n][asset_code] path[n][asset_issuer]
	Path []protocols.Asset `form:"path" valid:"-"`
	// Extra memo
	ExtraMemo string `form:"extra_memo" valid:"-"`
}

// ToValuesSpecial converts special values from http.Request to struct
func (request *SendRequest) FromRequestSpecial(r *http.Request, destination interface{}) error {
	var forwardDestination protocols.ForwardDestination
	forwardDestination.Domain = r.PostFormValue("forward_destination[domain]")
	forwardDestination.Fields = make(url.Values)

	err := r.ParseForm()
	if err != nil {
		return err
	}

	for key := range r.PostForm {
		matches := protocols.FederationDestinationFieldName.FindStringSubmatch(key)
		if len(matches) < 2 {
			continue
		}

		fieldName := matches[1]
		forwardDestination.Fields.Add(fieldName, r.PostFormValue(key))
	}

	if forwardDestination.Domain != "" && len(forwardDestination.Fields) > 0 {
		request.ForwardDestination = &forwardDestination
	}

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

	request.Path = path

	return nil
}

// ToValuesSpecial adds special values (not easily convertable) to given url.Values
func (request SendRequest) ToValuesSpecial(values url.Values) {
	if request.ForwardDestination != nil {
		values.Add("forward_destination[domain]", request.ForwardDestination.Domain)
		for key := range request.ForwardDestination.Fields {
			values.Add(fmt.Sprintf("forward_destination[fields][%s]", key), request.ForwardDestination.Fields.Get(key))
		}
	}

	for i, asset := range request.Path {
		values.Set(fmt.Sprintf(protocols.PathCodeField, i), asset.Code)
		values.Set(fmt.Sprintf(protocols.PathIssuerField, i), asset.Issuer)
	}
}

// Validate is additional validation method to validate special fields.
func (request *SendRequest) Validate(params ...interface{}) error {
	if request.Destination == "" && request.ForwardDestination == nil {
		return helpers.NewMissingParameter("destination")
	}

	asset := protocols.Asset{request.AssetCode, request.AssetIssuer}
	err := asset.Validate()
	if err != nil {
		return helpers.NewInvalidParameterError("asset", err.Error())
	}

	sendAsset := protocols.Asset{request.SendAssetCode, request.SendAssetIssuer}
	err = sendAsset.Validate()
	if err != nil {
		return helpers.NewInvalidParameterError("asset", err.Error())
	}

	for i, asset := range request.Path {
		err := asset.Validate()
		if err != nil {
			return helpers.NewInvalidParameterError(fmt.Sprintf("path[%d]", i), err.Error())
		}
	}

	return nil
}

// SendResponse represents response returned by /send endpoint
type SendResponse struct {
	helpers.SuccessResponse
	compliance.AuthResponse `json:"auth_response"`
	// xdr.Transaction base64-encoded. Sequence number of this transaction will be equal 0.
	TransactionXdr string `json:"transaction_xdr,omitempty"`
}

// Marshal marshals SendResponse
func (response *SendResponse) Marshal() ([]byte, error) {
	return json.MarshalIndent(response, "", "  ")
}
