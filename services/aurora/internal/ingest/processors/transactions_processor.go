package processors

import (
	"context"
	"github.com/diamnet/go/ingest"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/support/errors"
)

type TransactionProcessor struct {
	transactionsQ history.QTransactions
	sequence      uint32
	batch         history.TransactionBatchInsertBuilder
}

func NewTransactionProcessor(transactionsQ history.QTransactions, sequence uint32) *TransactionProcessor {
	return &TransactionProcessor{
		transactionsQ: transactionsQ,
		sequence:      sequence,
		batch:         transactionsQ.NewTransactionBatchInsertBuilder(maxBatchSize),
	}
}

func (p *TransactionProcessor) ProcessTransaction(ctx context.Context, transaction ingest.LedgerTransaction) error {
	if err := p.batch.Add(ctx, transaction, p.sequence); err != nil {
		return errors.Wrap(err, "Error batch inserting transaction rows")
	}

	return nil
}

func (p *TransactionProcessor) Commit(ctx context.Context) error {
	if err := p.batch.Exec(ctx); err != nil {
		return errors.Wrap(err, "Error flushing transaction batch")
	}

	return nil
}
