package aurora

import (
	"bytes"
	"encoding/json"
	"strings"

	"net/url"

	"github.com/diamnet/go/support/render/problem"
	"github.com/stretchr/testify/assert"
)

// Assertions provides an assertions helper.  Custom assertions for this package
// can be defined as methods on this struct.
type Assertions struct {
	*assert.Assertions
}

func (a *Assertions) PageOf(length int, body *bytes.Buffer) bool {

	var result map[string]interface{}
	err := json.Unmarshal(body.Bytes(), &result)

	if !a.NoError(err, "failed to parse body") {
		return false
	}

	embedded, ok := result["_embedded"]

	if !a.True(ok, "_embedded not found in response") {
		return false
	}

	records, ok := embedded.(map[string]interface{})["records"]

	if !a.True(ok, "no 'records' property on _embedded object") {
		return false
	}

	return a.Len(records, length)
}

// Problem asserts that `body` is a serialized problem equal to `expected`,
// using Type and Status to compare for equality.
func (a *Assertions) Problem(body *bytes.Buffer, expected problem.P) bool {
	var actual problem.P
	err := json.Unmarshal(body.Bytes(), &actual)
	if !a.NoError(err, "failed to parse body") {
		return false
	}

	actual.Type = strings.TrimPrefix(actual.Type, problem.ServiceHost)
	if expected.Type != "" && a.Equal(expected.Type, actual.Type, "problem type didn't match") {
		return false
	}

	if expected.Status != 0 && a.Equal(expected.Status, actual.Status, "problem status didn't match") {
		return false
	}

	return true
}

// ProblemType asserts that the provided `body` is a JSON serialized problem
// whose type is `typ`
func (a *Assertions) ProblemType(body *bytes.Buffer, typ string) bool {
	var actual problem.P
	err := json.Unmarshal(body.Bytes(), &actual)
	if !a.NoError(err, "failed to parse body") {
		return false
	}

	return a.Problem(body, problem.P{Type: typ})
}

// EqualUrlStrings asserts for equality between url strings, regardless of query params ordering
func (a *Assertions) EqualUrlStrings(expected string, actual string) bool {

	// this function generates a golang URL struct from
	// each string and re-encodes the query params,
	// which sorts and allows for a simple equality check

	expectedU, err := url.Parse(expected)
	if !a.NoError(err) {
		return false
	}
	expectedU.RawQuery = expectedU.Query().Encode()

	actualU, err := url.Parse(actual)
	if !a.NoError(err) {
		return false
	}
	actualU.RawQuery = actualU.Query().Encode()

	return a.Equal(expectedU, actualU)
}
