package processors

import (
	"github.com/hcnet/go/xdr"
)

// RootProcessor is useful when a pipeline needs to be split stream into
// multiple branches right away. This processor is a no-op - just passes the data
// to all children.
type RootProcessor struct {
	noStateProcessor
}

// CSVPrinter prints ledger entries to a file or stdout (when Filename is empty).
// Can be used both for processing state and ledgers.
// The state output matches the format of data in hcnet-core DB so can be
// used for diff-testing the state readers.
type CSVPrinter struct {
	noStateProcessor

	Filename string
}

// StatusLogger prints number of processed entries every N
// entries.
type StatusLogger struct {
	noStateProcessor

	N int
}

// EntryTypeFilter is a pipeline.StateProcessor that filters out all
// entries that are not of type `Type`.
type EntryTypeFilter struct {
	noStateProcessor

	Type xdr.LedgerEntryType
}

type noStateProcessor struct{}

func (n *noStateProcessor) Reset() {
	// No internal state
}
