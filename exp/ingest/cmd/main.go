package main

import (
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/diamnet/go/exp/ingest/ledgerbackend"

	ingestadapters "github.com/diamnet/go/exp/ingest/adapters"
)

const dbURI = "postgres://diamnet:postgres@localhost:8002/core"

func main() {
	// useBackend()
	useAdapter()
}

// Demos use of the LedgerBackendAdapter
func useAdapter() {
	backend, err := ledgerbackend.NewDatabaseBackend(dbURI)
	if err != nil {
		log.Fatal(err)
	}

	lba := ingestadapters.LedgerBackendAdapter{
		Backend: backend,
	}

	ledgerSequence, err := lba.GetLatestLedgerSequence()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Latest ledger =", ledgerSequence)

	lrc, err := lba.GetLedger(ledgerSequence)
	if err != nil {
		log.Fatal(err)
	}

	s := lrc.GetSequence()
	fmt.Println("lrc sequence:", s)

	h := lrc.GetHeader()
	fmt.Println("lrc header:", h)

	for {
		lt, err := lrc.Read()
		if err != nil {
			if err == io.EOF {
				log.Info("No more transactions to read")
				break
			}
			log.Fatal(err)
		}

		d := lt.Index
		fmt.Println("\nIndex:", d)

		a := lt.Result
		fmt.Println("TransactionResultPair:")
		fmt.Println("    fee charged", a.Result.FeeCharged)
		fmt.Println("    ext", a.Result.Ext)
		fmt.Println("    result", a.Result.Result)

		b := lt.Envelope
		fmt.Println("TransactionEnvelope:")
		fmt.Println("b", b.Tx)

		c := lt.Meta
		fmt.Println("TransactionMeta:")
		fmt.Println("    operations", c.Operations)
		fmt.Println("    V", c.V)
		fmt.Println("    V1.Operations", c.V1.Operations)
		fmt.Println("    V1.TxChanges", c.V1.TxChanges)
	}

	log.Infof("latest ledger is %d, closed at %s (%d)", ledgerSequence,
		time.Unix(int64(h.Header.ScpValue.CloseTime), 0), h.Header.ScpValue.CloseTime)

	lrc.Close()
	lba.Close()
}

// Demos direct use of the DatabaseBackend
// func useBackend() {
// 	dbb := ledgerbackend.DatabaseBackend{
// 		DataSourceName: dbURI,
// 	}
// 	defer dbb.Close()

// 	ledgerSequence, err := dbb.GetLatestLedgerSequence()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Latest ledger =", ledgerSequence)

// 	exists, ledgerCloseMeta, err := dbb.GetLedger(ledgerSequence)

// 	if err != nil {
// 		log.Fatal("error reading ledger from backend: ", err)
// 	}
// 	if !exists {
// 		log.Fatalf("Ledger %d was not found", ledgerSequence)
// 	}

// 	fmt.Println("N transactions =", len(ledgerCloseMeta.TransactionEnvelope))
// 	fmt.Println("ledgerCloseMeta.Transaction:", ledgerCloseMeta.TransactionEnvelope)

// 	fmt.Println("N transactionReults =", len(ledgerCloseMeta.TransactionResult))
// 	fmt.Println("ledgerCloseMeta.TransactionResults:", ledgerCloseMeta.TransactionResult)

// 	fmt.Println("N transactionMeta =", len(ledgerCloseMeta.TransactionMeta))
// 	fmt.Println("ledgerCloseMeta.TransactionMeta:", ledgerCloseMeta.TransactionMeta)

// 	a := ledgerCloseMeta.TransactionResult[0]
// 	fmt.Println("TransactionResultPair:", a)
// 	fmt.Println("    fee charged", a.Result.FeeCharged)
// 	fmt.Println("    ext", a.Result.Ext)
// 	fmt.Println("    result", a.Result.Result)

// 	b := ledgerCloseMeta.TransactionEnvelope[0]
// 	fmt.Println("TransactionEnvelope:", b)
// 	fmt.Println("b", b.Tx)

// 	c := ledgerCloseMeta.TransactionMeta[0]
// 	fmt.Println("TransactionMeta", c.Operations)
// 	fmt.Println("    operations", c.Operations)
// 	fmt.Println("    V", c.V)
// 	fmt.Println("    V1.Operations", c.V1.Operations)
// 	fmt.Println("    V1.TxChanges", c.V1.TxChanges)
// }
