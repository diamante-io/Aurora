package dbtest

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestMysql(t *testing.T) {
	db := Mysql(t)
	t.Log("tempdb url", db.DSN)

	conn, err := sqlx.Open("mysql", db.DSN)
	require.NoError(t, err)

	_, err = conn.Exec("CREATE TABLE t1 (c1 INT PRIMARY KEY) ;")
	require.NoError(t, err)

	db.Close()
	_, err = conn.Exec("CREATE TABLE t2 (c1 INT PRIMARY KEY) ;")
	require.Error(t, err)

	// ensure Close() can be called multiple times
	db.Close()
}
