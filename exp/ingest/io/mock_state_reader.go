package io

import (
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/mock"
)

var _ StateReader = (*MockStateReader)(nil)

type MockStateReader struct {
	mock.Mock
}

func (m *MockStateReader) GetSequence() uint32 {
	args := m.Called()
	return args.Get(0).(uint32)
}

func (m *MockStateReader) Read() (xdr.LedgerEntryChange, error) {
	args := m.Called()
	return args.Get(0).(xdr.LedgerEntryChange), args.Error(1)
}

func (m *MockStateReader) Close() error {
	args := m.Called()
	return args.Error(0)
}
