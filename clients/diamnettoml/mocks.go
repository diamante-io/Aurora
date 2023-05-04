package diamnettoml

import "github.com/stretchr/testify/mock"

// MockClient is a mockable diamnettoml client.
type MockClient struct {
	mock.Mock
}

// GetDiamnetToml is a mocking a method
func (m *MockClient) GetDiamnetToml(domain string) (*Response, error) {
	a := m.Called(domain)
	return a.Get(0).(*Response), a.Error(1)
}

// GetDiamnetTomlByAddress is a mocking a method
func (m *MockClient) GetDiamnetTomlByAddress(address string) (*Response, error) {
	a := m.Called(address)
	return a.Get(0).(*Response), a.Error(1)
}
