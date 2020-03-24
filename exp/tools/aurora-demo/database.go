package main

import (
	"database/sql"
	"encoding/hex"

	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/xdr"
)

/* Schema:

CREATE TABLE accounts_signers (
    account character varying(64),
    signer character varying(64)
);

ALTER TABLE ONLY accounts_signers
    ADD CONSTRAINT account_signer_unique UNIQUE (account, signer);

CREATE INDEX signers_index ON accounts_signers USING btree (signer);

CREATE TABLE transactions (
    hash character varying(64) NOT NULL,
    success boolean,
    ledger integer NOT NULL,
    source character varying(64) NOT NULL,
    fee_paid integer,
    operations integer
);

ALTER TABLE ONLY transactions
    ADD CONSTRAINT hash_unique UNIQUE (hash);

CREATE INDEX ledger_index ON transactions USING btree (ledger);

CREATE INDEX hash_index ON transactions USING btree (hash);

*/

type Database struct {
	*db.Session
}

type transaction struct {
	Hash       string `db:"hash"`
	Success    bool   `db:"success"`
	Ledger     uint32 `db:"ledger"`
	Source     string `db:"source"`
	FeePaid    uint32 `db:"fee_paid"`
	Operations int    `db:"operations"`
}

type accountSigner struct {
	Account string `db:"account"`
	Signer  string `db:"signer"`
}

func NewDatabase(uri string) (*Database, error) {
	session, err := db.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	db := &Database{
		Session: session,
	}

	return db, nil
}

// InsertTransaction adds a new transaction to the database. It probably
// shouldn't use `io.LedgerTransaction` but it's just a demo code.
func (d *Database) InsertTransaction(ledger uint32, tx io.LedgerTransaction) (sql.Result, error) {
	row := &transaction{
		Hash:       hex.EncodeToString(tx.Result.TransactionHash[:]),
		Success:    tx.Result.Result.Result.Code == xdr.TransactionResultCodeTxSuccess,
		Ledger:     ledger,
		Source:     tx.Envelope.Tx.SourceAccount.Address(),
		FeePaid:    uint32(tx.Envelope.Tx.Fee),
		Operations: len(tx.Envelope.Tx.Operations),
	}
	return d.Session.GetTable("transactions").Insert(row).Exec()
}

// InsertAccountSigner adds a account signer pair to the database.
func (d *Database) InsertAccountSigner(account, signer string) (sql.Result, error) {
	row := &accountSigner{
		Account: account,
		Signer:  signer,
	}
	return d.Session.GetTable("accounts_signers").Insert(row).Exec()
}

// RemoveAccountSigner removes account signer pair.
func (d *Database) RemoveAccountSigner(account, signer string) (sql.Result, error) {
	return d.Session.GetTable("accounts_signers").
		Delete("account = ? and signer = ?", account, signer).Exec()
}

// GetLatestLedger returns the last processed ledger.
func (d *Database) GetLatestLedger() (uint32, error) {
	var tx transaction
	err := d.Session.GetTable("transactions").Get(&tx, "1=1").OrderBy("ledger DESC").Exec()
	return tx.Ledger, err

}
