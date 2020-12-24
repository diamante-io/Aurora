// Package ledger provides useful utilities concerning ledgers within diamnet,
// specifically as a central location to store a cached snapshot of the state of
// both aurora's and diamnet-core's views of the ledger.  This package is
// intended to be at the lowest levels of aurora's dependency tree, please keep
// it free of dependencies to other aurora packages.
package ledger

import (
	"sync"
)

// State represents a snapshot of both aurora's and diamnet-core's view of the
// ledger.
type State struct {
	CoreLatest       int32  `db:"core_latest"`
	HistoryLatest    int32  `db:"history_latest"`
	HistoryElder     int32  `db:"history_elder"`
	ExpHistoryLatest uint32 `db:"exp_history_latest"`
}

// CurrentState returns the cached snapshot of ledger state
func CurrentState() State {
	lock.RLock()
	ret := current
	lock.RUnlock()
	return ret
}

// SetState updates the cached snapshot of the ledger state
func SetState(next State) {
	lock.Lock()
	current = next
	lock.Unlock()
}

var current State
var lock sync.RWMutex
