package main

import (
	"flag"
	"fmt"
	stdio "io"
	"log"

	"github.com/hcnet/go/exp/ingest/adapters"
	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/support/historyarchive"
)

func main() {
	ledgerPtr := flag.Uint64("ledger", 0, "`ledger to analyze` (tip: has to be of the form `ledger = 64*n - 1`, where n is > 0)")
	flag.Parse()
	var seqNum uint32 = uint32(*ledgerPtr)

	if seqNum == 0 {
		flag.Usage()
		return
	}

	archive, e := archive()
	if e != nil {
		panic(e)
	}
	haa := adapters.MakeHistoryArchiveAdapter(archive)

	sr, e := haa.GetState(seqNum, &io.MemoryTempSet{})
	if e != nil {
		panic(e)
	}

	accounts := map[string]bool{}
	var i uint64 = 0
	var count uint64 = 0
	for {
		le, e := sr.Read()
		if e != nil {
			panic(e)
		}
		if e == stdio.EOF {
			log.Printf("total seen %d entries of which %d were accounts", i, count)
			return
		}

		if ae, valid := le.State.Data.GetAccount(); valid {
			addr := ae.AccountId.Address()
			if _, exists := accounts[addr]; exists {
				log.Fatalf("error, total seen %d entries of which %d were unique accounts; repeated account: %s", i, count, addr)
			}

			accounts[addr] = true
			count += 1
		}
		i += 1

		if i%1000 == 0 {
			log.Printf("seen %d entries of which %d were accounts", i, count)
		}
	}
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
