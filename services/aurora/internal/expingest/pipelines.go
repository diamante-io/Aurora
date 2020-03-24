package expingest

import (
	"context"
	"fmt"

	"github.com/hcnet/go/exp/ingest"
	"github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/ingest/processors"
	"github.com/hcnet/go/exp/orderbook"
	supportPipeline "github.com/hcnet/go/exp/support/pipeline"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	auroraProcessors "github.com/hcnet/go/services/aurora/internal/expingest/processors"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/errors"
	ilog "github.com/hcnet/go/support/log"
	"github.com/hcnet/go/xdr"
)

type pType string

const (
	statePipeline  pType = "state_pipeline"
	ledgerPipeline pType = "ledger_pipeline"
)

func accountForSignerStateNode(q *history.Q) *supportPipeline.PipelineNode {
	return pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeAccount}).
		Pipe(
			pipeline.StateNode(&auroraProcessors.DatabaseProcessor{
				HistoryQ: q,
				Action:   auroraProcessors.AccountsForSigner,
			}),
		)
}

func orderBookDBStateNode(q *history.Q) *supportPipeline.PipelineNode {
	return pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeOffer}).
		Pipe(
			pipeline.StateNode(&auroraProcessors.DatabaseProcessor{
				OffersQ: q,
				Action:  auroraProcessors.Offers,
			}),
		)
}

func orderBookGraphStateNode(graph *orderbook.OrderBookGraph) *supportPipeline.PipelineNode {
	return pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeOffer}).
		Pipe(
			pipeline.StateNode(&auroraProcessors.OrderbookProcessor{
				OrderBookGraph: graph,
			}),
		)
}

func buildStatePipeline(historyQ *history.Q, graph *orderbook.OrderBookGraph) *pipeline.StatePipeline {
	statePipeline := &pipeline.StatePipeline{}

	statePipeline.SetRoot(
		pipeline.StateNode(&processors.RootProcessor{}).
			Pipe(
				accountForSignerStateNode(historyQ),
				orderBookDBStateNode(historyQ),
				orderBookGraphStateNode(graph),
			),
	)

	return statePipeline
}

func accountForSignerLedgerNode(q *history.Q) *supportPipeline.PipelineNode {
	return pipeline.LedgerNode(&auroraProcessors.DatabaseProcessor{
		HistoryQ: q,
		Action:   auroraProcessors.AccountsForSigner,
	})
}

func orderBookDBLedgerNode(q *history.Q) *supportPipeline.PipelineNode {
	return pipeline.LedgerNode(&auroraProcessors.DatabaseProcessor{
		OffersQ: q,
		Action:  auroraProcessors.Offers,
	})
}

func orderBookGraphLedgerNode(graph *orderbook.OrderBookGraph) *supportPipeline.PipelineNode {
	return pipeline.LedgerNode(&auroraProcessors.OrderbookProcessor{
		OrderBookGraph: graph,
	})
}

func buildLedgerPipeline(historyQ *history.Q, graph *orderbook.OrderBookGraph) *pipeline.LedgerPipeline {
	ledgerPipeline := &pipeline.LedgerPipeline{}

	ledgerPipeline.SetRoot(
		pipeline.LedgerNode(&processors.RootProcessor{}).
			Pipe(
				// This subtree will only run when `IngestUpdateDatabase` is set.
				pipeline.LedgerNode(&auroraProcessors.ContextFilter{auroraProcessors.IngestUpdateDatabase}).
					Pipe(
						accountForSignerLedgerNode(historyQ),
						orderBookDBLedgerNode(historyQ),
					),
				orderBookGraphLedgerNode(graph),
			),
	)

	return ledgerPipeline
}

func addPipelineHooks(
	p supportPipeline.PipelineInterface,
	historySession *db.Session,
	ingestSession ingest.Session,
	graph *orderbook.OrderBookGraph,
) {
	var pipelineType pType
	switch p.(type) {
	case *pipeline.StatePipeline:
		pipelineType = statePipeline
	case *pipeline.LedgerPipeline:
		pipelineType = ledgerPipeline
	default:
		panic(fmt.Sprintf("Unknown pipeline type %T", p))
	}

	historyQ := &history.Q{historySession}

	p.AddPreProcessingHook(func(ctx context.Context) (context.Context, error) {
		// Start a transaction only if not in a transaction already.
		// The only case this can happen is during the first run when
		// a transaction is started to get the latest ledger `FOR UPDATE`
		// in `System.Run()`.
		if tx := historySession.GetTx(); tx == nil {
			err := historySession.Begin()
			if err != nil {
				return ctx, errors.Wrap(err, "Error starting a transaction")
			}
		}

		// We need to get this value `FOR UPDATE` so all other instances
		// are blocked.
		lastIngestedLedger, err := historyQ.GetLastLedgerExpIngest()
		if err != nil {
			return ctx, errors.Wrap(err, "Error getting last ledger")
		}

		ledgerSeq := pipeline.GetLedgerSequenceFromContext(ctx)

		updateDatabase := false
		if pipelineType == statePipeline {
			// State pipeline is always fully run because loading offers
			// from a database is done outside the pipeline.
			updateDatabase = true
		} else {
			if lastIngestedLedger+1 == ledgerSeq {
				// lastIngestedLedger+1 == ledgerSeq what means that this instance
				// is the main ingesting instance in this round and should update a
				// database.
				updateDatabase = true
				ctx = context.WithValue(ctx, auroraProcessors.IngestUpdateDatabase, true)
			}
		}

		// If we are not going to update a DB release a lock by rolling back a
		// transaction.
		if !updateDatabase {
			historySession.Rollback()
		}

		log.WithFields(ilog.F{
			"ledger":            ledgerSeq,
			"type":              pipelineType,
			"updating_database": updateDatabase,
		}).Info("Processing ledger")

		return ctx, nil
	})

	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		defer historySession.Rollback()
		defer graph.Discard()

		ledgerSeq := pipeline.GetLedgerSequenceFromContext(ctx)

		if err != nil {
			log.
				WithFields(ilog.F{
					"ledger": ledgerSeq,
					"type":   pipelineType,
					"err":    err,
				}).
				Error("Error processing ledger")
			return err
		}

		if tx := historySession.GetTx(); tx != nil {
			// If we're in a transaction we're updating database with new data.
			// We get lastIngestedLedger from a DB here to do an extra check
			// if the current node should really be updating a DB.
			// This is "just in case" if lastIngestedLedger is not selected
			// `FOR UPDATE` due to a bug or accident. In such case we error and
			// rollback.
			lastIngestedLedger, err := historyQ.GetLastLedgerExpIngest()
			if err != nil {
				return errors.Wrap(err, "Error getting last ledger")
			}

			if lastIngestedLedger != 0 && lastIngestedLedger+1 != ledgerSeq {
				return errors.New("The local latest sequence is not equal to global sequence + 1")
			}

			if err := historyQ.UpdateLastLedgerExpIngest(ledgerSeq); err != nil {
				return errors.Wrap(err, "Error updating last ingested ledger")
			}

			if err := historyQ.UpdateExpIngestVersion(CurrentVersion); err != nil {
				return errors.Wrap(err, "Error updating expingest version")
			}

			if err := historySession.Commit(); err != nil {
				return errors.Wrap(err, "Error commiting db transaction")
			}
		}

		if err := graph.Apply(); err != nil {
			return errors.Wrap(err, "Error applying order book changes")
		}

		log.WithFields(ilog.F{"ledger": ledgerSeq, "type": pipelineType}).Info("Processed ledger")
		return nil
	})
}
