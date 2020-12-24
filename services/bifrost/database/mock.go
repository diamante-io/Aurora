package database

import (
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mockable database.
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) CreateAddressAssociation(chain Chain, diamnetAddress, address string, addressIndex uint32) error {
	a := m.Called(chain, diamnetAddress, address, addressIndex)
	return a.Error(0)
}

func (m *MockDatabase) GetAssociationByChainAddress(chain Chain, address string) (*AddressAssociation, error) {
	a := m.Called(chain, address)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*AddressAssociation), a.Error(1)
}

func (m *MockDatabase) GetAssociationByDiamNetPublicKey(diamnetPublicKey string) (*AddressAssociation, error) {
	a := m.Called(diamnetPublicKey)
	return a.Get(0).(*AddressAssociation), a.Error(1)
}

func (m *MockDatabase) AddProcessedTransaction(chain Chain, transactionID, receivingAddress string) (alreadyProcessing bool, err error) {
	a := m.Called(chain, transactionID, receivingAddress)
	return a.Get(0).(bool), a.Error(1)
}

func (m *MockDatabase) IncrementAddressIndex(chain Chain) (uint32, error) {
	a := m.Called(chain)
	return a.Get(0).(uint32), a.Error(1)
}

func (m *MockDatabase) ResetBlockCounters() error {
	a := m.Called()
	return a.Error(0)
}

func (m *MockDatabase) AddRecoveryTransaction(sourceAccount string, txEnvelope string) error {
	a := m.Called(sourceAccount, txEnvelope)
	return a.Error(0)
}
