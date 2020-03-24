package processors

import (
	"context"
	"fmt"
	stdio "io"

	"github.com/hcnet/go/exp/ingest/io"
	ingestpipeline "github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/support/pipeline"
)

func (p *ContextFilter) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
	defer r.Close()
	defer w.Close()

	// Exit early if Key not found in `ctx`
	if v := ctx.Value(p.Key); v == nil {
		return nil
	}

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
				// Reader does not need more data
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

func (p *ContextFilter) ProcessLedger(ctx context.Context, store *pipeline.Store, r io.LedgerReader, w io.LedgerWriter) error {
	defer r.Close()
	defer w.Close()

	// Exit early if Key not found in `ctx`
	if v := ctx.Value(p.Key); v == nil {
		return nil
	}

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
				// Reader does not need more data
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

func (p *ContextFilter) Name() string {
	return fmt.Sprintf("ContextFilter (%s)", p.Key)
}

func (p *ContextFilter) Reset() {}

var _ ingestpipeline.StateProcessor = &ContextFilter{}
var _ ingestpipeline.LedgerProcessor = &ContextFilter{}
