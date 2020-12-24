package adapters

import (
	"fmt"

	"github.com/diamnet/go/exp/ingest/io"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/historyarchive"
	"github.com/diamnet/go/xdr"
)

const msrBufferSize = 50000

// HistoryArchiveAdapter is an adapter for the historyarchive package to read from history archives
type HistoryArchiveAdapter struct {
	archive historyarchive.ArchiveInterface
}

// MakeHistoryArchiveAdapter is a factory method to make a HistoryArchiveAdapter
func MakeHistoryArchiveAdapter(archive historyarchive.ArchiveInterface) *HistoryArchiveAdapter {
	return &HistoryArchiveAdapter{archive: archive}
}

// GetLatestLedgerSequence returns the latest ledger sequence or an error
func (haa *HistoryArchiveAdapter) GetLatestLedgerSequence() (uint32, error) {
	has, e := haa.archive.GetRootHAS()
	if e != nil {
		return 0, fmt.Errorf("could not get root HAS: %s", e)
	}

	return has.CurrentLedger, nil
}

// BucketListHash returns the bucket list hash to compare with hash in the
// ledger header fetched from DiamNet-Core.
func (haa *HistoryArchiveAdapter) BucketListHash(sequence uint32) (xdr.Hash, error) {
	exists, err := haa.archive.CategoryCheckpointExists("history", sequence)
	if err != nil {
		return xdr.Hash{}, errors.Wrap(err, "error checking if category checkpoint exists")
	}
	if !exists {
		return xdr.Hash{}, fmt.Errorf("history checkpoint does not exist for ledger %d", sequence)
	}

	has, err := haa.archive.GetCheckpointHAS(sequence)
	if err != nil {
		return xdr.Hash{}, fmt.Errorf("unable to get checkpoint HAS at ledger sequence %d: %s", sequence, err)
	}

	return has.BucketListHash()
}

// GetState returns a reader with the state of the ledger at the provided sequence number
func (haa *HistoryArchiveAdapter) GetState(sequence uint32, tempSet io.TempSet) (io.StateReader, error) {
	exists, err := haa.archive.CategoryCheckpointExists("history", sequence)
	if err != nil {
		return nil, errors.Wrap(err, "error checking if category checkpoint exists")
	}
	if !exists {
		return nil, fmt.Errorf("history checkpoint does not exist for ledger %d", sequence)
	}

	sr, e := io.MakeSingleLedgerStateReader(haa.archive, tempSet, sequence, msrBufferSize)
	if e != nil {
		return nil, errors.Wrap(e, "could not make memory state reader")
	}

	return sr, nil
}

// GetLedger gets a ledger transaction result at the provided sequence number
func (haa *HistoryArchiveAdapter) GetLedger(sequence uint32) (io.ArchiveLedgerReader, error) {
	return nil, fmt.Errorf("not implemented yet")
}
