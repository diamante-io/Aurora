package bridge

import (
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols"
	"github.com/diamnet/go/support/errors"
)

// AuthorizeRequest represents request made to /authorize endpoint of bridge server
type AuthorizeRequest struct {
	AccountID string `form:"account_id" valid:"required,diamnet_accountid"`
	AssetCode string `form:"asset_code" valid:"required,diamnet_asset_code"`
}

func (r AuthorizeRequest) Validate(params ...interface{}) error {
	allowedAssets, ok := params[0].([]protocols.Asset)
	if !ok {
		return errors.New("Invalid assets validation param provided")
	}

	issuingAccountID, ok := params[1].(string)
	if !ok {
		return errors.New("Invalid IssuingAccount validation param provided")
	}

	if issuingAccountID == "" {
		return errors.New("Issuing Account config parameter required")
	}

	// Is asset allowed?
	allowed := false
	for _, asset := range allowedAssets {
		if asset.Code == r.AssetCode && asset.Issuer == issuingAccountID {
			allowed = true
			break
		}
	}

	if !allowed {
		return helpers.NewInvalidParameterError("asset", "Asset code not allowed.")
	}

	return nil
}
