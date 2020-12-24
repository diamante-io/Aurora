package diamnettoml

import "github.com/stretchr/testify/mock"

// MockClient is a mockable diamnettoml client.
type MockClient struct {
	mock.Mock
}

// GetDiamNetToml is a mocking a method
func (m *MockClient) GetDiamNetToml(domain string) (*Response, error) {
	a := m.Called(domain)
	return a.Get(0).(*Response), a.Error(1)
}

// GetDiamNetTomlByAddress is a mocking a method
func (m *MockClient) GetDiamNetTomlByAddress(address string) (*Response, error) {
	a := m.Called(address)
	return a.Get(0).(*Response), a.Error(1)
}
