// Package operationfeestats provides useful utilities concerning operation fee
// stats within diamnet,specifically as a central location to store a cached snapshot
// of the state of network per operation fees and surge pricing. This package is
// intended to be at the lowest levels of aurora's dependency tree, please keep
// it free of dependencies to other aurora packages.
package operationfeestats

import (
	"sync"
)

// State represents a snapshot of aurora's view of the state of operation fee's
// on the network.
type State struct {
	FeeMin      int64
	FeeMode     int64
	FeeP10      int64
	FeeP20      int64
	FeeP30      int64
	FeeP40      int64
	FeeP50      int64
	FeeP60      int64
	FeeP70      int64
	FeeP80      int64
	FeeP90      int64
	FeeP95      int64
	FeeP99      int64
	LastBaseFee int64
	LastLedger  int64

	LedgerCapacityUsage string
}

// CurrentState returns the cached snapshot of operation fee state
func CurrentState() State {
	lock.RLock()
	ret := current
	lock.RUnlock()
	return ret
}

// SetState updates the cached snapshot of the operation fee state
func SetState(next State) {
	lock.Lock()
	// in case of one query taking longer than another, this makes
	// sure we don't overwrite the latest fee stats with old stats
	if current.LastLedger < next.LastLedger {
		current = next
	}
	lock.Unlock()
}

// ResetState is used only for testing purposes
func ResetState() {
	current = State{}
}

var current State
var lock sync.RWMutex
