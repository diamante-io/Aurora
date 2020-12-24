// Package db is the base package for database access at diamnet.  It primarily
// exposes Session which is a lightweight wrapper around a *sqlx.DB that
// provides utility methods (See the repo tests for examples).
//
// In addition to the query methods, this package provides query logging and
// stateful transaction management.
//
// In addition to the lower-level access facilities, this package exposes a
// system to build queries more dynamically using the help of
// https://github.com/Masterminds/squirrel.  These builder method are access
// through the `Table` type.
package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/diamnet/go/support/errors"

	// Enable mysql
	_ "github.com/go-sql-driver/mysql"
	// Enable postgres
	_ "github.com/lib/pq"
)

// Conn represents a connection to a single database.
type Conn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Rebind(sql string) string
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
}

// DeleteBuilder is a helper struct used to construct sql queries of the DELETE
// variety.
type DeleteBuilder struct {
	Table *Table
	sql   squirrel.DeleteBuilder
}

// InsertBuilder is a helper struct used to construct sql queries of the INSERT
// variety.
// NOTE: InsertBuilder will use "zero" value of a type in case of nil pointer values.
// If you need to insert `NULL` use sql.Null* or build your own type that implements
// database/sql/driver.Valuer.
type InsertBuilder struct {
	Table *Table

	rows        []interface{}
	ignoredCols map[string]bool
	sql         squirrel.InsertBuilder
}

// GetBuilder is a helper struct used to construct sql queries of the SELECT
// variety.
type GetBuilder struct {
	Table *Table

	dest interface{}
	sql  squirrel.SelectBuilder
}

// SelectBuilder is a helper struct used to construct sql queries of the SELECT
// variety.
type SelectBuilder struct {
	Table *Table

	dest interface{}
	sql  squirrel.SelectBuilder
}

// UpdateBuilder is a helper struct used to construct sql queries of the UPDATE
// variety.
type UpdateBuilder struct {
	Table *Table

	source interface{}
	sql    squirrel.UpdateBuilder
}

// Session provides helper methods for making queries against `DB` and provides
// utilities such as automatic query logging and transaction management.  NOTE:
// A Session is designed to be lightweight and temporarily lived (usually
// request scoped) which is one reason it is acceptable for it to store a
// context.  It is not presently intended to cross goroutine boundaries and is
// not concurrency safe.
type Session struct {
	// DB is the database connection that queries should be executed against.
	DB *sqlx.DB

	// Ctx is the optional context in which the repo is operating under.
	Ctx context.Context

	tx *sqlx.Tx
}

// Table helps to build sql queries against a given table.  It logically
// represents a SQL table on the database that `Session` is connected to.
type Table struct {
	// Name is the name of the table
	Name string

	Session *Session
}

// Open the database at `dsn` and returns a new *Session using it.
func Open(dialect, dsn string) (*Session, error) {
	db, err := sqlx.Connect(dialect, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "connect failed")
	}

	return &Session{DB: db}, nil
}

// Wrap wraps a bare *sql.DB (from the database/sql stdlib package) in a
// *db.Session instance.  It is meant to be used in cases where you do not
// control the instantiation of the database connection, but would still like to
// leverage the facilities provided in Session.
func Wrap(base *sql.DB, dialect string) *Session {
	return &Session{DB: sqlx.NewDb(base, dialect)}
}

// ensure various types conform to Conn interface
var _ Conn = (*sqlx.Tx)(nil)
var _ Conn = (*sqlx.DB)(nil)
