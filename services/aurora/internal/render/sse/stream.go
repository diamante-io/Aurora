package sse

import (
	"context"
	"database/sql"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/diamnet/go/support/log"
	"github.com/diamnet/go/support/render/problem"
)

var (
	// default error
	errBadStream = errors.New("Unexpected stream error")

	// known errors
	errNoObject    = errors.New("Object not found")
	ErrRateLimited = errors.New("Rate limit exceeded")
)

var knownErrors = map[error]struct{}{
	sql.ErrNoRows:  struct{}{},
	ErrRateLimited: struct{}{},
}

type Stream struct {
	ctx      context.Context
	initSync sync.Once  // Variable to ensure that Init only writes the preamble once.
	mu       sync.Mutex // Mutex protects the following fields
	w        http.ResponseWriter
	done     bool
	sent     int
	limit    int
}

// NewStream creates a new stream against the provided response writer.
func NewStream(ctx context.Context, w http.ResponseWriter) *Stream {
	return &Stream{
		ctx: ctx,
		w:   w,
	}
}

// Init function is only executed once. It writes the preamble event which includes the HTTP response code and a
// hello message. This should be called before any method that writes to the client to ensure that the preamble
// has been sent first.
func (s *Stream) Init() {
	s.initSync.Do(func() {
		ok := WritePreamble(s.ctx, s.w)
		if !ok {
			s.done = true
		}
	})
}

func (s *Stream) Send(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Init()
	WriteEvent(s.ctx, s.w, e)
	s.sent++
}

func (s *Stream) SentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sent
}

func (s *Stream) SetLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limit = limit
}

func (s *Stream) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Init()
	WriteEvent(s.ctx, s.w, goodbyeEvent)
	s.done = true
}

// isDone checks to see if the stream is done. Not safe to call concurrently
// and meant for internal use.
func (s *Stream) isDone() bool {
	if s.limit == 0 {
		return s.done
	}

	return s.done || s.sent >= s.limit
}

// IsDone is safe to call concurrently and is exported.
func (s *Stream) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isDone()
}

func (s *Stream) Err(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If we haven't sent an event, we should simply return the normal HTTP
	// error because it means that we haven't sent the preamble.
	if s.sent == 0 {
		problem.Render(s.ctx, s.w, err)
		return
	}

	rootErr := errors.Cause(err)
	if rootErr == sql.ErrNoRows {
		//TODO: return errNoObject directly in SSE() methods.
		err = errNoObject
	}

	_, ok := knownErrors[rootErr]
	if !ok {
		log.Ctx(s.ctx).Error(err)
		err = errBadStream
	}

	s.Init()
	WriteEvent(s.ctx, s.w, Event{Error: err})
	s.done = true
}
