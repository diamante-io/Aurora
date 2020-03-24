package sse

import (
	"net/http"
	"sync"

	"github.com/r3labs/sse"
	"github.com/hcnet/go/support/log"
)

// AddressEvent is an event sent to address SSE stream.
type AddressEvent string

const (
	TransactionReceivedAddressEvent AddressEvent = "transaction_received"
	AccountCreatedAddressEvent      AddressEvent = "account_created"
	ExchangedEvent                  AddressEvent = "exchanged"
	ExchangedTimelockedEvent        AddressEvent = "exchanged_timelocked"
)

type Server struct {
	Storage Storage `inject:""`

	lastID       int64
	eventsServer *sse.Server
	initOnce     sync.Once
	log          *log.Entry
}

type ServerInterface interface {
	BroadcastEvent(address string, event AddressEvent, data []byte)
	StartPublishing() error
	CreateStream(address string)
	StreamExists(address string) bool
	HTTPHandler(w http.ResponseWriter, r *http.Request)
}

type Event struct {
	Address string       `db:"address"`
	Event   AddressEvent `db:"event"`
	Data    string       `db:"data"`
}

// Storage contains history of sent events. Because each transaction and
// HcNet account is always processed by a single Bifrost server, we need
// to broadcast events in case client streams events from the other Bifrost
// server.
//
// It's used to broadcast events to all instances of Bifrost server and
// to handle clients' reconnections.
type Storage interface {
	// AddEvent adds a new server-sent event to the storage.
	AddEvent(event Event) error
	// GetEventsSinceID returns all events since `id`. Used to load and publish
	// all broadcasted events.
	// It returns the last event ID, list of events or error.
	// If `id` is equal `-1`:
	//    * it should return the last event ID and empty list if at least one
	//      event has been broadcasted.
	//    * it should return 0 if no events have been broadcasted.
	GetEventsSinceID(id int64) (int64, []Event, error)
}
