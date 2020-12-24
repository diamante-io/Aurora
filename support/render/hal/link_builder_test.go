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
	check("/root", "//diamnet.org", "//diamnet.org/root")
	check("/root", "https://diamnet.org", "https://diamnet.org/root")
	check("//else.org/root", "", "//else.org/root")
	check("//else.org/root", "//diamnet.org", "//else.org/root")
	check("//else.org/root", "https://diamnet.org", "//else.org/root")
	check("https://else.org/root", "", "https://else.org/root")
	check("https://else.org/root", "//diamnet.org", "https://else.org/root")
	check("https://else.org/root", "https://diamnet.org", "https://else.org/root")

	// Regression: ensure that parameters are not escaped
	check("/accounts/{id}", "https://diamnet.org", "https://diamnet.org/accounts/{id}")
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
