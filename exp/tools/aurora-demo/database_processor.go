package main

import (
	"context"
	"fmt"
	stdio "io"

	"github.com/diamnet/go/exp/ingest/io"
	ingestpipeline "github.com/diamnet/go/exp/ingest/pipeline"
	"github.com/diamnet/go/exp/support/pipeline"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

type DatabaseProcessorActionType string

const (
	Transactions      DatabaseProcessorActionType = "Transactions"
	AccountsForSigner DatabaseProcessorActionType = "AccountsForSigner"
)

// DatabaseProcessor is a processor (both state and ledger) that's responsible
// for persisting ledger data used in aurora-demo in a database. It's possible
// to create multiple procesors of this type but they all should share the same
// Database object to share a common transaction. `Action` defines what each
// processor is responsible for.
type DatabaseProcessor struct {
	Database *Database
	Action   DatabaseProcessorActionType
}

func (p *DatabaseProcessor) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
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
			return errors.New("DatabaseProcessor requires LedgerEntryChangeTypeLedgerEntryState changes only")
		}

		switch p.Action {
		case AccountsForSigner:
			// We're interested in accounts only
			if entryChange.EntryType() != xdr.LedgerEntryTypeAccount {
				continue
			}

			accountEntry := entryChange.MustState().Data.MustAccount()
			account := accountEntry.AccountId.Address()
			for _, signer := range accountEntry.Signers {
				_, err = p.Database.InsertAccountSigner(account, signer.Key.Address())
				if err != nil {
					return errors.Wrap(err, "Error inserting account for signer")
				}
			}
		default:
			return errors.New("Unknown action")
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

func (p *DatabaseProcessor) ProcessLedger(ctx context.Context, store *pipeline.Store, r io.LedgerReader, w io.LedgerWriter) error {
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

		switch p.Action {
		case Transactions:
			_, err = p.Database.InsertTransaction(r.GetSequence(), transaction)
			if err != nil {
				return errors.Wrap(err, "Error inserting a transaction")
			}
		case AccountsForSigner:
			// TODO check if tx is success!
			err := p.processLedgerAccountsForSigner(transaction)
			if err != nil {
				return errors.Wrap(err, "Error in processLedgerAccountsForSigner")
			}
		default:
			return errors.New("Unknown action")
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

func (p *DatabaseProcessor) processLedgerAccountsForSigner(transaction io.LedgerTransaction) error {
	for _, change := range transaction.GetChanges() {
		if change.Type != xdr.LedgerEntryTypeAccount {
			continue
		}

		if !change.AccountSignersChanged() {
			continue
		}

		accountEntry := change.Pre.MustAccount()
		account := accountEntry.AccountId.Address()

		// This removes all Pre signers adds Post signers but can be
		// improved by finding a diff
		for _, signer := range change.Pre.MustAccount().Signers {
			_, err := p.Database.RemoveAccountSigner(account, signer.Key.Address())
			if err != nil {
				return errors.Wrap(err, "Error removing a signer")
			}
		}

		for _, signer := range change.Post.MustAccount().Signers {
			_, err := p.Database.InsertAccountSigner(account, signer.Key.Address())
			if err != nil {
				return errors.Wrap(err, "Error inserting a signer")
			}
		}
	}
	return nil
}

func (p *DatabaseProcessor) Name() string {
	return fmt.Sprintf("DatabaseProcessor (%s)", p.Action)
}

func (p *DatabaseProcessor) Reset() {}

var _ ingestpipeline.StateProcessor = &DatabaseProcessor{}
var _ ingestpipeline.LedgerProcessor = &DatabaseProcessor{}
