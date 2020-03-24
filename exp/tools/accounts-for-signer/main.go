package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hcnet/go/exp/ingest"
	"github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/ingest/processors"
	"github.com/hcnet/go/support/historyarchive"
	"github.com/hcnet/go/xdr"
)

func main() {
	archive, err := archive()
	if err != nil {
		panic(err)
	}

	statePipeline := &pipeline.StatePipeline{}
	statePipeline.SetRoot(
		// Passes accounts only
		pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeAccount}).
			Pipe(
				// Finds accounts for a single signer
				pipeline.StateNode(&AccountsForSignerProcessor{Signer: "GBMALBYJT6A73SYQWOWVVCGSPUPJPBX4AFDJ7A63GG64QCNRCAFYWWEN"}).
					Pipe(pipeline.StateNode(&processors.CSVPrinter{Filename: "./accounts_for_signer.csv"})),
			),
	)

	session := &ingest.SingleLedgerSession{
		Archive:       archive,
		StatePipeline: statePipeline,
	}

	doneStats := printPipelineStats(statePipeline)

	err = session.Run()
	if err != nil {
		fmt.Println("Session errored:")
		fmt.Println(err)
	} else {
		fmt.Println("Session finished without errors")
	}

	time.Sleep(10 * time.Second)
	doneStats <- true
	time.Sleep(10 * time.Second)
	// Print go routines count for the last time
	fmt.Printf("Goroutines = %v\n", runtime.NumGoroutine())
}

func archive() (*historyarchive.Archive, error) {
	return historyarchive.Connect(
		fmt.Sprintf("s3://history.hcnet.org/prd/core-live/core_live_001/"),
		historyarchive.ConnectOptions{
			S3Region:         "eu-west-1",
			UnsignedRequests: true,
		},
	)
}

func printPipelineStats(p *pipeline.StatePipeline) chan<- bool {
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
			fmt.Printf("\tNumCPU = %v\n\n", runtime.NumCPU())

			fmt.Printf("Duration: %s\n", time.Since(startTime))
			fmt.Println("Pipeline status:")
			p.PrintStatus()

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
