package bridge

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
	complianceServer "github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
	"github.com/hcnet/go/support/errors"
)

var (
	// input errors

	// PaymentCannotResolveDestination is an error response
	PaymentCannotResolveDestination = &helpers.ErrorResponse{Code: "cannot_resolve_destination", Message: "Cannot resolve federated HcNet address.", Status: http.StatusBadRequest}
	// PaymentCannotUseMemo is an error response
	PaymentCannotUseMemo = &helpers.ErrorResponse{Code: "cannot_use_memo", Message: "Memo given in request but federation returned memo fields.", Status: http.StatusBadRequest}
	// PaymentSourceNotExist is an error response
	PaymentSourceNotExist = &helpers.ErrorResponse{Code: "source_not_exist", Message: "Source account does not exist.", Status: http.StatusBadRequest}
	// PaymentAssetCodeNotAllowed is an error response
	PaymentAssetCodeNotAllowed = &helpers.ErrorResponse{Code: "asset_code_not_allowed", Message: "Given asset_code not allowed.", Status: http.StatusBadRequest}

	// compliance

	// PaymentPending is an error response
	PaymentPending = &helpers.ErrorResponse{Code: "pending", Message: "Transaction pending. Repeat your request after given time.", Status: http.StatusAccepted}
	// PaymentDenied is an error response
	PaymentDenied = &helpers.ErrorResponse{Code: "denied", Message: "Transaction denied by destination.", Status: http.StatusForbidden}
)

// PaymentRequest represents request made to /payment endpoint of the bridge server
type PaymentRequest struct {
	// Payment ID
	ID string `form:"id" valid:"optional"`
	// Source account secret
	Source string `form:"source" valid:"optional,hcnet_seed"`
	// Sender address (like alice*hcnet.org)
	Sender string `form:"sender" valid:"optional,hcnet_address"`
	// Destination address (like bob*hcnet.org)
	Destination string `form:"destination" valid:"optional,hcnet_destination"`
	// ForwardDestination
	ForwardDestination *protocols.ForwardDestination `form:"forward_destination" valid:"-"`
	// Memo type
	MemoType string `form:"memo_type" valid:"optional"`
	// Memo value
	Memo string `form:"memo" valid:"optional"`
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
	Path []protocols.Asset `form:"path" valid:"optional"`
	// Determined whether to use compliance protocol or to send a simple payment.
	UseCompliance bool `form:"use_compliance" valid:"-"`
	// Extra memo. If set, UseCompliance value will be ignored and it will use compliance.
	ExtraMemo string `form:"extra_memo" valid:"-"`
}

// ToValuesSpecial converts special values from http.Request to struct
func (request *PaymentRequest) FromRequestSpecial(r *http.Request, destination interface{}) error {
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
func (request PaymentRequest) ToValuesSpecial(values url.Values) {
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

// ToComplianceSendRequest transforms PaymentRequest to complianceServer.SendRequest
func (request *PaymentRequest) ToComplianceSendRequest() *complianceServer.SendRequest {
	sourceKeypair, _ := keypair.Parse(request.Source)
	return &complianceServer.SendRequest{
		ID: request.ID,
		// Compliance does not sign the transaction, it just needs a public key
		Source:             sourceKeypair.Address(),
		Sender:             request.Sender,
		Destination:        request.Destination,
		ForwardDestination: request.ForwardDestination,
		Amount:             request.Amount,
		AssetCode:          request.AssetCode,
		AssetIssuer:        request.AssetIssuer,
		SendMax:            request.SendMax,
		SendAssetCode:      request.SendAssetCode,
		SendAssetIssuer:    request.SendAssetIssuer,
		Path:               request.Path,
		ExtraMemo:          request.ExtraMemo,
	}
}

// Validate is additional validation method to validate special fields.
func (request *PaymentRequest) Validate(params ...interface{}) error {
	baseSeed, ok := params[0].(string)
	if !ok {
		return errors.New("Invalid `baseSeed` validation param provided")
	}

	// If baseSeed is empty then request.Source is required
	if baseSeed == "" && request.Source == "" {
		return helpers.NewMissingParameter("source")
	}

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

// NewPaymentPendingError creates a new PaymentPending error
func NewPaymentPendingError(seconds int) *helpers.ErrorResponse {
	return &helpers.ErrorResponse{
		Status:  PaymentPending.Status,
		Code:    PaymentPending.Code,
		Message: PaymentPending.Message,
		Data:    map[string]interface{}{"pending": seconds},
	}
}

// PaymentResponse represents a response from the bridge server when a payment is received.
type PaymentResponse struct {
	ID              string
	Type            string
	PagingToken     string
	From            string
	To              string
	AssetType       string
	AssetCode       string
	AssetIssuer     string
	Amount          string
	TransactionHash string
	MemoType        string
	Memo            string
}
