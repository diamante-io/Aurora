package aurora

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/actions"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	hProblem "github.com/hcnet/go/services/aurora/internal/render/problem"
	"github.com/hcnet/go/services/aurora/internal/render/sse"
	"github.com/hcnet/go/services/aurora/internal/resourceadapter"
	"github.com/hcnet/go/services/aurora/internal/toid"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
	"github.com/hcnet/go/support/render/httpjson"
	"github.com/hcnet/go/support/render/problem"
)

// Action is the "base type" for all actions in aurora.  It provides
// structs that embed it with access to the App struct.
//
// Additionally, this type is a trigger for go-codegen and causes
// the file at Action.tmpl to be instantiated for each struct that
// embeds Action.
type Action struct {
	actions.Base
	App *App
	Log *log.Entry

	hq *history.Q
	cq *core.Q
}

// CoreQ provides access to queries that access the hcnet core database.
func (action *Action) CoreQ() *core.Q {
	if action.cq == nil {
		action.cq = &core.Q{Session: action.App.CoreSession(action.R.Context())}
	}

	return action.cq
}

// HistoryQ provides access to queries that access the history portion of
// aurora's database.
func (action *Action) HistoryQ() *history.Q {
	if action.hq == nil {
		action.hq = &history.Q{Session: action.App.AuroraSession(action.R.Context())}
	}

	return action.hq
}

// Prepare sets the action's App field based upon the context
func (action *Action) Prepare(w http.ResponseWriter, r *http.Request) {
	base := &action.Base
	action.App = AppFromContext(r.Context())
	base.Prepare(w, r, action.App.ctx, action.App.config.SSEUpdateFrequency)
	if action.R.Context() != nil {
		action.Log = log.Ctx(action.R.Context())
	} else {
		action.Log = log.DefaultLogger
	}
}

// ValidateCursorAsDefault ensures that the cursor parameter is valid in the way
// it is normally used, i.e. it is either the string "now" or a string of
// numerals that can be parsed as an int64.
func (action *Action) ValidateCursorAsDefault() {
	if action.Err != nil {
		return
	}

	if action.GetString(actions.ParamCursor) == "now" {
		return
	}

	action.GetInt64(actions.ParamCursor)
}

// ValidateCursorWithinHistory compares the requested page of data against the
// ledger state of the history database.  In the event that the cursor is
// guaranteed to return no results, we return a 410 GONE http response.
func (action *Action) ValidateCursorWithinHistory() {
	if action.Err != nil {
		return
	}

	pq := action.GetPageQuery()
	if action.Err != nil {
		return
	}

	// an ascending query should never return a gone response:  An ascending query
	// prior to known history should return results at the beginning of history,
	// and an ascending query beyond the end of history should not error out but
	// rather return an empty page (allowing code that tracks the procession of
	// some resource more easily).
	if pq.Order != "desc" {
		return
	}

	var cursor int64
	var err error

	// HACK: checking for the presence of "-" to see whether we should use
	// CursorInt64 or CursorInt64Pair is gross.
	if strings.Contains(pq.Cursor, "-") {
		cursor, _, err = pq.CursorInt64Pair("-")
	} else {
		cursor, err = pq.CursorInt64()
	}

	if err != nil {
		action.SetInvalidField("cursor", errors.New("invalid value"))
		return
	}

	elder := toid.New(ledger.CurrentState().HistoryElder, 0, 0)

	if cursor <= elder.ToInt64() {
		action.Err = &hProblem.BeforeHistory
	}
}

// EnsureHistoryFreshness halts processing and raises
func (action *Action) EnsureHistoryFreshness() {
	if action.Err != nil {
		return
	}

	if action.App.IsHistoryStale() {
		ls := ledger.CurrentState()
		err := hProblem.StaleHistory
		err.Extras = map[string]interface{}{
			"history_latest_ledger": ls.HistoryLatest,
			"core_latest_ledger":    ls.CoreLatest,
		}
		action.Err = &err
	}
}

// FullURL returns the full url for this request
func (action *Action) FullURL() *url.URL {
	result := action.baseURL()
	result.Path = action.R.URL.Path
	result.RawQuery = action.R.URL.RawQuery
	return result
}

// baseURL returns the base url for this request, defined as a url containing
// the Host and Scheme portions of the request uri.
func (action *Action) baseURL() *url.URL {
	return httpx.BaseURL(action.R.Context())
}

// Fields of this struct are exported for json marshaling/unmarshaling in
// support/render/hal package.
type indexActionQueryParams struct {
	AccountID        string
	LedgerID         int32
	PagingParams     db2.PageQuery
	IncludeFailedTxs bool
	Signer           string
}

// Fields of this struct are exported for json marshaling/unmarshaling in
// support/render/hal package.
type showActionQueryParams struct {
	AccountID string
	TxHash    string
}

// getAccountInfo returns the information about an account based on the provided param.
func (w *web) getAccountInfo(ctx context.Context, qp *showActionQueryParams) (interface{}, error) {
	return actions.AccountInfo(ctx, &core.Q{w.coreSession(ctx)}, qp.AccountID)
}

// getAccountPage returns a page containing the account records.
func (w *web) getAccountPage(ctx context.Context, qp *indexActionQueryParams) (interface{}, error) {
	auroraSession, err := w.auroraSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting aurora db session")
	}

	return actions.AccountPage(ctx, &history.Q{auroraSession}, qp.Signer, qp.PagingParams)
}

// getTransactionPage returns a page containing the transaction records of an account or a ledger.
func (w *web) getTransactionPage(ctx context.Context, qp *indexActionQueryParams) (interface{}, error) {
	auroraSession, err := w.auroraSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting aurora db session")
	}

	return actions.TransactionPage(ctx, &history.Q{auroraSession}, qp.AccountID, qp.LedgerID, qp.IncludeFailedTxs, qp.PagingParams)
}

// getTransactionRecord returns a single transaction resource.
func (w *web) getTransactionResource(ctx context.Context, qp *showActionQueryParams) (interface{}, error) {
	auroraSession, err := w.auroraSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting aurora db session")
	}

	return actions.TransactionResource(ctx, &history.Q{auroraSession}, qp.TxHash)
}

// streamTransactions streams the transaction records of an account or a ledger.
func (w *web) streamTransactions(ctx context.Context, s *sse.Stream, qp *indexActionQueryParams) error {
	auroraSession, err := w.auroraSession(ctx)
	if err != nil {
		return errors.Wrap(err, "getting aurora db session")
	}

	return actions.StreamTransactions(ctx, s, &history.Q{auroraSession}, qp.AccountID, qp.LedgerID, qp.IncludeFailedTxs, qp.PagingParams)
}

// getOfferRecord returns a single offer resource.
func getOfferResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	offerID, err := getInt64ParamFromURL(r, "id")
	if err != nil {
		problem.Render(ctx, w, errors.Wrap(err, "couldn't parse offer id"))
		return
	}

	app := AppFromContext(ctx)
	record, err := app.HistoryQ().GetOfferByID(offerID)
	if err != nil {
		problem.Render(ctx, w, err)
		return
	}

	ledger := new(history.Ledger)
	err = app.HistoryQ().LedgerBySequence(
		ledger,
		int32(record.LastModifiedLedger),
	)
	if app.HistoryQ().NoRows(err) {
		ledger = nil
	} else if err != nil {
		problem.Render(ctx, w, err)
		return
	}

	var offerResponse aurora.Offer
	resourceadapter.PopulateHistoryOffer(ctx, &offerResponse, record, ledger)
	httpjson.Render(w, offerResponse, httpjson.HALJSON)
}
