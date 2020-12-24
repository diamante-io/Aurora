package actions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"time"

	auroraContext "github.com/diamnet/go/services/aurora/internal/context"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/render"
	hProblem "github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/problem"
)

// Base is a helper struct you can use as part of a custom action via
// composition.
//
// TODO: example usage
type Base struct {
	W   http.ResponseWriter
	R   *http.Request
	Err error

	appCtx             context.Context
	sseUpdateFrequency time.Duration
	isSetup            bool
}

// Prepare established the common attributes that get used in nearly every
// action.  "Child" actions may override this method to extend action, but it
// is advised you also call this implementation to maintain behavior.
func (base *Base) Prepare(w http.ResponseWriter, r *http.Request, appCtx context.Context, sseUpdateFrequency time.Duration) {
	base.W = w
	base.R = r
	base.sseUpdateFrequency = sseUpdateFrequency
	base.appCtx = appCtx
}

// Execute trigger content negotiation and the actual execution of one of the
// action's handlers.
func (base *Base) Execute(action interface{}) {
	ctx := base.R.Context()
	contentType := render.Negotiate(base.R)

	switch contentType {
	case render.MimeHal, render.MimeJSON:
		action, ok := action.(JSONer)
		if !ok {
			goto NotAcceptable
		}

		err := action.JSON()
		if err != nil {
			problem.Render(ctx, base.W, err)
			return
		}

	case render.MimeEventStream:
		switch action.(type) {
		case EventStreamer, SingleObjectStreamer:
		default:
			goto NotAcceptable
		}

		stream := sse.NewStream(ctx, base.W)

		var oldHash [32]byte
		for {
			lastLedgerState := ledger.CurrentState()

			// Rate limit the request if it's a call to stream since it queries the DB every second. See
			// https://github.com/diamnet/go/issues/715 for more details.
			app := base.R.Context().Value(&auroraContext.AppContextKey)
			rateLimiter := app.(RateLimiterProvider).GetRateLimiter()
			if rateLimiter != nil {
				limited, _, err := rateLimiter.RateLimiter.RateLimit(rateLimiter.VaryBy.Key(base.R), 1)
				if err != nil {
					stream.Err(errors.Wrap(err, "RateLimiter error"))
					return
				}
				if limited {
					stream.Err(sse.ErrRateLimited)
					return
				}
			}

			switch ac := action.(type) {
			case EventStreamer:
				err := ac.SSE(stream)
				if err != nil {
					stream.Err(err)
					return
				}

			case SingleObjectStreamer:
				newEvent, err := ac.LoadEvent()
				if err != nil {
					stream.Err(err)
					return
				}
				resource, err := json.Marshal(newEvent.Data)
				if err != nil {
					stream.Err(errors.Wrap(err, "unable to marshal next action resource"))
					return
				}

				nextHash := sha256.Sum256(resource)
				if bytes.Equal(nextHash[:], oldHash[:]) {
					break
				}

				oldHash = nextHash
				stream.SetLimit(10)
				stream.Send(newEvent)
			}

			// Manually send the preamble in case there are no data events in SSE to trigger a stream.Send call.
			// This method is called every iteration of the loop, but is protected by a sync.Once variable so it's
			// only executed once.
			stream.Init()

			if stream.IsDone() {
				return
			}

			// Make sure this is buffered channel of size 1. Otherwise, the go routine below
			// will never return if `newLedgers` channel is not read. From Effective Go:
			// > If the channel is unbuffered, the sender blocks until the receiver has received the value.
			newLedgers := make(chan bool, 1)
			go func() {
				for {
					time.Sleep(base.sseUpdateFrequency)
					currentLedgerState := ledger.CurrentState()
					if currentLedgerState.HistoryLatest >= lastLedgerState.HistoryLatest+1 {
						newLedgers <- true
						return
					}
				}
			}()

			select {
			case <-newLedgers:
				continue
			case <-ctx.Done():
			case <-base.appCtx.Done():
			}

			stream.Done()
			return
		}
	case render.MimeRaw:
		action, ok := action.(RawDataResponder)
		if !ok {
			goto NotAcceptable
		}

		err := action.Raw()
		if err != nil {
			problem.Render(ctx, base.W, err)
			return
		}
	default:
		goto NotAcceptable
	}
	return

NotAcceptable:
	problem.Render(ctx, base.W, hProblem.NotAcceptable)
}

// Do executes the provided func iff there is no current error for the action.
// Provides a nicer way to invoke a set of steps that each may set `action.Err`
// during execution
func (base *Base) Do(fns ...func()) {
	for _, fn := range fns {
		if base.Err != nil {
			return
		}

		fn()
	}
}

// Setup runs the provided funcs if and only if no call to Setup() has been
// made previously on this action.
func (base *Base) Setup(fns ...func()) {
	if base.isSetup {
		return
	}
	base.Do(fns...)
	base.isSetup = true
}
