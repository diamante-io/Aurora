package dbtest

import (
	"fmt"
	"os/exec"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Postgres provisions a new, blank database with a random name on the localhost
// of the running process.  It assumes that you have postgres running on the
// default port, have the command line postgres tools installed, and that the
// current user has access to the server.  It panics on the event of a failure.
func Postgres(t *testing.T) *DB {
	var result DB
	name := randomName()
	result.Dialect = "postgres"
	result.DSN = fmt.Sprintf("postgres://localhost/%s?sslmode=disable", name)
	result.t = t

	// create the db
	err := exec.Command("createdb", name).Run()
	require.NoError(t, err)

	result.closer = func() {
		err := exec.Command("dropdb", name).Run()
		require.NoError(t, err)
	}

	return &result
}
