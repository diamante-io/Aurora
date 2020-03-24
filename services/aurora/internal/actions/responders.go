package actions

import "github.com/hcnet/go/services/aurora/internal/render/sse"

// JSONer implementors can respond to a request whose response type was negotiated
// to be MimeHal or MimeJSON.
type JSONer interface {
	JSON() error
}

// RawDataResponder implementors can respond to a request whose response type was negotiated
// to be MimeRaw.
type RawDataResponder interface {
	Raw() error
}

// EventStreamer implementors can respond to a request whose response type was negotiated
// to be MimeEventStream.
type EventStreamer interface {
	SSE(*sse.Stream) error
}

// SingleObjectStreamer implementors can respond to a request whose response
// type was negotiated to be MimeEventStream. A SingleObjectStreamer loads an
// object whenever a ledger is closed.
type SingleObjectStreamer interface {
	LoadEvent() (sse.Event, error)
}
