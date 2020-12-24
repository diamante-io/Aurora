package db

import (
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
	InsertAuthorizedTransaction(transaction *AuthorizedTransaction) error
	GetAuthorizedTransactionByMemo(memo string) (*AuthorizedTransaction, error)

	InsertAllowedFI(fi *AllowedFI) error
	GetAllowedFIByDomain(domain string) (*AllowedFI, error)
	DeleteAllowedFIByDomain(domain string) error

	InsertAllowedUser(user *AllowedUser) error
	GetAllowedUserByDomainAndUserID(domain, userID string) (*AllowedUser, error)
	DeleteAllowedUserByDomainAndUserID(domain, userID string) error

	InsertAuthData(authData *AuthData) error
	GetAuthData(requestID string) (*AuthData, error)
}

type PostgresDatabase struct {
	session *db.Session
}

// AllowedFI represents allowed FI
type AllowedFI struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Domain    string    `db:"domain"`
	PublicKey string    `db:"public_key"`
	AllowedAt time.Time `db:"allowed_at"`
}

// AllowedUser represents allowed user
type AllowedUser struct {
	ID          int64     `db:"id"`
	FiName      string    `db:"fi_name"`
	FiDomain    string    `db:"fi_domain"`
	FiPublicKey string    `db:"fi_public_key"`
	UserID      string    `db:"user_id"`
	AllowedAt   time.Time `db:"allowed_at"`
}

// AuthorizedTransaction represents authorized transaction
type AuthorizedTransaction struct {
	ID             int64     `db:"id"`
	TransactionID  string    `db:"transaction_id"`
	Memo           string    `db:"memo"`
	TransactionXdr string    `db:"transaction_xdr"`
	AuthorizedAt   time.Time `db:"authorized_at"`
	Data           string    `db:"data"`
}

// AuthData represents auth data
type AuthData struct {
	ID        int64  `db:"id"`
	RequestID string `db:"request_id"`
	Domain    string `db:"domain"`
	AuthData  string `db:"auth_data"`
}
