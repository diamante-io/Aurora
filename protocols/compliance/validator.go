package compliance

import (
	"github.com/asaskevich/govalidator"
	"github.com/diamnet/go/address"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.CustomTypeTagMap.Set("diamnet_address", govalidator.CustomTypeValidator(isDiamNetAddress))
}

func isDiamNetAddress(i interface{}, context interface{}) bool {
	addr, ok := i.(string)

	if !ok {
		return false
	}

	_, _, err := address.Split(addr)

	return err == nil
}
