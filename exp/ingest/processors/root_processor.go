package processors

import (
	"context"
	stdio "io"

	"github.com/hcnet/go/exp/ingest/io"
	ingestpipeline "github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/support/pipeline"
)

func (p *RootProcessor) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
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

		err = w.Write(entryChange)
		if err != nil {
			if err == stdio.ErrClosedPipe {
				return nil
			}
			return err
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

func (p *RootProcessor) ProcessLedger(ctx context.Context, store *pipeline.Store, r io.LedgerReader, w io.LedgerWriter) error {
	defer r.Close()
	defer w.Close()

	for {
		transaction, err := r.Read()
		if err != nil {
			if err == stdio.EOF {
				break
			} else {
				return err
			}
		}

		err = w.Write(transaction)
		if err != nil {
			if err == stdio.ErrClosedPipe {
				return nil
			}
			return err
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

func (p *RootProcessor) Name() string {
	return "RootProcessor"
}

var _ ingestpipeline.StateProcessor = &RootProcessor{}
var _ ingestpipeline.LedgerProcessor = &RootProcessor{}
