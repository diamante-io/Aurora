package mocks

import (
	"net/http"
	"net/url"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient ...
type MockHTTPClient struct {
	mock.Mock
}

// HTTPClientInterface helps mocking http.Client in tests
type HTTPClientInterface interface {
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
}

// PostForm is a mocking a method
func (m *MockHTTPClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	a := m.Called(url, data)
	return a.Get(0).(*http.Response), a.Error(1)
}

// Get is a mocking a method
func (m *MockHTTPClient) Get(url string) (resp *http.Response, err error) {
	a := m.Called(url)
	return a.Get(0).(*http.Response), a.Error(1)
}

// Do is a mocking a method
func (m *MockHTTPClient) Do(req *http.Request) (resp *http.Response, err error) {
	a := m.Called(req)
	return a.Get(0).(*http.Response), a.Error(1)
}
