package db

import (
	"database/sql"

	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/errors"
)

const (
	authorizedTransactionTableName = "authorized_transaction"
	allowedFITableName             = "allowed_fi"
	allowedUserTableName           = "allowed_user"
	authDataTableName              = "auth_data"
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

// InsertAuthorizedTransaction inserts a new authorized transaction into DB.
func (d *PostgresDatabase) InsertAuthorizedTransaction(transaction *AuthorizedTransaction) error {
	authorizedTransactionTable := d.getTable(authorizedTransactionTableName, nil)
	_, err := authorizedTransactionTable.Insert(transaction).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting authorized trasaction")
	}

	return nil
}

// GetAuthorizedTransactionByMemo returns authorized transaction searching by memo
func (d *PostgresDatabase) GetAuthorizedTransactionByMemo(memo string) (*AuthorizedTransaction, error) {
	authorizedTransactionTable := d.getTable(authorizedTransactionTableName, nil)
	var authorizedTransaction AuthorizedTransaction
	err := authorizedTransactionTable.Get(&authorizedTransaction, map[string]interface{}{"memo": memo}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting authorized transaction by memo")
		}
	}

	return &authorizedTransaction, nil
}

// InsertAllowedFI inserts a new allowed FI into DB.
func (d *PostgresDatabase) InsertAllowedFI(fi *AllowedFI) error {
	allowedFITable := d.getTable(allowedFITableName, nil)
	_, err := allowedFITable.Insert(fi).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting allowed FI")
	}

	return nil
}

// GetAllowedFIByDomain returns allowed FI by a domain
func (d *PostgresDatabase) GetAllowedFIByDomain(domain string) (*AllowedFI, error) {
	allowedFITable := d.getTable(allowedFITableName, nil)
	var allowedFI AllowedFI
	err := allowedFITable.Get(&allowedFI, map[string]interface{}{"domain": domain}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting allowed FI by domain")
		}
	}

	return &allowedFI, nil
}

// DeleteAllowedFIByDomain deletes allowed FI by a domain
func (d *PostgresDatabase) DeleteAllowedFIByDomain(domain string) error {
	allowedFITable := d.getTable(allowedFITableName, nil)
	_, err := allowedFITable.Delete(map[string]interface{}{"domain": domain}).Exec()
	return errors.Wrap(err, "Error removing allowed FI by domain")
}

// InsertAllowedUser inserts a new allowed user into DB.
func (d *PostgresDatabase) InsertAllowedUser(user *AllowedUser) error {
	allowedUserTable := d.getTable(allowedUserTableName, nil)
	_, err := allowedUserTable.Insert(user).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting allowed user")
	}

	return nil
}

// GetAllowedUserByDomainAndUserID returns allowed user by domain and userID
func (d *PostgresDatabase) GetAllowedUserByDomainAndUserID(domain, userID string) (*AllowedUser, error) {
	allowedUserTable := d.getTable(allowedUserTableName, nil)
	var allowedUser AllowedUser
	err := allowedUserTable.Get(&allowedUser, map[string]interface{}{"fi_domain": domain, "user_id": userID}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting allowed user by domain and user ID")
		}
	}

	return &allowedUser, nil
}

// DeleteAllowedUserByDomainAndUserID deletes allowed user by domain and userID
func (d *PostgresDatabase) DeleteAllowedUserByDomainAndUserID(domain, userID string) error {
	allowedUserTable := d.getTable(allowedUserTableName, nil)
	_, err := allowedUserTable.Delete(map[string]interface{}{"fi_domain": domain, "user_id": userID}).Exec()
	return errors.Wrap(err, "Error removing allowed user by domain and userID")
}

// InsertAuthData inserts a new auth data into DB.
func (d *PostgresDatabase) InsertAuthData(authData *AuthData) error {
	authDataTable := d.getTable(authDataTableName, nil)
	_, err := authDataTable.Insert(authData).IgnoreCols("id").Exec()
	if err != nil {
		return errors.Wrap(err, "Error inserting auth data")
	}

	return nil
}

// GetAuthData gets auth data by request ID
func (d *PostgresDatabase) GetAuthData(requestID string) (*AuthData, error) {
	authDataTable := d.getTable(authDataTableName, nil)
	var authData AuthData
	err := authDataTable.Get(&authData, map[string]interface{}{"request_id": requestID}).Exec()
	if err != nil {
		switch errors.Cause(err) {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, errors.Wrap(err, "Error getting auth data by request ID")
		}
	}

	return &authData, nil
}
