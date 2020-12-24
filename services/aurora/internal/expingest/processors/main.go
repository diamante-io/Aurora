package processors

import (
	"github.com/diamnet/go/exp/orderbook"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
)

type PipelineContextKey string

const (
	IngestUpdateDatabase = PipelineContextKey("IngestUpdateDatabase")
)

type DatabaseProcessorActionType string

const (
	AccountsForSigner DatabaseProcessorActionType = "AccountsForSigner"
	Offers            DatabaseProcessorActionType = "Offers"
)

// DatabaseProcessor is a processor (both state and ledger) that's responsible
// for persisting ledger data used in expingest in a database. It's possible
// to create multiple procesors of this type but they all should share the same
// *history.Q object to share a common transaction. `Action` defines what each
// processor is responsible for.
type DatabaseProcessor struct {
	HistoryQ history.QSigners
	OffersQ  history.QOffers
	Action   DatabaseProcessorActionType
}

// OrderbookProcessor is a processor (both state and ledger) that's responsible
// for updating orderbook graph with new/updated/removed offers. Orderbook graph
// can be later used for path finding.
type OrderbookProcessor struct {
	OrderBookGraph *orderbook.OrderBookGraph
}

// ContextFilter writes read objects only if a given key is present in the
// pipline context.
type ContextFilter struct {
	Key PipelineContextKey
}
