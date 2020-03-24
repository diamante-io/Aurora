package auroraclient

import "github.com/hcnet/go/support/errors"

// BuildURL returns the url for getting metrics about a running aurora instance
func (mr metricsRequest) BuildURL() (endpoint string, err error) {
	endpoint = mr.endpoint
	if endpoint == "" {
		err = errors.New("invalid request: too few parameters")
	}

	return
}
