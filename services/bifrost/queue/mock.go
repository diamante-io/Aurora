package queue

import (
	"github.com/stretchr/testify/mock"
)

// MockQueue is a mockable queue.
type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) QueueAdd(tx Transaction) error {
	a := m.Called(tx)
	return a.Error(0)
}

func (m *MockQueue) QueuePool() (*Transaction, error) {
	a := m.Called()
	return a.Get(0).(*Transaction), a.Error(1)
}
