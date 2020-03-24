package io

import (
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/mock"
)

var _ StateWriter = (*MockStateWriter)(nil)

type MockStateWriter struct {
	mock.Mock
}

func (m *MockStateWriter) Write(entryChange xdr.LedgerEntryChange) error {
	args := m.Called(entryChange)
	return args.Error(0)
}

func (m *MockStateWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}
