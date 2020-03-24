package io

import (
	"github.com/stretchr/testify/mock"
)

var _ LedgerWriter = (*MockLedgerWriter)(nil)

type MockLedgerWriter struct {
	mock.Mock
}

func (m *MockLedgerWriter) Write(tx LedgerTransaction) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockLedgerWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}
