package hcnettoml

import (
	"fmt"
	"io"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/hcnet/go/address"
	"github.com/hcnet/go/support/errors"
)

// GetHcNetToml returns hcnet.toml file for a given domain
func (c *Client) GetHcNetToml(domain string) (resp *Response, err error) {
	var hresp *http.Response
	hresp, err = c.HTTP.Get(c.url(domain))
	if err != nil {
		err = errors.Wrap(err, "http request errored")
		return
	}
	defer hresp.Body.Close()

	if !(hresp.StatusCode >= 200 && hresp.StatusCode < 300) {
		err = errors.New("http request failed with non-200 status code")
		return
	}

	limitReader := io.LimitReader(hresp.Body, HcNetTomlMaxSize)
	_, err = toml.DecodeReader(limitReader, &resp)

	// There is one corner case not handled here: response is exactly
	// HcNetTomlMaxSize long and is incorrect toml. Check discussion:
	// https://github.com/hcnet/go/pull/24#discussion_r89909696
	if err != nil && limitReader.(*io.LimitedReader).N == 0 {
		err = errors.Errorf("hcnet.toml response exceeds %d bytes limit", HcNetTomlMaxSize)
		return
	}

	if err != nil {
		err = errors.Wrap(err, "toml decode failed")
		return
	}

	return
}

// GetHcNetTomlByAddress returns hcnet.toml file of a domain fetched from a
// given address
func (c *Client) GetHcNetTomlByAddress(addy string) (*Response, error) {
	_, domain, err := address.Split(addy)
	if err != nil {
		return nil, errors.Wrap(err, "parse address failed")
	}

	return c.GetHcNetToml(domain)
}

// url returns the appropriate url to load for resolving domain's hcnet.toml
// file
func (c *Client) url(domain string) string {
	var scheme string

	if c.UseHTTP {
		scheme = "http"
	} else {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, domain, WellKnownPath)
}
