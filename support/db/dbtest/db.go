package dbtest

import (
	"github.com/jmoiron/sqlx"
	"github.com/hcnet/go/support/db/sqlutils"
	"github.com/stretchr/testify/require"
)

// Close closes and deletes the database represented by `db`
func (db *DB) Close() {
	if db.closed {
		return
	}

	db.closer()
	db.closed = true
}

// Load executes all of the statements in the provided sql script against the
// test database, panicking if any fail.  The receiver is returned allowing for
// chain-style calling within your test functions.
func (db *DB) Load(sql string) *DB {
	conn := db.Open()
	defer conn.Close()

	tx, err := conn.Begin()
	require.NoError(db.t, err)

	defer tx.Rollback()

	for i, cmd := range sqlutils.AllStatements(sql) {
		_, err = tx.Exec(cmd)
		require.NoError(db.t, err, "failed execing statement: %d", i)
	}

	err = tx.Commit()
	require.NoError(db.t, err)

	return db
}

// Open opens a sqlx connection to the db.
func (db *DB) Open() *sqlx.DB {
	conn, err := sqlx.Open(db.Dialect, db.DSN)
	require.NoError(db.t, err)

	return conn
}
