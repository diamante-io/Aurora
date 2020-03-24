package federation

import (
	"net/http"
	"net/url"

	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/clients/hcnettoml"
	proto "github.com/hcnet/go/protocols/federation"
)

// FederationResponseMaxSize is the maximum size of response from a federation server
const FederationResponseMaxSize = 100 * 1024

// DefaultTestNetClient is a default federation client for testnet
var DefaultTestNetClient = &Client{
	HTTP:        http.DefaultClient,
	Aurora:     hc.DefaultTestNetClient,
	HcNetTOML: hcnettoml.DefaultClient,
}

// DefaultPublicNetClient is a default federation client for pubnet
var DefaultPublicNetClient = &Client{
	HTTP:        http.DefaultClient,
	Aurora:     hc.DefaultPublicNetClient,
	HcNetTOML: hcnettoml.DefaultClient,
}

// Client represents a client that is capable of resolving a federation request
// using the internet.
type Client struct {
	HcNetTOML HcNetTOML
	HTTP        HTTP
	Aurora     Aurora
	AllowHTTP   bool
}

type ClientInterface interface {
	LookupByAddress(addy string) (*proto.NameResponse, error)
	LookupByAccountID(aid string) (*proto.IDResponse, error)
	ForwardRequest(domain string, fields url.Values) (*proto.NameResponse, error)
}

// Aurora represents a aurora client that can be consulted for data when
// needed as part of the federation protocol
type Aurora interface {
	HomeDomainForAccount(aid string) (string, error)
}

// HTTP represents the http client that a federation client uses to make http
// requests.
type HTTP interface {
	Get(url string) (*http.Response, error)
}

// HcNetTOML represents a client that can resolve a given domain name to
// hcnet.toml file.  The response is used to find the federation server that a
// query should be made against.
type HcNetTOML interface {
	GetHcNetToml(domain string) (*hcnettoml.Response, error)
}

// confirm interface conformity
var _ HcNetTOML = hcnettoml.DefaultClient
var _ HTTP = http.DefaultClient
var _ ClientInterface = &Client{}
