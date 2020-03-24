package sse

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockServer is a mockable SSE server.
type MockServer struct {
	mock.Mock
}

func (m *MockServer) BroadcastEvent(address string, event AddressEvent, data []byte) {
	m.Called(address, event, data)
}

func (m *MockServer) StartPublishing() error {
	a := m.Called()
	return a.Error(0)
}

func (m *MockServer) CreateStream(address string) {
	m.Called(address)
}

func (m *MockServer) StreamExists(address string) bool {
	a := m.Called(address)
	return a.Get(0).(bool)
}

func (m *MockServer) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}
