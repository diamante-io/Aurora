package diamnettoml

import "net/http"

// DiamNetTomlMaxSize is the maximum size of diamnet.toml file
const DiamNetTomlMaxSize = 100 * 1024

// WellKnownPath represents the url path at which the diamnet.toml file should
// exist to conform to the federation protocol.
const WellKnownPath = "/.well-known/diamnet.toml"

// DefaultClient is a default client using the default parameters
var DefaultClient = &Client{HTTP: http.DefaultClient}

// Client represents a client that is capable of resolving a DiamNet.toml file
// using the internet.
type Client struct {
	// HTTP is the http client used when resolving a DiamNet.toml file
	HTTP HTTP

	// UseHTTP forces the client to resolve against servers using plain HTTP.
	// Useful for debugging.
	UseHTTP bool
}

type ClientInterface interface {
	GetDiamNetToml(domain string) (*Response, error)
	GetDiamNetTomlByAddress(addy string) (*Response, error)
}

// HTTP represents the http client that a stellertoml resolver uses to make http
// requests.
type HTTP interface {
	Get(url string) (*http.Response, error)
}

// Response represents the results of successfully resolving a diamnet.toml file
type Response struct {
	AuthServer       string `toml:"AUTH_SERVER"`
	FederationServer string `toml:"FEDERATION_SERVER"`
	EncryptionKey    string `toml:"ENCRYPTION_KEY"`
	SigningKey       string `toml:"SIGNING_KEY"`
}

// GetDiamNetToml returns diamnet.toml file for a given domain
func GetDiamNetToml(domain string) (*Response, error) {
	return DefaultClient.GetDiamNetToml(domain)
}

// GetDiamNetTomlByAddress returns diamnet.toml file of a domain fetched from a
// given address
func GetDiamNetTomlByAddress(addy string) (*Response, error) {
	return DefaultClient.GetDiamNetTomlByAddress(addy)
}
