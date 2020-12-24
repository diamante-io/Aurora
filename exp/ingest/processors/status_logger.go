package processors

import (
	"context"
	"fmt"
	stdio "io"

	"github.com/diamnet/go/exp/ingest/io"
	ingestpipeline "github.com/diamnet/go/exp/ingest/pipeline"
	"github.com/diamnet/go/exp/support/pipeline"
)

func (p *StatusLogger) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
	defer r.Close()
	defer w.Close()

	n := 0
	for {
		entry, err := r.Read()
		if err != nil {
			if err == stdio.EOF {
				break
			} else {
				return err
			}
		}

		n++

		if n%p.N == 0 {
			fmt.Printf("Read %d entries...\n", n)
		}

		err = w.Write(entry)
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

func (p *StatusLogger) Name() string {
	return fmt.Sprintf("StatusLogger (N=%d)", p.N)
}

var _ ingestpipeline.StateProcessor = &StatusLogger{}
