package protocols

import (
	"net/url"
	"regexp"
)

var FederationDestinationFieldName = regexp.MustCompile(`forward_destination\[fields\]\[([a-z_-]+)\]`)

const (
	PathCodeField   = "path[%d][asset_code]"
	PathIssuerField = "path[%d][asset_issuer]"
)

// Asset represents asset
type Asset struct {
	Code   string `valid:"required" name:"asset_code" json:"code"`
	Issuer string `valid:"optional" name:"asset_issuer" json:"issuer"`
}

// ForwardDestination contains fields required to create forward federation request
type ForwardDestination struct {
	Domain string     `name:"domain"`
	Fields url.Values `name:"fields"`
}
