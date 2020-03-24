// Package history contains database record definitions useable for
// reading rows from a the history portion of aurora's database
package history

import (
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null"

	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/xdr"
)

const (
	// account effects

	// EffectAccountCreated effects occur when a new account is created
	EffectAccountCreated EffectType = 0 // from create_account

	// EffectAccountRemoved effects occur when one account is merged into another
	EffectAccountRemoved EffectType = 1 // from merge_account

	// EffectAccountCredited effects occur when an account receives some currency
	EffectAccountCredited EffectType = 2 // from create_account, payment, path_payment, merge_account

	// EffectAccountDebited effects occur when an account sends some currency
	EffectAccountDebited EffectType = 3 // from create_account, payment, path_payment, create_account

	// EffectAccountThresholdsUpdated effects occur when an account changes its
	// multisig thresholds.
	EffectAccountThresholdsUpdated EffectType = 4 // from set_options

	// EffectAccountHomeDomainUpdated effects occur when an account changes its
	// home domain.
	EffectAccountHomeDomainUpdated EffectType = 5 // from set_options

	// EffectAccountFlagsUpdated effects occur when an account changes its
	// account flags, either clearing or setting.
	EffectAccountFlagsUpdated EffectType = 6 // from set_options

	// EffectAccountInflationDestinationUpdated effects occur when an account changes its
	// inflation destination.
	EffectAccountInflationDestinationUpdated EffectType = 7 // from set_options

	// signer effects

	// EffectSignerCreated occurs when an account gains a signer
	EffectSignerCreated EffectType = 10 // from set_options

	// EffectSignerRemoved occurs when an account loses a signer
	EffectSignerRemoved EffectType = 11 // from set_options

	// EffectSignerUpdated occurs when an account changes the weight of one of its
	// signers.
	EffectSignerUpdated EffectType = 12 // from set_options

	// trustline effects

	// EffectTrustlineCreated occurs when an account trusts an anchor
	EffectTrustlineCreated EffectType = 20 // from change_trust

	// EffectTrustlineRemoved occurs when an account removes struct by setting the
	// limit of a trustline to 0
	EffectTrustlineRemoved EffectType = 21 // from change_trust

	// EffectTrustlineUpdated occurs when an account changes a trustline's limit
	EffectTrustlineUpdated EffectType = 22 // from change_trust, allow_trust

	// EffectTrustlineAuthorized occurs when an anchor has AUTH_REQUIRED flag set
	// to true and it authorizes another account's trustline
	EffectTrustlineAuthorized EffectType = 23 // from allow_trust

	// EffectTrustlineDeauthorized occurs when an anchor revokes access to a asset
	// it issues.
	EffectTrustlineDeauthorized EffectType = 24 // from allow_trust

	// trading effects

	// EffectOfferCreated occurs when an account offers to trade an asset
	EffectOfferCreated EffectType = 30 // from manage_offer, creat_passive_offer

	// EffectOfferRemoved occurs when an account removes an offer
	EffectOfferRemoved EffectType = 31 // from manage_offer, creat_passive_offer, path_payment

	// EffectOfferUpdated occurs when an offer is updated by the offering account.
	EffectOfferUpdated EffectType = 32 // from manage_offer, creat_passive_offer, path_payment

	// EffectTrade occurs when a trade is initiated because of a path payment or
	// offer operation.
	EffectTrade EffectType = 33 // from manage_offer, creat_passive_offer, path_payment

	// data effects

	// EffectDataCreated occurs when an account gets a new data field
	EffectDataCreated EffectType = 40 // from manage_data

	// EffectDataRemoved occurs when an account removes a data field
	EffectDataRemoved EffectType = 41 // from manage_data

	// EffectDataUpdated occurs when an account changes a data field's value
	EffectDataUpdated EffectType = 42 // from manage_data

	// EffectSequenceBumped occurs when an account bumps their sequence number
	EffectSequenceBumped EffectType = 43 // from bump_sequence

)

// ExperimentalIngestionTables is a list of tables populated by the experimental
// ingestion system
var ExperimentalIngestionTables = []string{
	"accounts_signers",
	"offers",
}

// Account is a row of data from the `history_accounts` table
type Account struct {
	ID      int64
	Address string `db:"address"`
}

// AccountsQ is a helper struct to aid in configuring queries that loads
// slices of account structs.
type AccountsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// AccountSigner is a row of data from the `accounts_signers` table
type AccountSigner struct {
	Account string `db:"account"`
	Signer  string `db:"signer"`
	Weight  int32  `db:"weight"`
}

// Asset is a row of data from the `history_assets` table
type Asset struct {
	ID     int64  `db:"id"`
	Type   string `db:"asset_type"`
	Code   string `db:"asset_code"`
	Issuer string `db:"asset_issuer"`
}

// AssetStat is a row in the asset_stats table representing the stats per Asset
type AssetStat struct {
	ID          int64  `db:"id"`
	Amount      string `db:"amount"`
	NumAccounts int32  `db:"num_accounts"`
	Flags       int8   `db:"flags"`
	Toml        string `db:"toml"`
}

// Effect is a row of data from the `history_effects` table
type Effect struct {
	HistoryAccountID   int64       `db:"history_account_id"`
	Account            string      `db:"address"`
	HistoryOperationID int64       `db:"history_operation_id"`
	Order              int32       `db:"order"`
	Type               EffectType  `db:"type"`
	DetailsString      null.String `db:"details"`
}

// SequenceBumped is a struct of data from `effects.DetailsString`
// when the effect type is sequence bumped.
type SequenceBumped struct {
	NewSeq int64 `json:"new_seq"`
}

// EffectsQ is a helper struct to aid in configuring queries that loads
// slices of Ledger structs.
type EffectsQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// EffectType is the numeric type for an effect, used as the `type` field in the
// `history_effects` table.
type EffectType int

// FeeStats is a row of data from the min, mode, percentile aggregate functions over the
// `history_transactions` table.
type FeeStats struct {
	Min  null.Int `db:"min"`
	Mode null.Int `db:"mode"`
	P10  null.Int `db:"p10"`
	P20  null.Int `db:"p20"`
	P30  null.Int `db:"p30"`
	P40  null.Int `db:"p40"`
	P50  null.Int `db:"p50"`
	P60  null.Int `db:"p60"`
	P70  null.Int `db:"p70"`
	P80  null.Int `db:"p80"`
	P90  null.Int `db:"p90"`
	P95  null.Int `db:"p95"`
	P99  null.Int `db:"p99"`
}

// KeyValueStoreRow represents a row in key value store.
type KeyValueStoreRow struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

// LatestLedger represents a response from the raw LatestLedgerBaseFeeAndSequence
// query.
type LatestLedger struct {
	BaseFee  int32 `db:"base_fee"`
	Sequence int32 `db:"sequence"`
}

// Ledger is a row of data from the `history_ledgers` table
type Ledger struct {
	TotalOrderID
	Sequence                   int32       `db:"sequence"`
	ImporterVersion            int32       `db:"importer_version"`
	LedgerHash                 string      `db:"ledger_hash"`
	PreviousLedgerHash         null.String `db:"previous_ledger_hash"`
	TransactionCount           int32       `db:"transaction_count"`
	SuccessfulTransactionCount *int32      `db:"successful_transaction_count"`
	FailedTransactionCount     *int32      `db:"failed_transaction_count"`
	OperationCount             int32       `db:"operation_count"`
	ClosedAt                   time.Time   `db:"closed_at"`
	CreatedAt                  time.Time   `db:"created_at"`
	UpdatedAt                  time.Time   `db:"updated_at"`
	TotalCoins                 int64       `db:"total_coins"`
	FeePool                    int64       `db:"fee_pool"`
	BaseFee                    int32       `db:"base_fee"`
	BaseReserve                int32       `db:"base_reserve"`
	MaxTxSetSize               int32       `db:"max_tx_set_size"`
	ProtocolVersion            int32       `db:"protocol_version"`
	LedgerHeaderXDR            null.String `db:"ledger_header"`
}

// LedgerCapacityUsageStats contains ledgers fullness stats.
type LedgerCapacityUsageStats struct {
	CapacityUsage null.String `db:"ledger_capacity_usage"`
}

// LedgerCache is a helper struct to load ledger data related to a batch of
// sequences.
type LedgerCache struct {
	Records map[int32]Ledger

	lock   sync.Mutex
	queued map[int32]struct{}
}

// LedgersQ is a helper struct to aid in configuring queries that loads
// slices of Ledger structs.
type LedgersQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// Operation is a row of data from the `history_operations` table
type Operation struct {
	TotalOrderID
	TransactionID    int64             `db:"transaction_id"`
	TransactionHash  string            `db:"transaction_hash"`
	TxResult         string            `db:"tx_result"`
	ApplicationOrder int32             `db:"application_order"`
	Type             xdr.OperationType `db:"type"`
	DetailsString    null.String       `db:"details"`
	SourceAccount    string            `db:"source_account"`
	// Check db2/history.Transaction.Successful field comment for more information.
	TransactionSuccessful *bool `db:"transaction_successful"`
}

// Offer is row of data from the `offers` table from hcnet-core
type Offer struct {
	SellerID string    `db:"sellerid"`
	OfferID  xdr.Int64 `db:"offerid"`

	SellingAsset xdr.Asset `db:"sellingasset"`
	BuyingAsset  xdr.Asset `db:"buyingasset"`

	Amount             xdr.Int64 `db:"amount"`
	Pricen             int32     `db:"pricen"`
	Priced             int32     `db:"priced"`
	Price              float64   `db:"price"`
	Flags              uint32    `db:"flags"`
	LastModifiedLedger uint32    `db:"last_modified_ledger"`
}

// OperationsQ is a helper struct to aid in configuring queries that loads
// slices of Operation structs.
type OperationsQ struct {
	Err                 error
	parent              *Q
	sql                 sq.SelectBuilder
	opIdCol             string
	includeFailed       bool
	includeTransactions bool
}

// Q is a helper struct on which to hang common_trades queries against a history
// portion of the aurora database.
type Q struct {
	*db.Session
}

// QSigners defines signer related queries.
type QSigners interface {
	GetLastLedgerExpIngestNonBlocking() (uint32, error)
	GetLastLedgerExpIngest() (uint32, error)
	UpdateLastLedgerExpIngest(ledgerSequence uint32) error
	AccountsForSigner(signer string, page db2.PageQuery) ([]AccountSigner, error)
	CreateAccountSigner(account, signer string, weight int32) error
	RemoveAccountSigner(account, signer string) error
}

// QOffers defines offer related queries.
type QOffers interface {
	GetAllOffers() ([]Offer, error)
	UpsertOffer(offer xdr.OfferEntry, lastModifiedLedger xdr.Uint32) error
	RemoveOffer(offerID xdr.Int64) error
}

// TotalOrderID represents the ID portion of rows that are identified by the
// "TotalOrderID".  See total_order_id.go in the `db` package for details.
type TotalOrderID struct {
	ID int64 `db:"id"`
}

// Trade represents a trade from the trades table, joined with asset information from the assets table and account
// addresses from the accounts table
type Trade struct {
	HistoryOperationID int64     `db:"history_operation_id"`
	Order              int32     `db:"order"`
	LedgerCloseTime    time.Time `db:"ledger_closed_at"`
	OfferID            int64     `db:"offer_id"`
	BaseOfferID        *int64    `db:"base_offer_id"`
	BaseAccount        string    `db:"base_account"`
	BaseAssetType      string    `db:"base_asset_type"`
	BaseAssetCode      string    `db:"base_asset_code"`
	BaseAssetIssuer    string    `db:"base_asset_issuer"`
	BaseAmount         xdr.Int64 `db:"base_amount"`
	CounterOfferID     *int64    `db:"counter_offer_id"`
	CounterAccount     string    `db:"counter_account"`
	CounterAssetType   string    `db:"counter_asset_type"`
	CounterAssetCode   string    `db:"counter_asset_code"`
	CounterAssetIssuer string    `db:"counter_asset_issuer"`
	CounterAmount      xdr.Int64 `db:"counter_amount"`
	BaseIsSeller       bool      `db:"base_is_seller"`
	PriceN             null.Int  `db:"price_n"`
	PriceD             null.Int  `db:"price_d"`
}

// TradesQ is a helper struct to aid in configuring queries that loads
// slices of trade structs.
type TradesQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// Transaction is a row of data from the `history_transactions` table
type Transaction struct {
	TotalOrderID
	TransactionHash  string      `db:"transaction_hash"`
	LedgerSequence   int32       `db:"ledger_sequence"`
	LedgerCloseTime  time.Time   `db:"ledger_close_time"`
	ApplicationOrder int32       `db:"application_order"`
	Account          string      `db:"account"`
	AccountSequence  string      `db:"account_sequence"`
	MaxFee           int32       `db:"max_fee"`
	FeeCharged       int32       `db:"fee_charged"`
	OperationCount   int32       `db:"operation_count"`
	TxEnvelope       string      `db:"tx_envelope"`
	TxResult         string      `db:"tx_result"`
	TxMeta           string      `db:"tx_meta"`
	TxFeeMeta        string      `db:"tx_fee_meta"`
	SignatureString  string      `db:"signatures"`
	MemoType         string      `db:"memo_type"`
	Memo             null.String `db:"memo"`
	ValidAfter       null.Int    `db:"valid_after"`
	ValidBefore      null.Int    `db:"valid_before"`
	CreatedAt        time.Time   `db:"created_at"`
	UpdatedAt        time.Time   `db:"updated_at"`
	// NULL indicates successful transaction. We wanted a migration to be fast
	// however Postgres performs a table rewrite if a new column has a default
	// non-null value. We need `NULL` to indicate successful transaction because
	// otherwise all existing transactions would be interpreted as failed until
	// ledger is reingested.
	Successful *bool `db:"successful"`
}

// TransactionsQ is a helper struct to aid in configuring queries that loads
// slices of transaction structs.
type TransactionsQ struct {
	Err           error
	parent        *Q
	sql           sq.SelectBuilder
	includeFailed bool
}

// ElderLedger loads the oldest ledger known to the history database
func (q *Q) ElderLedger(dest interface{}) error {
	return q.GetRaw(dest, `SELECT COALESCE(MIN(sequence), 0) FROM history_ledgers`)
}

// LatestLedger loads the latest known ledger
func (q *Q) LatestLedger(dest interface{}) error {
	return q.GetRaw(dest, `SELECT COALESCE(MAX(sequence), 0) FROM history_ledgers`)
}

// LatestLedgerBaseFeeAndSequence loads the latest known ledger's base fee and
// sequence number.
func (q *Q) LatestLedgerBaseFeeAndSequence(dest interface{}) error {
	return q.GetRaw(dest, `
		SELECT base_fee, sequence
		FROM history_ledgers
		WHERE sequence = (SELECT COALESCE(MAX(sequence), 0) FROM history_ledgers)
	`)
}

// OldestOutdatedLedgers populates a slice of ints with the first million
// outdated ledgers, based upon the provided `currentVersion` number
func (q *Q) OldestOutdatedLedgers(dest interface{}, currentVersion int) error {
	return q.SelectRaw(dest, `
		SELECT sequence
		FROM history_ledgers
		WHERE importer_version < $1
		ORDER BY sequence ASC
		LIMIT 1000000`, currentVersion)
}
