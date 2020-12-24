package helpers

import (
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/diamnet/go/address"
	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/strkey"
)

func init() {
	govalidator.CustomTypeTagMap.Set("diamnet_accountid", govalidator.CustomTypeValidator(isDiamNetAccountID))
	govalidator.CustomTypeTagMap.Set("diamnet_seed", govalidator.CustomTypeValidator(isDiamNetSeed))
	govalidator.CustomTypeTagMap.Set("diamnet_asset_code", govalidator.CustomTypeValidator(isDiamNetAssetCode))
	govalidator.CustomTypeTagMap.Set("diamnet_address", govalidator.CustomTypeValidator(isDiamNetAddress))
	govalidator.CustomTypeTagMap.Set("diamnet_amount", govalidator.CustomTypeValidator(isDiamNetAmount))
	govalidator.CustomTypeTagMap.Set("diamnet_destination", govalidator.CustomTypeValidator(isDiamNetDestination))

}

func Validate(request Request, params ...interface{}) error {
	valid, err := govalidator.ValidateStruct(request)

	if !valid {
		fields := govalidator.ErrorsByField(err)
		for field, errorValue := range fields {
			switch {
			case errorValue == "non zero value required":
				return NewMissingParameter(field)
			case strings.HasSuffix(errorValue, "does not validate as diamnet_accountid"):
				return NewInvalidParameterError(field, "Account ID must start with `G` and contain 56 alphanum characters.")
			case strings.HasSuffix(errorValue, "does not validate as diamnet_seed"):
				return NewInvalidParameterError(field, "Account secret must start with `S` and contain 56 alphanum characters.")
			case strings.HasSuffix(errorValue, "does not validate as diamnet_asset_code"):
				return NewInvalidParameterError(field, "Asset code must be 1-12 alphanumeric characters.")
			case strings.HasSuffix(errorValue, "does not validate as diamnet_address"):
				return NewInvalidParameterError(field, "DiamNet address must be of form user*domain.com")
			case strings.HasSuffix(errorValue, "does not validate as diamnet_destination"):
				return NewInvalidParameterError(field, "DiamNet destination must be of form user*domain.com or start with `G` and contain 56 alphanum characters.")
			case strings.HasSuffix(errorValue, "does not validate as diamnet_amount"):
				return NewInvalidParameterError(field, "Amount must be positive and have up to 7 decimal places.")
			default:
				return NewInvalidParameterError(field, errorValue)
			}
		}
	}

	return request.Validate(params...)
}

// These are copied from support/config. Should we move them to /strkey maybe?
func isDiamNetAccountID(i interface{}, context interface{}) bool {
	enc, ok := i.(string)

	if !ok {
		return false
	}

	_, err := strkey.Decode(strkey.VersionByteAccountID, enc)
	return err == nil
}

func isDiamNetSeed(i interface{}, context interface{}) bool {
	enc, ok := i.(string)

	if !ok {
		return false
	}

	_, err := strkey.Decode(strkey.VersionByteSeed, enc)
	return err == nil
}

func isDiamNetAssetCode(i interface{}, context interface{}) bool {
	code, ok := i.(string)

	if !ok {
		return false
	}

	if !govalidator.IsByteLength(code, 1, 12) {
		return false
	}

	if !govalidator.IsAlphanumeric(code) {
		return false
	}

	return true
}

func isDiamNetAddress(i interface{}, context interface{}) bool {
	addr, ok := i.(string)

	if !ok {
		return false
	}

	_, _, err := address.Split(addr)
	return err == nil
}

func isDiamNetAmount(i interface{}, context interface{}) bool {
	am, ok := i.(string)

	if !ok {
		return false
	}

	_, err := amount.Parse(am)
	return err == nil
}

// isDiamNetDestination checks if `i` is either account public key or DiamNet address.
func isDiamNetDestination(i interface{}, context interface{}) bool {
	dest, ok := i.(string)

	if !ok {
		return false
	}

	_, err1 := strkey.Decode(strkey.VersionByteAccountID, dest)
	_, _, err2 := address.Split(dest)

	if err1 != nil && err2 != nil {
		return false
	}

	return true
}
