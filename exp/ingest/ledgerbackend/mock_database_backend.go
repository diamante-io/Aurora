package ledgerbackend

import (
	"github.com/stretchr/testify/mock"
)

var _ LedgerBackend = (*MockDatabaseBackend)(nil)

type MockDatabaseBackend struct {
	mock.Mock
}

func (m *MockDatabaseBackend) GetLatestLedgerSequence() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockDatabaseBackend) GetLedger(sequence uint32) (bool, LedgerCloseMeta, error) {
	args := m.Called()
	return args.Bool(0), args.Get(1).(LedgerCloseMeta), args.Error(2)
}

func (m *MockDatabaseBackend) Close() error {
	args := m.Called()
	return args.Error(0)
}
