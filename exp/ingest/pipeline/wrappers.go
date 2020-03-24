package pipeline

import (
	"context"

	"github.com/hcnet/go/exp/ingest/io"
	supportPipeline "github.com/hcnet/go/exp/support/pipeline"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

func (w *stateProcessorWrapper) Process(ctx context.Context, store *supportPipeline.Store, reader supportPipeline.Reader, writer supportPipeline.Writer) error {
	return w.StateProcessor.ProcessState(
		ctx,
		store,
		&readerWrapperState{reader},
		&writerWrapperState{writer},
	)
}

func (w *ledgerProcessorWrapper) Process(ctx context.Context, store *supportPipeline.Store, reader supportPipeline.Reader, writer supportPipeline.Writer) error {
	return w.LedgerProcessor.ProcessLedger(
		ctx,
		store,
		&readerWrapperLedger{reader},
		&writerWrapperLedger{writer},
	)
}

func (w *stateReaderWrapper) GetContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, LedgerSequenceContextKey, w.StateReader.GetSequence())
	return ctx
}

func (w *stateReaderWrapper) Read() (interface{}, error) {
	return w.StateReader.Read()
}

func (w *readerWrapperState) GetSequence() uint32 {
	return GetLedgerSequenceFromContext(w.Reader.GetContext())
}

func (w *ledgerReaderWrapper) GetContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, LedgerSequenceContextKey, w.LedgerReader.GetSequence())
	ctx = context.WithValue(ctx, LedgerHeaderContextKey, w.LedgerReader.GetHeader())
	return ctx
}

func (w *ledgerReaderWrapper) Read() (interface{}, error) {
	return w.LedgerReader.Read()
}

func (w *readerWrapperState) Read() (xdr.LedgerEntryChange, error) {
	object, err := w.Reader.Read()
	if err != nil {
		return xdr.LedgerEntryChange{}, err
	}

	entry, ok := object.(xdr.LedgerEntryChange)
	if !ok {
		return xdr.LedgerEntryChange{}, errors.New("Read object is not xdr.LedgerEntryChange")
	}

	return entry, nil
}

func (w *readerWrapperLedger) GetSequence() uint32 {
	return GetLedgerSequenceFromContext(w.Reader.GetContext())
}

func (w *readerWrapperLedger) GetHeader() xdr.LedgerHeaderHistoryEntry {
	return GetLedgerHeaderFromContext(w.Reader.GetContext())
}

func (w *readerWrapperLedger) Read() (io.LedgerTransaction, error) {
	object, err := w.Reader.Read()
	if err != nil {
		return io.LedgerTransaction{}, err
	}

	entry, ok := object.(io.LedgerTransaction)
	if !ok {
		return io.LedgerTransaction{}, errors.New("Read object is not io.LedgerTransaction")
	}

	return entry, nil
}

func (w *writerWrapperState) Write(entry xdr.LedgerEntryChange) error {
	return w.Writer.Write(entry)
}

func (w *writerWrapperLedger) Write(entry io.LedgerTransaction) error {
	return w.Writer.Write(entry)
}
