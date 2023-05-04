// Package config provides a common infrastructure for reading configuration
// data stored in local TOML files.
package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/asaskevich/govalidator"
	"github.com/diamnet/go/amount"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/support/errors"
)

// TLS represents a common configuration snippet for configuring TLS in a server process
type TLS struct {
	CertificateFile string `toml:"certificate-file" valid:"required"`
	PrivateKeyFile  string `toml:"private-key-file" valid:"required"`
}

// InvalidConfigError is the error that is returned when an invalid
// configuration is encountered by the `Read` func.
type InvalidConfigError struct {
	InvalidFields map[string]string
}

// Read takes the TOML configuration file at `path`, parses it into `dest` and
// then uses github.com/asaskevich/govalidator to validate the struct.
func Read(path string, dest interface{}) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return decode(string(bs), dest)
}

func decode(content string, dest interface{}) error {
	metadata, err := toml.Decode(content, dest)
	if err != nil {
		return errors.Wrap(err, "decode-file failed")
	}

	// Undecoded keys correspond to keys in the TOML document
	// that do not have a concrete type in config struct.
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		return errors.New("Unknown fields: " + fmt.Sprintf("%+v", undecoded))
	}

	valid, err := govalidator.ValidateStruct(dest)

	if valid {
		return nil
	}

	fields := govalidator.ErrorsByField(err)

	return &InvalidConfigError{
		InvalidFields: fields,
	}
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.CustomTypeTagMap.Set("diamnet_accountid", govalidator.CustomTypeValidator(isDiamnetAccountID))
	govalidator.CustomTypeTagMap.Set("diamnet_seed", govalidator.CustomTypeValidator(isDiamnetSeed))
	govalidator.CustomTypeTagMap.Set("diamnet_amount", govalidator.CustomTypeValidator(isDiamnetAmount))

}

func isDiamnetAmount(i interface{}, context interface{}) bool {
	enc, ok := i.(string)

	if !ok {
		return false
	}

	_, err := amount.Parse(enc)

	return err == nil
}

func isDiamnetAccountID(i interface{}, context interface{}) bool {
	enc, ok := i.(string)

	if !ok {
		return false
	}

	_, err := strkey.Decode(strkey.VersionByteAccountID, enc)

	return err == nil
}

func isDiamnetSeed(i interface{}, context interface{}) bool {
	enc, ok := i.(string)

	if !ok {
		return false
	}

	_, err := strkey.Decode(strkey.VersionByteSeed, enc)

	return err == nil
}
