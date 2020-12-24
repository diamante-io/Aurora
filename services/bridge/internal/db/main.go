package db

import (
	"database/sql"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/diamnet/go/support/db"
)

//go:generate go-bindata -ignore .+\.go$ -pkg db -o bindata.go ./...

// Migrations represents all of the schema migration
var Migrations migrate.MigrationSource = &migrate.AssetMigrationSource{
	Asset:    Asset,
	AssetDir: AssetDir,
	Dir:      "migrations",
}

type Database interface {
	GetLastCursorValue() (cursor *string, err error)

	InsertReceivedPayment(payment *ReceivedPayment) error
	UpdateReceivedPayment(payment *ReceivedPayment) error
	GetReceivedPaymentByID(id int64) (*ReceivedPayment, error)
	GetReceivedPaymentByOperationID(operationID string) (*ReceivedPayment, error)
	GetReceivedPayments(page, limit uint64) ([]*ReceivedPayment, error)

	InsertSentTransaction(transaction *SentTransaction) error
	UpdateSentTransaction(transaction *SentTransaction) error
	GetSentTransactionByPaymentID(paymentID string) (*SentTransaction, error)
	GetSentTransactions(page, limit uint64) ([]*SentTransaction, error)
}

type PostgresDatabase struct {
	session *db.Session
}

// ReceivedPayment represents payment received by the gateway server
type ReceivedPayment struct {
	ID            int64     `db:"id" json:"id"`
	OperationID   string    `db:"operation_id" json:"operation_id"`
	ProcessedAt   time.Time `db:"processed_at" json:"processed_at"`
	PagingToken   string    `db:"paging_token" json:"paging_token"`
	Status        string    `db:"status" json:"status"`
	TransactionID string    `db:"transaction_id" json:"transaction_id"`
}

// SentTransactionStatus type represents sent transaction status
type SentTransactionStatus string

const (
	// SentTransactionStatusSending is a status indicating that transaction is sending
	SentTransactionStatusSending SentTransactionStatus = "sending"
	// SentTransactionStatusSuccess is a status indicating that transaction has been successfully sent
	SentTransactionStatusSuccess SentTransactionStatus = "success"
	// SentTransactionStatusFailure is a status indicating that there has been an error while sending a transaction
	SentTransactionStatusFailure SentTransactionStatus = "failure"
)

// SentTransaction represents transaction sent by the gateway server
type SentTransaction struct {
	ID            int64                 `db:"id" json:"id"`
	PaymentID     sql.NullString        `db:"payment_id" json:"payment_id"`
	TransactionID string                `db:"transaction_id" json:"transaction_id"`
	Status        SentTransactionStatus `db:"status" json:"status"` // sending/success/failure
	Source        string                `db:"source" json:"source"`
	SubmittedAt   time.Time             `db:"submitted_at" json:"submitted_at"`
	SucceededAt   *time.Time            `db:"succeeded_at" json:"succeeded_at"`
	Ledger        *int32                `db:"ledger" json:"ledger"`
	EnvelopeXdr   string                `db:"envelope_xdr" json:"envelope_xdr"`
	ResultXdr     *string               `db:"result_xdr" json:"result_xdr"`
}
