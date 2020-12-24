package db

import (
	"database/sql/driver"

	"github.com/diamnet/go/support/errors"
)

// Scan implements database/sql.Scanner interface
func (s *SentTransactionStatus) Scan(src interface{}) error {
	value, ok := src.(string)
	if !ok {
		return errors.New("Cannot convert value to SentTransactionStatus")
	}
	*s = SentTransactionStatus(value)
	return nil
}

// Value implements driver.Valuer
func (status SentTransactionStatus) Value() (driver.Value, error) {
	return driver.Value(string(status)), nil
}

var _ driver.Valuer = SentTransactionStatus("")
