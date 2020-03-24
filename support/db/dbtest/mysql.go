package dbtest

import (
	"fmt"
	"os/exec"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

// Mysql provisions a new, blank database with a random name on the localhost of
// the running process.  It assumes that you have mysql running and that the
// root user has access with no password.  It panics on
// the event of a failure.
func Mysql(t *testing.T) *DB {
	var result DB
	name := randomName()
	result.Dialect = "mysql"
	result.DSN = fmt.Sprintf("root@/%s", name)
	result.t = t
	// create the db
	err := exec.Command("mysql", "--user=root", "-e", fmt.Sprintf("CREATE DATABASE %s;", name)).Run()
	require.NoError(t, err)

	result.closer = func() {
		err := exec.Command("mysql", "--user=root", "-e", fmt.Sprintf("DROP DATABASE %s;", name)).Run()
		require.NoError(t, err)
	}

	return &result
}
