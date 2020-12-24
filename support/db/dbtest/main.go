// Package dbtest is a package to ease the pain of developing test code that
// works against external databases.  It provides helper functions to provision
// temporary databases and load them with test data.
package dbtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/diamnet/go/support/errors"
)

// DB represents an ephemeral database that can be starts blank and can be used
// to run tests against.
type DB struct {
	Dialect string
	DSN     string
	t       *testing.T
	closer  func()
	closed  bool
}

// randomName returns a new psuedo-random name that is sufficient for naming a
// test database.  In the event that reading from the source of randomness
// fails, a panic will occur.
func randomName() string {
	raw := make([]byte, 6)

	_, err := rand.Read(raw)
	if err != nil {
		err = errors.Wrap(err, "read from rand failed")
		panic(err)
	}

	enc := hex.EncodeToString(raw)

	return fmt.Sprintf("test_%s", enc)
}
