package mocks

import (
	"github.com/hcnet/go/clients/hcnettoml"
	"github.com/stretchr/testify/mock"
)

// MockHcNettomlResolver ...
type MockHcNettomlResolver struct {
	mock.Mock
}

// GetHcNetToml is a mocking a method
func (m *MockHcNettomlResolver) GetHcNetToml(domain string) (resp *hcnettoml.Response, err error) {
	a := m.Called(domain)
	return a.Get(0).(*hcnettoml.Response), a.Error(1)
}

// GetHcNetTomlByAddress is a mocking a method
func (m *MockHcNettomlResolver) GetHcNetTomlByAddress(addy string) (*hcnettoml.Response, error) {
	a := m.Called(addy)
	return a.Get(0).(*hcnettoml.Response), a.Error(1)
}
