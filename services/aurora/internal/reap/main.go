// Package reap contains the history reaping subsystem for aurora.  This system
// is designed to remove data from the history database such that it does not
// grow indefinitely.  The system can be configured with a number of ledgers to
// maintain at a minimum.
package reap

import (
	"time"

	"github.com/hcnet/go/support/db"
)

// System represents the history reaping subsystem of aurora.
type System struct {
	AuroraDB      *db.Session
	RetentionCount uint

	nextRun time.Time
}

// New initializes the reaper, causing it to begin polling the hcnet-core
// database for now ledgers and ingesting data into the aurora database.
func New(retention uint, aurora *db.Session) *System {
	r := &System{
		AuroraDB:      aurora,
		RetentionCount: retention,
	}

	r.nextRun = time.Now().Add(1 * time.Hour)
	return r
}
