package history

import (
	"database/sql"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/support/errors"
)

const (
	ingestVersion = "exp_ingest_version"
	// Distributed ingestion in Aurora relies on this key and it is part
	// of migration files. If you need to update the key name remember
	// to upgrade it in migration files too!
	lastLedgerKey = "exp_ingest_last_ledger"
)

// GetLastLedgerExpIngestNonBlocking works like GetLastLedgerExpIngest but
// it does not block the value and does not return error if the value
// has not been previously set.
// This is used in status reporting (ex. in root resource of Aurora).
func (q *Q) GetLastLedgerExpIngestNonBlocking() (uint32, error) {
	lastIngestedLedger, err := q.getValueFromStore(lastLedgerKey, false)
	if err != nil {
		return 0, err
	}

	if lastIngestedLedger == "" {
		return 0, nil
	} else {
		ledgerSequence, err := strconv.ParseUint(lastIngestedLedger, 10, 32)
		if err != nil {
			return 0, errors.Wrap(err, "Error converting lastIngestedLedger value")
		}

		return uint32(ledgerSequence), nil
	}
}

// GetLastLedgerExpIngest returns the last ledger ingested by expingest system
// in Aurora. Returns ErrKeyNotFound error if no value has been previously set.
// This is using `SELECT ... FOR UPDATE` what means it's blocking the row for all other
// transactions.This behaviour is critical in distributed ingestion so do not change
// it unless you know what you are doing.
// The value can be set using UpdateLastLedgerExpIngest.
func (q *Q) GetLastLedgerExpIngest() (uint32, error) {
	lastIngestedLedger, err := q.getValueFromStore(lastLedgerKey, true)
	if err != nil {
		return 0, err
	}

	if lastIngestedLedger == "" {
		// This key should always be in a DB (is added in migrations). Otherwise
		// locking won't work.
		return 0, errors.Errorf("`%s` key cannot be found in the key value store", ingestVersion)
	} else {
		ledgerSequence, err := strconv.ParseUint(lastIngestedLedger, 10, 32)
		if err != nil {
			return 0, errors.Wrap(err, "Error converting lastIngestedLedger value")
		}

		return uint32(ledgerSequence), nil
	}
}

// UpdateLastLedgerExpIngest upsets the last ledger ingested by expingest system.
// Can be read using GetLastLedgerExpIngest.
func (q *Q) UpdateLastLedgerExpIngest(ledgerSequence uint32) error {
	return q.updateValueInStore(
		lastLedgerKey,
		strconv.FormatUint(uint64(ledgerSequence), 10),
	)
}

// GetExpIngestVersion returns the exp ingest version. Returns zero
// if there is no value.
func (q *Q) GetExpIngestVersion() (int, error) {
	expVersion, err := q.getValueFromStore(ingestVersion, false)
	if err != nil {
		return 0, err
	}

	if expVersion == "" {
		return 0, nil
	} else {
		version, err := strconv.ParseInt(expVersion, 10, 32)
		if err != nil {
			return 0, errors.Wrap(err, "Error converting expVersion value")
		}

		return int(version), nil
	}
}

// UpdateExpIngestVersion upsets the exp ingest version.
func (q *Q) UpdateExpIngestVersion(ledgerSequence int) error {
	return q.updateValueInStore(
		ingestVersion,
		strconv.FormatUint(uint64(ledgerSequence), 10),
	)
}

// getValueFromStore returns a value for a given key from KV store. If value
// is not present in the key value store "" will be returned.
func (q *Q) getValueFromStore(key string, forUpdate bool) (string, error) {
	query := sq.Select("key_value_store.value").
		From("key_value_store").
		Where("key_value_store.key = ?", key)

	if forUpdate {
		query = query.Suffix("FOR UPDATE")
	}

	var value string
	if err := q.Get(&value, query); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return "", nil
		}
		return "", errors.Wrap(err, "could not get value")
	}

	return value, nil
}

// updateValueInStore updates a value for a given key in KV store
func (q *Q) updateValueInStore(key, value string) error {
	query := sq.Insert("key_value_store").
		Columns("key", "value").
		Values(key, value).
		Suffix("ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value")

	_, err := q.Exec(query)
	return err
}
