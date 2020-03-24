package hal

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkBuilder(t *testing.T) {
	// Link expansion test
	check := func(href string, base string, expectedResult string) {
		lb := LinkBuilder{mustParseURL(base)}
		result := lb.expandLink(href)
		assert.Equal(t, expectedResult, result)
	}

	check("/root", "", "/root")
	check("/root", "//hcnet.org", "//hcnet.org/root")
	check("/root", "https://hcnet.org", "https://hcnet.org/root")
	check("//else.org/root", "", "//else.org/root")
	check("//else.org/root", "//hcnet.org", "//else.org/root")
	check("//else.org/root", "https://hcnet.org", "//else.org/root")
	check("https://else.org/root", "", "https://else.org/root")
	check("https://else.org/root", "//hcnet.org", "https://else.org/root")
	check("https://else.org/root", "https://hcnet.org", "https://else.org/root")

	// Regression: ensure that parameters are not escaped
	check("/accounts/{id}", "https://hcnet.org", "https://hcnet.org/accounts/{id}")
}

func mustParseURL(base string) *url.URL {
	if base == "" {
		return nil
	}

	u, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	return u
}
