package aurora

import (
	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/render/hal"
)

// This file contains the actions:
//
// LedgerIndexAction: pages of ledgers
// LedgerShowAction: single ledger by sequence

// Interface verifications
var _ actions.JSONer = (*LedgerIndexAction)(nil)
var _ actions.EventStreamer = (*LedgerIndexAction)(nil)

// LedgerIndexAction renders a page of ledger resources, identified by
// a normal page query.
type LedgerIndexAction struct {
	Action
	PagingParams db2.PageQuery
	Records      []history.Ledger
	Page         hal.Page
}

// JSON is a method for actions.JSON
func (action *LedgerIndexAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

// SSE is a method for actions.SSE
func (action *LedgerIndexAction) SSE(stream *sse.Stream) error {
	action.Setup(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
	)
	action.Do(
		action.loadRecords,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			records := action.Records[stream.SentCount():]
			for _, record := range records {
				var res aurora.Ledger
				resourceadapter.PopulateLedger(action.R.Context(), &res, record)
				stream.Send(sse.Event{ID: res.PagingToken(), Data: res})
			}
		},
	)

	return action.Err
}

func (action *LedgerIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.PagingParams = action.GetPageQuery()
}

func (action *LedgerIndexAction) loadRecords() {
	action.Err = action.HistoryQ().Ledgers().Page(action.PagingParams).Select(&action.Records)
}

func (action *LedgerIndexAction) loadPage() {
	for _, record := range action.Records {
		var res aurora.Ledger
		resourceadapter.PopulateLedger(action.R.Context(), &res, record)
		action.Page.Add(res)
	}

	action.Page.FullURL = action.FullURL()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}

// Interface verification
var _ actions.JSONer = (*LedgerShowAction)(nil)

// LedgerShowAction renders a ledger found by its sequence number.
type LedgerShowAction struct {
	Action
	Sequence int32
	Record   history.Ledger
}

// JSON is a method for actions.JSON
func (action *LedgerShowAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.verifyWithinHistory,
		action.loadRecord,
		func() {
			var res aurora.Ledger
			resourceadapter.PopulateLedger(action.R.Context(), &res, action.Record)
			hal.Render(action.W, res)
		},
	)
	return action.Err
}

func (action *LedgerShowAction) loadParams() {
	action.Sequence = action.GetInt32("ledger_id")
}

func (action *LedgerShowAction) loadRecord() {
	action.Err = action.HistoryQ().LedgerBySequence(&action.Record, action.Sequence)
}

func (action *LedgerShowAction) verifyWithinHistory() {
	if action.Sequence < ledger.CurrentState().HistoryElder {
		action.Err = &problem.BeforeHistory
	}
}
