package aurora

import (
	"net/http"
)

func (action AssetsAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action DataShowAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action EffectIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action FixedPathIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action LedgerIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action LedgerShowAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action MetricsAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action NotFoundAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action NotImplementedAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action OffersByAccountAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action OperationFeeStatsAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action OperationIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action OperationShowAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action OrderBookShowAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action PathIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action RateLimitExceededAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action RootAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action TradeAggregateIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action TradeIndexAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}

func (action TransactionCreateAction) Handle(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	ap.Execute(&action)
}
