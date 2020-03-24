package aurora

import (
	"encoding/json"
	"net/http"

	"github.com/hcnet/go/support/errors"
)

func decodeResponse(resp *http.Response, object interface{}) (err error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		auroraError := &Error{
			Response: resp,
		}
		decodeError := decoder.Decode(&auroraError.Problem)
		if decodeError != nil {
			return errors.Wrap(decodeError, "error decoding aurora.Problem")
		}
		return auroraError
	}

	err = decoder.Decode(&object)
	if err != nil {
		return
	}
	return
}