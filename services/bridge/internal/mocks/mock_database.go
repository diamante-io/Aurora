package mocks

import (
	"github.com/hcnet/go/services/bridge/internal/db"
	"github.com/stretchr/testify/mock"
)

// MockDatabase ...
type MockDatabase struct {
	mock.Mock
}

// GetLastCursorValue is a mocking a method
func (m *MockDatabase) GetLastCursorValue() (cursor *string, err error) {
	a := m.Called()
	return a.Get(0).(*string), a.Error(1)
}

// InsertReceivedPayment is a mocking a method
func (m *MockDatabase) InsertReceivedPayment(payment *db.ReceivedPayment) error {
	a := m.Called(payment)
	return a.Error(0)
}

// UpdateReceivedPayment is a mocking a method
func (m *MockDatabase) UpdateReceivedPayment(payment *db.ReceivedPayment) error {
	a := m.Called(payment)
	return a.Error(0)
}

// GetReceivedPaymentByID is a mocking a method
func (m *MockDatabase) GetReceivedPaymentByID(operationID int64) (*db.ReceivedPayment, error) {
	a := m.Called(operationID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.ReceivedPayment), a.Error(1)
}

// GetReceivedPaymentByOperationID is a mocking a method
func (m *MockDatabase) GetReceivedPaymentByOperationID(operationID string) (*db.ReceivedPayment, error) {
	a := m.Called(operationID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.ReceivedPayment), a.Error(1)
}

func (m *MockDatabase) GetReceivedPayments(page, limit uint64) ([]*db.ReceivedPayment, error) {
	a := m.Called(page, limit)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*db.ReceivedPayment), a.Error(1)
}

// InsertSentTransaction is a mocking a method
func (m *MockDatabase) InsertSentTransaction(transaction *db.SentTransaction) error {
	a := m.Called(transaction)
	return a.Error(0)
}

// UpdateSentTransaction is a mocking a method
func (m *MockDatabase) UpdateSentTransaction(transaction *db.SentTransaction) error {
	a := m.Called(transaction)
	return a.Error(0)
}

func (m *MockDatabase) GetSentTransactions(page, limit uint64) ([]*db.SentTransaction, error) {
	a := m.Called(page, limit)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).([]*db.SentTransaction), a.Error(1)
}

func (m *MockDatabase) GetSentTransactionByPaymentID(paymentID string) (*db.SentTransaction, error) {
	a := m.Called(paymentID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*db.SentTransaction), a.Error(1)
}
