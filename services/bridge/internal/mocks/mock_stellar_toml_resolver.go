package mocks

import (
	"github.com/diamnet/go/clients/diamnettoml"
	"github.com/stretchr/testify/mock"
)

// MockDiamNettomlResolver ...
type MockDiamNettomlResolver struct {
	mock.Mock
}

// GetDiamNetToml is a mocking a method
func (m *MockDiamNettomlResolver) GetDiamNetToml(domain string) (resp *diamnettoml.Response, err error) {
	a := m.Called(domain)
	return a.Get(0).(*diamnettoml.Response), a.Error(1)
}

// GetDiamNetTomlByAddress is a mocking a method
func (m *MockDiamNettomlResolver) GetDiamNetTomlByAddress(addy string) (*diamnettoml.Response, error) {
	a := m.Called(addy)
	return a.Get(0).(*diamnettoml.Response), a.Error(1)
}
