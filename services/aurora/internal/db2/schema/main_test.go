package schema

import (
	"testing"

	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/db/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tdb := dbtest.Postgres(t)
	defer tdb.Close()
	sess := &db.Session{DB: tdb.Open()}

	defer sess.DB.Close()

	err := Init(sess)

	assert.NoError(t, err)
}
