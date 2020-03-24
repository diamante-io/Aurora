package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockSignerVerifier ...
type MockSignerVerifier struct {
	mock.Mock
}

// Sign is a mocking a method
func (m *MockSignerVerifier) Sign(secretSeed string, message []byte) (string, error) {
	a := m.Called(secretSeed, message)
	return a.String(0), a.Error(1)
}

// Verify is a mocking a method
func (m *MockSignerVerifier) Verify(publicKey string, message, signature []byte) error {
	a := m.Called(publicKey, message, signature)
	return a.Error(0)
}
