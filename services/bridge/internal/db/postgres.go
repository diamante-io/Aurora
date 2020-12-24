package db

import (
	"database/sql"

	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
)

const (
	receivedPaymentTableName = "received_payment"
	sentTransactionTableName = "sent_transaction"
)

func (d *PostgresDatabase) Open(dsn string) error {
	var err error
	d.session, err = db.Open("postgres", dsn)
	if err != nil {
		return err
	}

	return nil
}

func (d *PostgresDatabase) GetDB() *sql.DB {
	if d.session == nil {
		return nil
	}

	return d.session.DB.DB
}

func (d *PostgresDatabase) getTable(name string, session *db.Session) *db.Table {
	if session == nil {
		session = d.session
	}

	return &db.Table{
		Name:    name,
		Session: session,
	}
}

// InsertSentTransaction inserts anew transaction into DB. After successful insert ID
// field on `transaction` will updated to ID of a new row.
func (d *PostgresDatabase) InsertSentTransaction(transaction *SentTransaction) error {
	sentTransactionTable := d.getTable(sentTransactionTableName, nil)
	_, err := sentTransactionTable.Insert(transaction).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting sent transaction")
	}

	newTransaction, err := d.GetSentTransactionByHash(transaction.TransactionID)
	if err != nil {
		return errors.Wrap(err, "Error getting new transaction")
	}

	transaction.ID = newTransaction.ID
	return nil
}

func (d *PostgresDatabase) UpdateSentTransaction(transaction *SentTransaction) error {
	if transaction.ID == 0 {
		return errors.New("ID equals 0")
	}

	sentTransactionTable := d.getTable(sentTransactionTableName, nil)
	// TODO: something's wrong with db.Table.Update(). Setting the first argument does not work as expected.
	_, err := sentTransactionTable.Update(nil, map[string]interface{}{"id": transaction.ID}).
		SetStruct(transaction, []string{"id"}).
		Exec()
	if err != nil {
		return errors.Wrap(err, "Error updating sent transaction")
	}

	return nil
}

// InsertReceivedPayment inserts a new payment into DB. After successful insert ID
// field on `payment` will updated to ID of a new row.
func (d *PostgresDatabase) InsertReceivedPayment(payment *ReceivedPayment) error {
	receivedPaymentTable := d.getTable(receivedPaymentTableName, nil)
	_, err := receivedPaymentTable.Insert(payment).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting received payment")
	}

	newPayment, err := d.GetReceivedPaymentByOperationID(payment.OperationID)
	if err != nil {
		return errors.Wrap(err, "Error getting new operation")
	}

	payment.ID = newPayment.ID
	return nil
}

func (d *PostgresDatabase) UpdateReceivedPayment(payment *ReceivedPayment) error {
	if payment.ID == 0 {
		return errors.New("ID equals 0")
	}

	receivedPaymentTable := d.getTable(receivedPaymentTableName, nil)
	// TODO: something's wrong with db.Table.Update(). Setting the first argument does not work as expected.
	_, err := receivedPaymentTable.Update(nil, map[string]interface{}{"id": payment.ID}).
		SetStruct(payment, []string{"id"}).
		Exec()
	if err != nil {
		return errors.Wrap(err, "Error updating received payment")
	}

	return nil
}

// GetLastCursorValue returns last cursor value from a DB
func (d *PostgresDatabase) GetLastCursorValue() (cursor *string, err error) {
	receivedPayment, err := d.getLastReceivedPayment()
	if err != nil {
		return nil, err
	} else if receivedPayment == nil {
		return nil, nil
	} else {
		return &receivedPayment.PagingToken, nil
	}
}

// GetSentTransactionByPaymentID returns sent transaction searching by payment ID
func (d *PostgresDatabase) GetSentTransactionByPaymentID(paymentID string) (*SentTransaction, error) {
	sentTransactionTable := d.getTable(sentTransactionTableName, nil)
	var transaction SentTransaction
	err := sentTransactionTable.Get(&transaction, map[string]interface{}{"payment_id": paymentID}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting sent transaction by payment ID")
		}
	}

	return &transaction, nil
}

// GetSentTransactionByHash returns sent transaction searching by hash
func (d *PostgresDatabase) GetSentTransactionByHash(hash string) (*SentTransaction, error) {
	sentTransactionTable := d.getTable(sentTransactionTableName, nil)
	var transaction SentTransaction
	err := sentTransactionTable.Get(&transaction, map[string]interface{}{"transaction_id": hash}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting sent transaction by hash")
		}
	}

	return &transaction, nil
}

// GetReceivedPaymentByID returns received payment by id
func (d *PostgresDatabase) GetReceivedPaymentByID(id int64) (*ReceivedPayment, error) {
	return d.getReceivedPayment(map[string]interface{}{"id": id})
}

// GetReceivedPaymentByOperationID returns received payment by operation_id
func (d *PostgresDatabase) GetReceivedPaymentByOperationID(operationID string) (*ReceivedPayment, error) {
	return d.getReceivedPayment(map[string]interface{}{"operation_id": operationID})
}

func (d *PostgresDatabase) getReceivedPayment(params map[string]interface{}) (*ReceivedPayment, error) {
	receivedPaymentTable := d.getTable(receivedPaymentTableName, nil)
	var receivedPayment ReceivedPayment
	err := receivedPaymentTable.Get(&receivedPayment, params).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting received payment transaction by operation ID")
		}
	}

	return &receivedPayment, nil
}

// GetReceivedPayments returns received payments
func (d *PostgresDatabase) GetReceivedPayments(page, limit uint64) ([]*ReceivedPayment, error) {
	receivedPaymentTable := d.getTable(receivedPaymentTableName, nil)
	payments := []*ReceivedPayment{}

	if page == 0 {
		page = 1
	}

	offset := (page - 1) * limit

	err := receivedPaymentTable.Select(&payments, "1=1").Limit(limit).Offset(offset).OrderBy("id desc").Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return payments, nil
		default:
			return payments, errors.Wrap(err, "Error getting received payments")
		}
	}

	return payments, nil
}

// GetSentTransactions returns received payments
func (d *PostgresDatabase) GetSentTransactions(page, limit uint64) ([]*SentTransaction, error) {
	sentTransactionTable := d.getTable(sentTransactionTableName, nil)
	transactions := []*SentTransaction{}

	if page == 0 {
		page = 1
	}

	offset := (page - 1) * limit

	err := sentTransactionTable.Select(&transactions, "1=1").Limit(limit).Offset(offset).OrderBy("id desc").Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return transactions, nil
		default:
			return transactions, errors.Wrap(err, "Error getting sent transactions")
		}
	}

	return transactions, nil
}

// getLastReceivedPayment returns the last received payment
func (d *PostgresDatabase) getLastReceivedPayment() (*ReceivedPayment, error) {
	receivedPaymentTable := d.getTable(receivedPaymentTableName, nil)
	// TODO: `1=1`. We should be able to get row without WHERE clause.
	// When it's set to nil: `pq: syntax error at or near "ORDER""`
	var receivedPayment ReceivedPayment
	err := receivedPaymentTable.Get(&receivedPayment, "1=1").OrderBy("id DESC").Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting last received payment")
		}
	}

	return &receivedPayment, nil
}
