package aurora

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/hchi"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/render"
	hProblem "github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/toid"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/support/render/problem"
)

// streamFunc represents the signature of the function that handles requests
// with stream mode turned on using server-sent events.
type streamFunc func(context.Context, *sse.Stream, *indexActionQueryParams) error

// streamableEndpointHandler handles endpoints that have the stream mode
// available. It inspects the Accept header to determine which function to be
// executed. If it's "application/hal+json" or "application/json", then jfn
// will be executed with params. If it's "text/event-stream", then either sfn
// or jfn will be executed with the streamHandler with params.
func (we *web) streamableEndpointHandler(jfn interface{}, streamSingleObjectEnabled bool, sfn streamFunc, params interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch render.Negotiate(r) {
		case render.MimeHal, render.MimeJSON:
			if jfn == nil {
				problem.Render(ctx, w, hProblem.NotAcceptable)
				return
			}
			h, err := hal.Handler(jfn, params)
			if err != nil {
				panic(err)
			}

			h.ServeHTTP(w, r)
			return

		case render.MimeEventStream:
			if sfn == nil && !streamSingleObjectEnabled {
				problem.Render(ctx, w, hProblem.NotAcceptable)
				return
			}

			we.streamHandler(jfn, sfn, params).ServeHTTP(w, r)
			return
		}

		problem.Render(ctx, w, hProblem.NotAcceptable)
	})
}

// streamHandler handles requests with stream mode turned on using server-sent
// events. It will execute one of the provided streaming functions with the
// provided params.
// Note that we don't return an error if both jfn and sfn are not nil. sfn will
// simply take precedence.
func (we *web) streamHandler(jfn interface{}, sfn streamFunc, params interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stream := sse.NewStream(ctx, w)
		var oldHash [32]byte
		for {
			lastLedgerState := ledger.CurrentState()

			// Rate limit the request if it's a call to stream since it queries the DB every second. See
			// https://github.com/diamnet/go/issues/715 for more details.
			rateLimiter := we.rateLimiter
			if rateLimiter != nil {
				limited, _, err := rateLimiter.RateLimiter.RateLimit(rateLimiter.VaryBy.Key(r), 1)
				if err != nil {
					stream.Err(errors.Wrap(err, "RateLimiter error"))
					return
				}
				if limited {
					stream.Err(sse.ErrRateLimited)
					return
				}
			}

			if sfn != nil {
				err := sfn(ctx, stream, params.(*indexActionQueryParams))
				if err != nil {
					stream.Err(err)
					return
				}
			} else if jfn != nil {
				data, ok, err := hal.ExecuteFunc(ctx, jfn, params)
				if err != nil {
					if !ok {
						panic(err)
					}
					stream.Err(err)
					return
				}
				resource, err := json.Marshal(data)
				if err != nil {
					stream.Err(errors.Wrap(err, "unable to marshal next action resource"))
					return
				}

				nextHash := sha256.Sum256(resource)
				if !bytes.Equal(nextHash[:], oldHash[:]) {
					oldHash = nextHash
					stream.SetLimit(10)
					stream.Send(sse.Event{Data: data})
				}
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
					time.Sleep(we.sseUpdateFrequency)
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
			case <-we.appCtx.Done():
			}

			stream.Done()
			return
		}
	})
}

// streamShowActionHandler gets the showAction query params from the request
// and pass it on to streamableEndpointHandler.
func (we *web) streamShowActionHandler(jfn interface{}, requireAccountID bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		param, err := getShowActionQueryParams(r, requireAccountID)
		if err != nil {
			problem.Render(ctx, w, err)
			return
		}

		we.streamableEndpointHandler(jfn, true, nil, param).ServeHTTP(w, r)
	})
}

// streamIndexActionHandler gets the required params for indexable endpoints from
// the URL, validates the cursor is within history, and finally passes the
// indexAction query params to the more general purpose streamableEndpointHandler.
func (we *web) streamIndexActionHandler(jfn interface{}, sfn streamFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params, err := getIndexActionQueryParams(r, we.ingestFailedTx)
		if err != nil {
			problem.Render(ctx, w, err)
			return
		}

		err = validateCursorWithinHistory(params.PagingParams)
		if err != nil {
			problem.Render(ctx, w, err)
			return
		}

		we.streamableEndpointHandler(jfn, false, sfn, params).ServeHTTP(w, r)
	})
}

// showActionHandler handles all non-streamable endpoints.
func showActionHandler(jfn interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contentType := render.Negotiate(r)
		if jfn == nil || (contentType != render.MimeHal && contentType != render.MimeJSON) {
			problem.Render(ctx, w, hProblem.NotAcceptable)
			return
		}

		params, err := getShowActionQueryParams(r, false)
		if err != nil {
			problem.Render(ctx, w, err)
			return
		}

		h, err := hal.Handler(jfn, params)
		if err != nil {
			panic(err)
		}

		h.ServeHTTP(w, r)
	})
}

// accountIndexActionHandler handles /accounts index endpoints.
func accountIndexActionHandler(jfn interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contentType := render.Negotiate(r)
		if jfn == nil || (contentType != render.MimeHal && contentType != render.MimeJSON) {
			problem.Render(ctx, w, hProblem.NotAcceptable)
			return
		}

		params, err := getAccountsIndexActionQueryParams(r)
		if err != nil {
			problem.Render(ctx, w, err)
			return
		}

		h, err := hal.Handler(jfn, params)
		if err != nil {
			panic(err)
		}

		h.ServeHTTP(w, r)
	})
}

// getAccountID retrieves the account id by the provided key. The key is
// usually "account_id", "source_account", and "destination_account". The
// function would return an error if the account id is empty and the required
// flag is true.
func getAccountID(r *http.Request, key string, required bool) (string, error) {
	val, err := hchi.GetStringFromURL(r, key)
	if err != nil {
		return "", err
	}

	if val == "" && !required {
		return val, nil
	}

	_, err = strkey.Decode(strkey.VersionByteAccountID, val)
	if err != nil {
		// TODO: add errInvalidValue
		return "", problem.MakeInvalidFieldProblem(key, errors.New("invalid address"))
	}

	return val, nil
}

// getSignerKey retrieves the signer key by the provided key. The key is
// usually "signer". The function would return an error if the account id is
// empty and the required flag is true.
func getSignerKey(r *http.Request, key string, required bool) (string, error) {
	val, err := hchi.GetStringFromURL(r, key)
	if err != nil {
		return "", err
	}

	if val == "" && !required {
		return val, nil
	}

	version, _, err := strkey.DecodeAny(val)
	if err != nil || version == strkey.VersionByteSeed {
		return "", problem.MakeInvalidFieldProblem(key, errors.New("invalid signer"))
	}

	return val, nil
}

// getShowActionQueryParams gets the available query params for all non-indexable endpoints.
func getShowActionQueryParams(r *http.Request, requireAccountID bool) (*showActionQueryParams, error) {
	txHash, err := hchi.GetStringFromURL(r, "tx_id")
	if err != nil {
		return nil, errors.Wrap(err, "getting tx id")
	}

	addr, err := getAccountID(r, "account_id", requireAccountID)
	if err != nil {
		return nil, errors.Wrap(err, "getting account id")
	}

	return &showActionQueryParams{
		AccountID: addr,
		TxHash:    txHash,
	}, nil
}

// getAccountsIndexActionQueryParams gets the available query params for /accounts endpoints.
func getAccountsIndexActionQueryParams(r *http.Request) (*indexActionQueryParams, error) {
	signer, err := getSignerKey(r, "signer", true)
	if err != nil {
		return nil, errors.Wrap(err, "getting signer key")
	}

	pq, err := getAccountsPageQuery(r)
	if err != nil {
		return nil, errors.Wrap(err, "getting page query")
	}

	return &indexActionQueryParams{
		Signer:       signer,
		PagingParams: pq,
	}, nil
}

// getIndexActionQueryParams gets the available query params for all indexable endpoints.
func getIndexActionQueryParams(r *http.Request, ingestFailedTransactions bool) (*indexActionQueryParams, error) {
	addr, err := getAccountID(r, "account_id", false)
	if err != nil {
		return nil, errors.Wrap(err, "getting account id")
	}

	lid, err := getInt32ParamFromURL(r, "ledger_id")
	if err != nil {
		return nil, errors.Wrap(err, "getting ledger id")
	}

	// account_id and ledger_id are mutually excludesive.
	if addr != "" && lid != int32(0) {
		return nil, problem.BadRequest
	}

	pq, err := getPageQuery(r, false)
	if err != nil {
		return nil, errors.Wrap(err, "getting page query")
	}

	includeFailedTx, err := getBoolParamFromURL(r, "include_failed")
	if err != nil {
		return nil, errors.Wrap(err, "getting include_failed param")
	}
	if includeFailedTx && !ingestFailedTransactions {
		return nil, problem.MakeInvalidFieldProblem("include_failed",
			errors.New("`include_failed` parameter is unavailable when Aurora is not ingesting failed "+
				"transactions. Set `INGEST_FAILED_TRANSACTIONS=true` to start ingesting them."))
	}

	return &indexActionQueryParams{
		AccountID:        addr,
		LedgerID:         lid,
		PagingParams:     pq,
		IncludeFailedTxs: includeFailedTx,
	}, nil
}

// validateCursorWithinHistory first checks whether the cursor in the page
// param is valid basesd on the order then verifies whether the cursor is
// within history.
func validateCursorWithinHistory(pq db2.PageQuery) error {
	// an ascending query should never return a gone response:  An ascending query
	// prior to known history should return results at the beginning of history,
	// and an ascending query beyond the end of history should not error out but
	// rather return an empty page (allowing code that tracks the procession of
	// some resource more easily).
	if pq.Order != "desc" {
		return nil
	}

	var (
		cursor int64
		err    error
	)
	// cursor from effect streaming endpoint may contain the DefaultPairSep.
	if strings.Contains(pq.Cursor, db2.DefaultPairSep) {
		cursor, _, err = pq.CursorInt64Pair(db2.DefaultPairSep)
	} else {
		cursor, err = pq.CursorInt64()
	}
	if err != nil {
		return problem.MakeInvalidFieldProblem(actions.ParamCursor, errors.New("invalid value"))
	}

	elder := toid.New(ledger.CurrentState().HistoryElder, 0, 0)
	if cursor <= elder.ToInt64() {
		return &hProblem.BeforeHistory
	}

	return nil
}
