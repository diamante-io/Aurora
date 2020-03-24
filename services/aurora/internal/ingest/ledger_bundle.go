package ingest

import (
	"database/sql"
	"fmt"

	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/errors"
)

// Load runs queries against `core` to fill in the records of the bundle.
func (lb *LedgerBundle) Load(db *db.Session) error {
	q := &core.Q{Session: db}
	// Load Header
	err := q.LedgerHeaderBySequence(&lb.Header, lb.Sequence)
	if err != nil {
		// Remove when Aurora is able to handle gaps in hcnet-core DB.
		// More info:
		// * https://github.com/hcnet/go/issues/335
		// * https://www.hcnet.org/developers/software/known-issues.html#gaps-detected
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("Gap detected in hcnet-core database (ledger=%d). More information: https://www.hcnet.org/developers/software/known-issues.html#gaps-detected", lb.Sequence))
		}
		return errors.Wrap(err, "failed to load header")
	}

	// Load transactions
	err = q.TransactionsByLedger(&lb.Transactions, lb.Sequence)
	if err != nil {
		return errors.Wrap(err, "failed to load transactions")
	}

	err = q.TransactionFeesByLedger(&lb.TransactionFees, lb.Sequence)
	if err != nil {
		return errors.Wrap(err, "failed to load transaction fees")
	}

	return nil
}
