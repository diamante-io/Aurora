package compliance

import (
	"github.com/hcnet/go/protocols/compliance"
)

// Strategy defines strategy for handling auth requests.
// The SanctionsCheck and GetUserData functions will be called in the
// order above. Both methods can set `Pending` field so make sure
// the final value is the max of two.
type Strategy interface {
	// SanctionsCheck performs AML sanctions check of the sender.
	SanctionsCheck(data compliance.AuthData, response *compliance.AuthResponse) error
	// GetUserData check if user data is required and if so decides
	// whether to allow access to customer data or not.
	GetUserData(data compliance.AuthData, response *compliance.AuthResponse) error
}

// CallbackStrategy sends requests to given callbacks to decide
// whether to allow incoming transaction.
// If SanctionsCheckURL is empty it will allow every transaction.
// If GetUserDataURL is empty it will deny access to user data for each request.
type CallbackStrategy struct {
	// SanctionsCheckURL callback should respond with one of the following
	// status codes:
	//   * `200 OK` when sender/receiver is allowed and the payment should be processed,
	//   * `202 Accepted` when your callback needs some time for processing,
	//   * `403 Forbidden` when sender/receiver is denied.
	// Any other status code will be considered an error.
	//
	// When `202 Accepted` is returned the response body should contain JSON object
	// with a pending field which represents the estimated number of seconds needed
	// for processing. For example, the following response means to try the payment
	// again in an hour:
	//
	//   {"pending": 3600}
	SanctionsCheckURL string
	// GetUserDataURL callback should respond with one of the following
	// status codes:
	//   * `200 OK` when you allow to share recipient data.
	//   * `202 Accepted` when your callback needs some time for processing,
	//   * `403 Forbidden` when you deny to share recipient data.
	// Any other status code will be considered an error.
	//
	// When `200 OK` is returned the response body should contain JSON object with
	// your customer data. The customer information that is exchanged between FIs is
	// flexible but the typical fields are:
	//
	//   * Full Name
	//   * Date of birth
	//   * Physical address
	//
	// When `202 Accepted` is returned the response body should contain JSON object
	// with a pending field which represents the estimated number of seconds needed
	// for processing. For example, the following response means to try the payment
	// again in an hour:
	//
	//   {"pending": 3600}
	GetUserDataURL string
}

// AuthHandler ...
type AuthHandler struct {
	Strategy Strategy
	// PersistTransaction save authorized transaction to persistent storage so
	// memo preimage (attachment) can be fetched when a transaction is sent.
	PersistTransaction func(data compliance.AuthData) error
}

var _ Strategy = &CallbackStrategy{}
