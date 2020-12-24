package sse

import (
	"net/http"
	"time"

	"github.com/r3labs/sse"
	"github.com/diamnet/go/services/bifrost/common"
	"github.com/diamnet/go/support/log"
)

func (s *Server) init() {
	s.eventsServer = sse.New()
	s.lastID = -1
	s.log = common.CreateLogger("SSEServer")
}

func (s *Server) BroadcastEvent(address string, event AddressEvent, data []byte) {
	s.initOnce.Do(s.init)

	eventRecord := Event{
		Address: address,
		Event:   event,
		Data:    string(data),
	}
	err := s.Storage.AddEvent(eventRecord)
	if err != nil {
		s.log.WithFields(log.F{"err": err, "event": eventRecord}).Error("Error broadcasting event")
	}
}

// StartPublishing starts publishing events from the shared storage.
func (s *Server) StartPublishing() error {
	s.initOnce.Do(s.init)

	var err error
	s.lastID, _, err = s.Storage.GetEventsSinceID(s.lastID)
	if err != nil {
		return err
	}

	go func() {
		// Start publishing
		for {
			lastID, events, err := s.Storage.GetEventsSinceID(s.lastID)
			if err != nil {
				s.log.WithField("err", err).Error("Error GetEventsSinceID")
				time.Sleep(time.Second)
				continue
			}

			if len(events) == 0 {
				time.Sleep(time.Second)
				continue
			}

			for _, event := range events {
				s.publishEvent(event.Address, event.Event, []byte(event.Data))
			}

			s.lastID = lastID
		}
	}()

	return nil
}

func (s *Server) publishEvent(address string, event AddressEvent, data []byte) {
	s.initOnce.Do(s.init)

	// Create SSE stream if not exists
	if !s.eventsServer.StreamExists(address) {
		s.eventsServer.CreateStream(address)
	}

	// github.com/r3labs/sse does not send new lines - TODO create PR
	if data == nil {
		data = []byte("{}\n")
	} else {
		data = append(data, byte('\n'))
	}

	s.eventsServer.Publish(address, &sse.Event{
		ID:    []byte(event),
		Event: []byte(event),
		Data:  data,
	})
}

func (s *Server) CreateStream(address string) {
	s.initOnce.Do(s.init)
	s.eventsServer.CreateStream(address)
}

func (s *Server) StreamExists(address string) bool {
	s.initOnce.Do(s.init)
	return s.eventsServer.StreamExists(address)
}

func (s *Server) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	s.initOnce.Do(s.init)
	s.eventsServer.HTTPHandler(w, r)
}
