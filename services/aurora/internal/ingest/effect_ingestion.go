package ingest

import (
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/xdr"
)

// Add writes an effect to the database while automatically tracking the index
// to use.
func (ei *EffectIngestion) Add(aid xdr.AccountId, typ history.EffectType, details interface{}) bool {
	if ei.err != nil {
		return false
	}

	ei.added++

	ei.err = ei.Dest.Effect(Address(aid.Address()), ei.OperationID, ei.added, typ, details)
	return ei.err == nil
}

// Finish marks this ingestion as complete, returning any error that was recorded.
func (ei *EffectIngestion) Finish() error {
	err := ei.err
	ei.err = nil
	return err
}
