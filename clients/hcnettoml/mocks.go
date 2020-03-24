package hcnettoml

import "github.com/stretchr/testify/mock"

// MockClient is a mockable hcnettoml client.
type MockClient struct {
	mock.Mock
}

// GetHcNetToml is a mocking a method
func (m *MockClient) GetHcNetToml(domain string) (*Response, error) {
	a := m.Called(domain)
	return a.Get(0).(*Response), a.Error(1)
}

// GetHcNetTomlByAddress is a mocking a method
func (m *MockClient) GetHcNetTomlByAddress(address string) (*Response, error) {
	a := m.Called(address)
	return a.Get(0).(*Response), a.Error(1)
}
