package mocks

import (
	"net/url"

	fprotocol "github.com/hcnet/go/protocols/federation"
	"github.com/stretchr/testify/mock"
)

// MockFederationResolver ...
type MockFederationResolver struct {
	mock.Mock
}

// LookupByAddress is a mocking a method
func (m *MockFederationResolver) LookupByAddress(addy string) (*fprotocol.NameResponse, error) {
	a := m.Called(addy)
	return a.Get(0).(*fprotocol.NameResponse), a.Error(1)
}

// LookupByAccountID is a mocking a method
func (m *MockFederationResolver) LookupByAccountID(aid string) (*fprotocol.IDResponse, error) {
	a := m.Called(aid)
	return a.Get(0).(*fprotocol.IDResponse), a.Error(1)
}

// ForwardRequest is a mocking a method
func (m *MockFederationResolver) ForwardRequest(domain string, fields url.Values) (*fprotocol.NameResponse, error) {
	a := m.Called(domain, fields)
	return a.Get(0).(*fprotocol.NameResponse), a.Error(1)
}
