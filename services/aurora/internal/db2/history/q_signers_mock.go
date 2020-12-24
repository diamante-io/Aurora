package history

import (
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/stretchr/testify/mock"
)

type MockQSigners struct {
	mock.Mock
}

func (m *MockQSigners) GetLastLedgerExpIngestNonBlocking() (uint32, error) {
	a := m.Called()
	return a.Get(0).(uint32), a.Error(1)
}

func (m *MockQSigners) GetLastLedgerExpIngest() (uint32, error) {
	a := m.Called()
	return a.Get(0).(uint32), a.Error(1)
}

func (m *MockQSigners) UpdateLastLedgerExpIngest(ledgerSequence uint32) error {
	a := m.Called(ledgerSequence)
	return a.Error(0)
}

func (m *MockQSigners) AccountsForSigner(signer string, page db2.PageQuery) ([]AccountSigner, error) {
	a := m.Called(signer, page)
	return a.Get(0).([]AccountSigner), a.Error(1)
}

func (m *MockQSigners) CreateAccountSigner(account, signer string, weight int32) error {
	a := m.Called(account, signer, weight)
	return a.Error(0)
}

func (m *MockQSigners) RemoveAccountSigner(account, signer string) error {
	a := m.Called(account, signer)
	return a.Error(0)
}
