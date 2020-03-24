package dbtest

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	db := Postgres(t)
	t.Log("tempdb url", db.DSN)

	err := exec.Command("psql", db.DSN, "-c", "SELECT 1").Run()
	require.NoError(t, err)

	db.Close()
	err = exec.Command("psql", db.DSN, "-c", "SELECT 1").Run()
	require.Error(t, err)

	// ensure Close() can be called multiple times
	db.Close()
}
