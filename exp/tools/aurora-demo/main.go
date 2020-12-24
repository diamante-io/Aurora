package main

import (
	"fmt"

	"github.com/diamnet/go/clients/diamnetcore"
	"github.com/diamnet/go/exp/ingest"
	"github.com/diamnet/go/exp/ingest/io"
	"github.com/diamnet/go/exp/ingest/ledgerbackend"
	"github.com/diamnet/go/exp/orderbook"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/historyarchive"
)

func main() {
	dsn := "postgres://localhost:5432/aurorademo?sslmode=disable"
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ledgerBackend, err := ledgerbackend.NewDatabaseBackend("postgres://localhost:5432/core?sslmode=disable")
	if err != nil {
		panic(err)
	}

	orderBookGraph := orderbook.NewOrderBookGraph()

	session := &ingest.LiveSession{
		Archive:       archive(),
		LedgerBackend: ledgerBackend,
		DiamNetCoreClient: &diamnetcore.Client{
			URL: "http://localhost:11620",
		},

		StatePipeline: buildStatePipeline(db, orderBookGraph),
		// logs every 50,000 state entries
		StateReporter:  NewLoggingStateReporter(50000),
		LedgerPipeline: buildLedgerPipeline(db, orderBookGraph),
		TempSet:        &io.PostgresTempSet{DSN: dsn},
		LedgerReporter: NewLoggingLedgerReporter(),
	}

	addPipelineHooks(session.StatePipeline, db, session, orderBookGraph)
	addPipelineHooks(session.LedgerPipeline, db, session, orderBookGraph)

	printPipelinesStats(session.StatePipeline, session.LedgerPipeline)

	// This is broken when the last ledger does not contain transactions
	// but it's just a demo (we don't store ledgers, just transactions).
	ledger, err := db.GetLatestLedger()
	if err != nil && !db.NoRows(errors.Cause(err)) {
		panic(err)
	}

	if ledger == 0 {
		err = session.Run()
	} else {
		err = session.Resume(ledger + 1)
	}

	if err != nil {
		panic(err)
	}
}

func archive() *historyarchive.Archive {
	a, err := historyarchive.Connect(
		fmt.Sprintf("s3://history.diamnet.org/prd/core-live/core_live_001/"),
		historyarchive.ConnectOptions{
			S3Region:         "eu-west-1",
			UnsignedRequests: true,
		},
	)
	if err != nil {
		panic(err)
	}
	return a
}
