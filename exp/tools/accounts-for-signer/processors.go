package main

import (
	"context"
	stdio "io"
	"sync"

	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/exp/support/pipeline"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

type SimpleProcessor struct {
	sync.Mutex
	callCount int
}

func (n *SimpleProcessor) IncrementAndReturnCallCount() int {
	n.Lock()
	defer n.Unlock()
	n.callCount++
	return n.callCount
}

func (n *SimpleProcessor) Reset() {
	n.callCount = 0
}

type AccountsForSignerProcessor struct {
	SimpleProcessor

	Signer string
}

func (p *AccountsForSignerProcessor) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
	defer r.Close()
	defer w.Close()

	for {
		entryChange, err := r.Read()
		if err != nil {
			if err == stdio.EOF {
				break
			} else {
				return err
			}
		}

		if entryChange.Type != xdr.LedgerEntryChangeTypeLedgerEntryState {
			return errors.New("CSVPrinter requires LedgerEntryChangeTypeLedgerEntryState changes only")
		}

		entry := entryChange.State

		for _, signer := range entry.Data.Account.Signers {
			if signer.Key.Address() == p.Signer {
				err := w.Write(entryChange)
				if err != nil {
					if err == stdio.ErrClosedPipe {
						// Reader does not need more data
						return nil
					}
					return err
				}
				break
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			continue
		}
	}

	return nil
}

func (p *AccountsForSignerProcessor) Name() string {
	return "AccountsForSignerProcessor"
}
