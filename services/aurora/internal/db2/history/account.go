package history

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/xdr"
)

// Accounts provides a helper to filter rows from the `history_accounts` table
// with pre-defined filters.  See `AccountsQ` methods for the available filters.
func (q *Q) Accounts() *AccountsQ {
	return &AccountsQ{
		parent: q,
		sql:    selectAccount,
	}
}

// AccountByAddress loads a row from `history_accounts`, by address
func (q *Q) AccountByAddress(dest interface{}, addy string) error {
	sql := selectAccount.Limit(1).Where("ha.address = ?", addy)
	return q.Get(dest, sql)
}

// AccountByID loads a row from `history_accounts`, by id
func (q *Q) AccountByID(dest interface{}, id int64) error {
	sql := selectAccount.Limit(1).Where("ha.id = ?", id)
	return q.Get(dest, sql)
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *AccountsQ) Page(page db2.PageQuery) *AccountsQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "ha.id")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *AccountsQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

// AccountsByAddresses loads a rows from `history_accounts`, by addresses
func (q *Q) AccountsByAddresses(dest interface{}, addresses []string) error {
	sql := selectAccount.Where(map[string]interface{}{
		"ha.address": addresses, // ha.address IN (...)
	})
	return q.Select(dest, sql)
}

// CreateAccounts creates rows for addresses in history_accounts table and
// put. `ON CONFLICT` is required when running a distributed ingestion.
func (q *Q) CreateAccounts(dest interface{}, addresses []string) error {
	sql := sq.Insert("history_accounts").Columns("address")
	for _, address := range addresses {
		sql = sql.Values(address)
	}
	sql = sql.Suffix("ON CONFLICT (address) DO UPDATE SET address=EXCLUDED.address RETURNING *")

	return q.Select(dest, sql)
}

// Return id for account. If account doesn't exist, it will be created and the new id returned.
// `ON CONFLICT` is required when running a distributed ingestion.
func (q *Q) GetCreateAccountID(
	aid xdr.AccountId,
) (result int64, err error) {

	var existing Account

	err = q.AccountByAddress(&existing, aid.Address())

	//account already exists, return id
	if err == nil {
		result = existing.ID
		return
	}

	// unexpected error
	if !q.NoRows(err) {
		return
	}

	//insert account and return id
	err = q.GetRaw(
		&result,
		`INSERT INTO history_accounts (address) VALUES (?) ON CONFLICT (address) DO UPDATE SET address=EXCLUDED.address RETURNING id`,
		aid.Address(),
	)

	return
}

var selectAccount = sq.Select("ha.*").From("history_accounts ha")
