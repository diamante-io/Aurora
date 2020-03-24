package auroraclient

import "github.com/hcnet/go/support/errors"

// BuildURL returns the url for getting fee stats about a running aurora instance
func (fr feeStatsRequest) BuildURL() (endpoint string, err error) {
	endpoint = fr.endpoint
	if endpoint == "" {
		err = errors.New("invalid request: too few parameters")
	}

	return
}
