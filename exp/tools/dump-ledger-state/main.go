package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/diamnet/go/exp/ingest"
	"github.com/diamnet/go/exp/ingest/io"
	"github.com/diamnet/go/exp/ingest/pipeline"
	"github.com/diamnet/go/exp/ingest/processors"
	"github.com/diamnet/go/support/historyarchive"
	"github.com/diamnet/go/xdr"
)

func main() {
	dsn := "postgres://localhost:5432/aurorademo?sslmode=disable"

	archive, err := archive()
	if err != nil {
		panic(err)
	}

	statePipeline := &pipeline.StatePipeline{}
	statePipeline.SetRoot(
		pipeline.StateNode(&processors.RootProcessor{}).
			Pipe(
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeAccount}).
					Pipe(pipeline.StateNode(&processors.CSVPrinter{Filename: "./accounts.csv"})),
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeData}).
					Pipe(pipeline.StateNode(&processors.CSVPrinter{Filename: "./accountdata.csv"})),
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeOffer}).
					Pipe(pipeline.StateNode(&processors.CSVPrinter{Filename: "./offers.csv"})),
				pipeline.StateNode(&processors.EntryTypeFilter{Type: xdr.LedgerEntryTypeTrustline}).
					Pipe(pipeline.StateNode(&processors.CSVPrinter{Filename: "./trustlines.csv"})),
			),
	)

	session := &ingest.SingleLedgerSession{
		LedgerSequence: 25154239,
		Archive:        archive,
		StatePipeline:  statePipeline,
		TempSet:        &io.PostgresTempSet{DSN: dsn},
	}

	doneStats := printPipelineStats(statePipeline)

	err = session.Run()
	if err != nil {
		fmt.Println("Session errored:")
		fmt.Println(err)
	} else {
		fmt.Println("Session finished without errors")
	}

	// Remove sorted files
	sortedFiles := []string{
		"./accounts_sorted.csv",
		"./accountdata_sorted.csv",
		"./offers_sorted.csv",
		"./trustlines_sorted.csv",
	}
	for _, file := range sortedFiles {
		err := os.Remove(file)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(10 * time.Second)
	doneStats <- true
	time.Sleep(10 * time.Second)
	// Print go routines count for the last time
	fmt.Printf("Goroutines = %v\n", runtime.NumGoroutine())
}

func archive() (*historyarchive.Archive, error) {
	return historyarchive.Connect(
		fmt.Sprintf("s3://history.diamnet.org/prd/core-live/core_live_001/"),
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
