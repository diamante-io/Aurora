package core

import (
	"encoding/base64"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// IsAuthRequired returns true if the account has the "AUTH_REQUIRED" option
// turned on.
func (ac Account) IsAuthRequired() bool {
	return (ac.Flags & xdr.AccountFlagsAuthRequiredFlag) != 0
}

// IsAuthRevocable returns true if the account has the "AUTH_REVOCABLE" option
// turned on.
func (ac Account) IsAuthRevocable() bool {
	return (ac.Flags & xdr.AccountFlagsAuthRevocableFlag) != 0
}

// IsAuthImmutable returns true if the account has the "AUTH_IMMUTABLE" option
// turned on.
func (ac Account) IsAuthImmutable() bool {
	return (ac.Flags & xdr.AccountFlagsAuthImmutableFlag) != 0
}

// AccountByAddress loads a row from `accounts`, by address
func (q *Q) AccountByAddress(dest *Account, addy string) error {
	sql := selectAccount.Limit(1).Where("accountid = ?", addy)
	err := q.Get(dest, sql)
	if err != nil {
		return err
	}

	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	if schemaVersion >= 9 {
		// Since schema version 9, home_domain is base64 encoded.
		decoded, err := base64.StdEncoding.DecodeString(dest.HomeDomain.String)
		if err != nil {
			return errors.Wrap(err, "Unable to base64 decode HomeDomain")
		}
		dest.HomeDomain.String = string(decoded)
	}

	return nil
}

// SequencesForAddresses loads the current sequence number for every accountid
// specified in `addys`
func (q *Q) SequencesForAddresses(dest interface{}, addys []string) error {
	sql := sq.
		Select("seqnum as sequence", "accountid as address").
		From("accounts").
		Where(sq.Eq{"accountid": addys})

	return q.Select(dest, sql)
}

// SequenceProvider returns a new sequence provider.
func (q *Q) SequenceProvider() *SequenceProvider {
	return &SequenceProvider{Q: q}
}

// Get implements `txsub.SequenceProvider`
func (sp *SequenceProvider) Get(addys []string) (map[string]uint64, error) {
	rows := []struct {
		Address  string
		Sequence uint64
	}{}

	err := sp.Q.SequencesForAddresses(&rows, addys)
	if err != nil {
		return nil, err
	}

	results := make(map[string]uint64)
	for _, r := range rows {
		results[r.Address] = r.Sequence
	}
	return results, nil
}

var selectAccount = sq.Select(
	"a.accountid",
	"a.balance",
	"a.seqnum",
	"a.numsubentries",
	"a.inflationdest",
	"a.homedomain",
	"a.thresholds",
	"a.flags",
	"a.lastmodified",
	// Liabilities can be NULL so can error without `coalesce`:
	// `Invalid value for xdr.Int64`
	"coalesce(a.buyingliabilities, 0) as buyingliabilities",
	"coalesce(a.sellingliabilities, 0) as sellingliabilities",
).From("accounts a")
