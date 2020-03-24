package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/hcnet/go/exp/ingest"
	"github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/ingest/processors"
	"github.com/hcnet/go/exp/orderbook"
	supportPipeline "github.com/hcnet/go/exp/support/pipeline"
	"github.com/hcnet/go/xdr"
)

func buildStatePipeline(db *Database, orderBookGraph *orderbook.OrderBookGraph) *pipeline.StatePipeline {
	statePipeline := &pipeline.StatePipeline{}

	statePipeline.SetRoot(
		pipeline.StateNode(&processors.RootProcessor{}).
			Pipe(
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeAccount}).
					Pipe(
						pipeline.StateNode(&DatabaseProcessor{
							Database: db,
							Action:   AccountsForSigner,
						}),
					),
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeOffer}).
					Pipe(
						pipeline.StateNode(&OrderbookProcessor{
							OrderBookGraph: orderBookGraph,
						}),
					),
			),
	)

	return statePipeline
}

func buildLedgerPipeline(db *Database, orderBookGraph *orderbook.OrderBookGraph) *pipeline.LedgerPipeline {
	ledgerPipeline := &pipeline.LedgerPipeline{}

	ledgerPipeline.SetRoot(
		pipeline.LedgerNode(&processors.RootProcessor{}).Pipe(
			pipeline.LedgerNode(&DatabaseProcessor{
				Database: db,
				Action:   AccountsForSigner,
			}),
			pipeline.LedgerNode(&DatabaseProcessor{
				Database: db,
				Action:   Transactions,
			}),
			pipeline.LedgerNode(&OrderbookProcessor{
				OrderBookGraph: orderBookGraph,
			}),
		),
	)

	return ledgerPipeline
}

func addPipelineHooks(
	p supportPipeline.PipelineInterface,
	db *Database,
	session ingest.Session,
	orderBookGraph *orderbook.OrderBookGraph,
) {
	p.AddPreProcessingHook(func(ctx context.Context) (context.Context, error) {
		ledgerSeq := pipeline.GetLedgerSequenceFromContext(ctx)
		fmt.Printf("%T Processing ledger: %d\n", p, ledgerSeq)
		return ctx, db.Begin()
	})

	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		ledgerSeq := pipeline.GetLedgerSequenceFromContext(ctx)

		if err != nil {
			fmt.Printf("%T Error processing ledger: %s\n", p, err)
			return db.Rollback()
		}

		// Acquire write lock
		session.UpdateLock()
		defer session.UpdateUnlock()

		// Run commits simultaneously
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			err = orderBookGraph.Apply()
			wg.Done()
		}()

		go func() {
			err = db.Commit()
			wg.Done()
		}()

		wg.Wait()
		if err != nil {
			return err
		}

		fmt.Printf("%T Processed ledger: %d\n", p, ledgerSeq)
		return nil
	})
}

func printPipelinesStats(state *pipeline.StatePipeline, ledger *pipeline.LedgerPipeline) chan<- bool {
	startTime := time.Now()
	done := make(chan bool)
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
			fmt.Printf("\tHeapAlloc = %v MiB", bToMb(m.HeapAlloc))
			fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
			fmt.Printf("\tNumGC = %v", m.NumGC)
			fmt.Printf("\tGoroutines = %v", runtime.NumGoroutine())
			fmt.Printf("\tNumCPU = %v", runtime.NumCPU())
			fmt.Printf("\tDuration = %s\n", time.Since(startTime))
			fmt.Println("----------------------------------------")

			fmt.Println("State pipeline status:")
			state.PrintStatus()
			fmt.Println("----------------------------------------")

			fmt.Println("Ledger pipeline status:")
			ledger.PrintStatus()
			fmt.Println("========================================")

			select {
			case <-ticker.C:
				continue
			case <-done:
				// Pipeline done
				return
			}
		}
	}()

	return done
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
