package scraper

import (
	"net/url"
	"testing"

	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/stretchr/testify/assert"
)

func TestShouldDiscardAsset(t *testing.T) {
	testAsset := hProtocol.AssetStat{
		Amount: "",
	}

	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount: "0.0",
	}
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount: "0",
	}
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 8,
	}
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 12,
	}
	testAsset.Code = "REMOVE"
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 100,
	}
	testAsset.Code = "SOMETHINGVALID"
	testAsset.Links.Toml.Href = ""
	assert.Equal(t, shouldDiscardAsset(testAsset, true), false)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 40,
	}
	testAsset.Code = "SOMETHINGVALID"
	testAsset.Links.Toml.Href = "http://www.diamnet.org/.well-known/diamnet.toml"
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 40,
	}
	testAsset.Code = "SOMETHINGVALID"
	testAsset.Links.Toml.Href = ""
	assert.Equal(t, shouldDiscardAsset(testAsset, true), true)

	testAsset = hProtocol.AssetStat{
		Amount:      "123901.0129310",
		NumAccounts: 40,
	}
	testAsset.Code = "SOMETHINGVALID"
	testAsset.Links.Toml.Href = "https://www.diamnet.org/.well-known/diamnet.toml"
	assert.Equal(t, shouldDiscardAsset(testAsset, true), false)
}

func TestDomainsMatch(t *testing.T) {
	tomlURL, _ := url.Parse("https://diamnet.org/diamnet.toml")
	orgURL, _ := url.Parse("https://diamnet.org/")
	assert.True(t, domainsMatch(tomlURL, orgURL))

	tomlURL, _ = url.Parse("https://assets.diamnet.org/diamnet.toml")
	orgURL, _ = url.Parse("https://diamnet.org/")
	assert.False(t, domainsMatch(tomlURL, orgURL))

	tomlURL, _ = url.Parse("https://diamnet.org/diamnet.toml")
	orgURL, _ = url.Parse("https://home.diamnet.org/")
	assert.True(t, domainsMatch(tomlURL, orgURL))

	tomlURL, _ = url.Parse("https://diamnet.org/diamnet.toml")
	orgURL, _ = url.Parse("https://home.diamnet.com/")
	assert.False(t, domainsMatch(tomlURL, orgURL))

	tomlURL, _ = url.Parse("https://diamnet.org/diamnet.toml")
	orgURL, _ = url.Parse("https://diamnet.com/")
	assert.False(t, domainsMatch(tomlURL, orgURL))
}

func TestIsDomainVerified(t *testing.T) {
	tomlURL := "https://diamnet.org/diamnet.toml"
	orgURL := "https://diamnet.org/"
	hasCurrency := true
	assert.True(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = "https://diamnet.org/diamnet.toml"
	orgURL = ""
	hasCurrency = true
	assert.True(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = ""
	orgURL = ""
	hasCurrency = true
	assert.False(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = "https://diamnet.org/diamnet.toml"
	orgURL = "https://diamnet.org/"
	hasCurrency = false
	assert.False(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = "http://diamnet.org/diamnet.toml"
	orgURL = "https://diamnet.org/"
	hasCurrency = true
	assert.False(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = "https://diamnet.org/diamnet.toml"
	orgURL = "http://diamnet.org/"
	hasCurrency = true
	assert.False(t, isDomainVerified(orgURL, tomlURL, hasCurrency))

	tomlURL = "https://diamnet.org/diamnet.toml"
	orgURL = "https://diamnet.com/"
	hasCurrency = true
	assert.False(t, isDomainVerified(orgURL, tomlURL, hasCurrency))
}
