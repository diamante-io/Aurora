package config

import (
	"testing"

	"github.com/asaskevich/govalidator"
	"github.com/stretchr/testify/assert"
)

// NOTE: this test is presently failing because govalidator doesn't support
// optional fields that also use a custom validator.  We'll remove the build tag
// above that disabled it from running during tests when we fix upstream.
func TestOptionalDiamnetFields(t *testing.T) {
	var val struct {
		F1 string `valid:"diamnet_accountid,optional"`
		F2 string `valid:"optional,diamnet_accountid"`
		F3 string `valid:"diamnet_seed,optional"`
		F4 string `valid:"optional,diamnet_accountid"`
	}

	// run the validation
	ok, err := govalidator.ValidateStruct(val)
	assert.NoError(t, err)
	assert.True(t, ok)
}
