package mocks

import (
	hProtocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/txnbuild"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/mock"
)

// MockTransactionSubmitter mocks TransactionSubmitter
type MockTransactionSubmitter struct {
	mock.Mock
}

// SubmitTransaction is a mocking a method
func (ts *MockTransactionSubmitter) SubmitTransaction(paymentID *string, seed string, operation []txnbuild.Operation, memo txnbuild.Memo) (hProtocol.TransactionSuccess, error) {
	a := ts.Called(paymentID, seed, operation, memo)
	return a.Get(0).(hProtocol.TransactionSuccess), a.Error(1)
}

// SignAndSubmitRawTransaction is a mocking a method
func (ts *MockTransactionSubmitter) SignAndSubmitRawTransaction(paymentID *string, seed string, tx *xdr.Transaction) (hProtocol.TransactionSuccess, error) {
	a := ts.Called(paymentID, seed, tx)
	return a.Get(0).(hProtocol.TransactionSuccess), a.Error(1)
}
