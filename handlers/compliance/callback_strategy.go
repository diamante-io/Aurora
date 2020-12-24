package compliance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	proto "github.com/diamnet/go/protocols/compliance"
	"github.com/diamnet/go/support/errors"
)

type pendingResponse struct {
	Pending int `json:"pending"`
}

// SanctionsCheck performs AML sanctions check of the sender.
func (s *CallbackStrategy) SanctionsCheck(data proto.AuthData, response *proto.AuthResponse) error {
	if s.SanctionsCheckURL == "" {
		response.TxStatus = proto.AuthStatusOk
		return nil
	}

	resp, body, err := sendRequest(s.SanctionsCheckURL, url.Values{"attachment": {data.AttachmentJSON}})
	if err != nil {
		return errors.Wrap(err, "Error connecting sanctions server")
	}

	err = parseResponse(resp, body, response)
	if err != nil {
		return errors.Wrap(err, "Error parsing sanctions server response")
	}

	return nil
}

// GetUserData check if user data is required and if so decides
// whether to allow access to customer data or not.
func (s *CallbackStrategy) GetUserData(data proto.AuthData, response *proto.AuthResponse) error {
	// If sender doesn't need info, return AuthStatusOk
	if !data.NeedInfo {
		response.InfoStatus = proto.AuthStatusOk
		return nil
	}

	// If there is no way to fetch data, return AuthStatusDenied
	if s.GetUserDataURL == "" {
		response.InfoStatus = proto.AuthStatusDenied
		return nil
	}

	resp, body, err := sendRequest(s.GetUserDataURL, url.Values{"attachment": {data.AttachmentJSON}})
	if err != nil {
		return errors.Wrap(err, "Error connecting fetch info server")
	}

	err = parseResponse(resp, body, response)
	if err != nil {
		return errors.Wrap(err, "Error parsing fetch info server response")
	}

	return nil
}

func sendRequest(url string, params url.Values) (resp *http.Response, body []byte, err error) {
	resp, err = http.PostForm(url, params)
	if err != nil {
		err = errors.Wrap(err, "Error connecting server")
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "Error reading server response")
		return
	}

	return
}

func parseResponse(resp *http.Response, body []byte, response *proto.AuthResponse) error {
	switch resp.StatusCode {
	case http.StatusOK: // AuthStatusOk
		response.TxStatus = proto.AuthStatusOk
		response.DestInfo = string(body)
	case http.StatusAccepted: // AuthStatusPending
		response.TxStatus = proto.AuthStatusPending

		var pending int
		pendingResponseObj := pendingResponse{}
		err := json.Unmarshal(body, &pendingResponseObj)
		if err != nil {
			return errors.New("Cannot parse pending response")
		}

		pending = pendingResponseObj.Pending

		// Check if SanctionsCheck pending time is smaller
		if pending > response.Pending {
			response.Pending = pending
		}
	case http.StatusForbidden: // AuthStatusDenied
		response.TxStatus = proto.AuthStatusDenied
	default:
		return fmt.Errorf("Invalid status code from server: %d", resp.StatusCode)
	}

	return nil
}
