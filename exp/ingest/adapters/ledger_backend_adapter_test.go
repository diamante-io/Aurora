package adapters

import (
	"testing"

	"github.com/diamnet/go/exp/ingest/ledgerbackend"
	"github.com/stretchr/testify/assert"
)

func TestGetLatestLedgerSequenceHappyPath(t *testing.T) {
	mockBackend := new(ledgerbackend.MockDatabaseBackend)
	mockBackend.On("GetLatestLedgerSequence").Return(uint32(1), nil)

	lba := LedgerBackendAdapter{Backend: mockBackend}
	seq, err := lba.GetLatestLedgerSequence()

	if assert.NoError(t, err) {
		assert.Equal(t, uint32(1), seq, "latest seqnum returned from backend")
	}
}

func TestGetLatestLedgerSequenceBackendRequired(t *testing.T) {
	lba := LedgerBackendAdapter{}
	_, err := lba.GetLatestLedgerSequence()

	assert.EqualError(t, err, "missing LedgerBackendAdapter.Backend")
}

func TestGetLedgerBackendRequired(t *testing.T) {
	lba := LedgerBackendAdapter{}
	_, err := lba.GetLedger(uint32(1))

	assert.EqualError(t, err, "missing LedgerBackendAdapter.Backend")
}

func TestCloseBackendRequired(t *testing.T) {
	lba := LedgerBackendAdapter{}
	err := lba.Close()

	assert.EqualError(t, err, "missing LedgerBackendAdapter.Backend")
}
