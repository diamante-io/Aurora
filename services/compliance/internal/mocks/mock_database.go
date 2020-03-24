package mocks

import (
	"github.com/hcnet/go/services/compliance/internal/db"
	"github.com/stretchr/testify/mock"
)

// MockDatabase ...
type MockDatabase struct {
	mock.Mock
}

// GetAuthorizedTransactionByMemo is a mocking a method
func (m *MockDatabase) GetAuthorizedTransactionByMemo(memo string) (*db.AuthorizedTransaction, error) {
	a := m.Called(memo)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.AuthorizedTransaction), a.Error(1)
}

// InsertAuthorizedTransaction is a mocking a method
func (m *MockDatabase) InsertAuthorizedTransaction(transaction *db.AuthorizedTransaction) error {
	a := m.Called(transaction)
	return a.Error(0)
}

// InsertAllowedFI is a mocking a method
func (m *MockDatabase) InsertAllowedFI(fi *db.AllowedFI) error {
	a := m.Called(fi)
	return a.Error(0)
}

// GetAllowedFIByDomain is a mocking a method
func (m *MockDatabase) GetAllowedFIByDomain(domain string) (*db.AllowedFI, error) {
	a := m.Called(domain)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.AllowedFI), a.Error(1)
}

// DeleteAllowedFIByDomain is a mocking a method
func (m *MockDatabase) DeleteAllowedFIByDomain(domain string) error {
	a := m.Called(domain)
	return a.Error(0)
}

// InsertAllowedUser is a mocking a method
func (m *MockDatabase) InsertAllowedUser(user *db.AllowedUser) error {
	a := m.Called(user)
	return a.Error(0)
}

// GetAllowedUserByDomainAndUserID is a mocking a method
func (m *MockDatabase) GetAllowedUserByDomainAndUserID(domain, userID string) (*db.AllowedUser, error) {
	a := m.Called(domain, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.AllowedUser), a.Error(1)
}

// DeleteAllowedUserByDomainAndUserID is a mocking a method
func (m *MockDatabase) DeleteAllowedUserByDomainAndUserID(domain, userID string) error {
	a := m.Called(domain, userID)
	return a.Error(0)
}

// InsertAuthData is a mocking a method
func (m *MockDatabase) InsertAuthData(authData *db.AuthData) error {
	a := m.Called(authData)
	return a.Error(0)
}

// GetAuthData is a mocking a method
func (m *MockDatabase) GetAuthData(requestID string) (*db.AuthData, error) {
	a := m.Called(requestID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.AuthData), a.Error(1)
}
