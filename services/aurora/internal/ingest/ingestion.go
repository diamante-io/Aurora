package ingest

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/db2/sqx"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// ClearAll clears the entire history database
func (ingest *Ingestion) ClearAll() error {
	tables := []string{
		string(AssetStatsTableName),
		string(AccountsTableName),
		string(AssetsTableName),
		string(EffectsTableName),
		string(LedgersTableName),
		string(OperationParticipantsTableName),
		string(OperationsTableName),
		string(TradesTableName),
		string(TransactionParticipantsTableName),
		string(TransactionsTableName),
	}
	return ingest.DB.TruncateTables(tables)
}

// Clear removes a range of data from the history database, exclusive of the end
// id provided.
func (ingest *Ingestion) Clear(start int64, end int64) error {
	clear := ingest.DB.DeleteRange

	err := clear(start, end, "history_effects", "history_operation_id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_effects")
	}
	err = clear(start, end, "history_operation_participants", "history_operation_id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_operation_participants")
	}
	err = clear(start, end, "history_operations", "id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_operations")
	}
	err = clear(start, end, "history_transaction_participants", "history_transaction_id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_transaction_participants")
	}
	err = clear(start, end, "history_transactions", "id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_transactions")
	}
	err = clear(start, end, "history_ledgers", "id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_ledgers")
	}
	err = clear(start, end, "history_trades", "history_operation_id")
	if err != nil {
		return errors.Wrap(err, "Error clearing history_trades")
	}

	return nil
}

// Close finishes the current transaction and finishes this ingestion.
func (ingest *Ingestion) Close() error {
	return ingest.commit()
}

// Effect adds a new row into the `history_effects` table.
func (ingest *Ingestion) Effect(address Address, opid int64, order int, typ history.EffectType, details interface{}) error {
	djson, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "Error marshaling details")
	}

	ingest.builders[EffectsTableName].Values(address, opid, order, typ, djson)
	return nil
}

// Flush writes the currently buffered rows to the db, and if successful
// starts a new transaction.
func (ingest *Ingestion) Flush() error {
	tables := []TableName{
		EffectsTableName,
		LedgersTableName,
		OperationParticipantsTableName,
		OperationsTableName,
		TradesTableName,
		TransactionParticipantsTableName,
		TransactionsTableName,
	}
	// Update IDs for accounts
	err := ingest.UpdateAccountIDs(tables)
	if err != nil {
		return errors.Wrap(err, "Error while updating account ids")
	}

	for _, tableName := range tables {
		err = ingest.builders[tableName].Exec(ingest.DB)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error adding values while inserting to %s", tableName))
		}
	}

	err = ingest.commit()
	if err != nil {
		return errors.Wrap(err, "ingest.commit error")
	}

	return ingest.Start()
}

// UpdateAccountIDs updates IDs of the accounts before inserting
// objects into DB.
func (ingest *Ingestion) UpdateAccountIDs(tables []TableName) error {
	// address => ID in DB
	accounts := map[Address]int64{}
	addresses := []string{}

	// Collect addresses to fetch
	for _, tableName := range tables {
		for _, address := range ingest.builders[tableName].GetAddresses() {
			if _, exists := accounts[address]; !exists {
				addresses = append(addresses, string(address))
			}
			accounts[address] = 0
		}
	}

	if len(addresses) == 0 {
		return nil
	}

	// Get IDs and update map
	q := history.Q{Session: ingest.DB}
	dbAccounts := make([]history.Account, 0, len(addresses))
	err := q.AccountsByAddresses(&dbAccounts, addresses)
	if err != nil {
		return errors.Wrap(err, "q.AccountsByAddresses error")
	}

	for _, row := range dbAccounts {
		accounts[Address(row.Address)] = row.ID
	}

	// Insert non-existent addresses and update map
	addresses = []string{}
	for address, id := range accounts {
		if id == 0 {
			addresses = append(addresses, string(address))
		}
	}

	if len(addresses) > 0 {
		// TODO we should probably batch this too
		dbAccounts = make([]history.Account, 0, len(addresses))
		err = q.CreateAccounts(&dbAccounts, addresses)
		if err != nil {
			return errors.Wrap(err, "q.CreateAccounts error")
		}

		for _, row := range dbAccounts {
			accounts[Address(row.Address)] = row.ID
		}
	}

	// Update IDs in objects
	for _, tableName := range tables {
		ingest.builders[tableName].ReplaceAddressesWithIDs(accounts)
	}

	return nil
}

// Ledger adds a ledger to the current ingestion
func (ingest *Ingestion) Ledger(
	id int64,
	header *core.LedgerHeader,
	successTxsCount int,
	failedTxsCount int,
	ops int,
) {
	ingest.builders[LedgersTableName].Values(
		CurrentVersion,
		id,
		header.Sequence,
		header.LedgerHash,
		null.NewString(header.PrevHash, header.Sequence > 1),
		header.Data.TotalCoins,
		header.Data.FeePool,
		header.Data.BaseFee,
		header.Data.BaseReserve,
		header.Data.MaxTxSetSize,
		time.Unix(header.CloseTime, 0).UTC(),
		time.Now().UTC(),
		time.Now().UTC(),
		successTxsCount, // `transaction_count`
		successTxsCount, // `successful_transaction_count`
		failedTxsCount,
		ops,
		header.Data.LedgerVersion,
		header.DataXDR(),
	)
}

// Operation ingests the provided operation data into a new row in the
// `history_operations` table
func (ingest *Ingestion) Operation(
	id int64,
	txid int64,
	order int32,
	source xdr.AccountId,
	typ xdr.OperationType,
	details map[string]interface{},

) error {
	djson, err := json.Marshal(details)
	if err != nil {
		return errors.Wrap(err, "Error marshaling details")
	}

	ingest.builders[OperationsTableName].Values(id, txid, order, source.Address(), typ, djson)
	return nil
}

// OperationParticipants ingests the provided accounts `aids` as participants of
// operation with id `op`, creating a new row in the
// `history_operation_participants` table.
func (ingest *Ingestion) OperationParticipants(op int64, aids []xdr.AccountId) {
	for _, aid := range aids {
		ingest.builders[OperationParticipantsTableName].Values(op, Address(aid.Address()))
	}
}

// Rollback aborts this ingestions transaction
func (ingest *Ingestion) Rollback() (err error) {
	err = ingest.DB.Rollback()
	return
}

// Start makes the ingestion reeady, initializing the insert builders and tx
func (ingest *Ingestion) Start() (err error) {
	err = ingest.DB.Begin()
	if err != nil {
		return
	}

	ingest.createInsertBuilders()

	return
}

// Trade records a trade into the history_trades table
func (ingest *Ingestion) Trade(
	opid int64,
	order int32,
	buyer xdr.AccountId,
	trade xdr.ClaimOfferAtom,
	ledgerClosedAt int64,
) error {

	q := history.Q{Session: ingest.DB}

	sellerAccountId, err := q.GetCreateAccountID(trade.SellerId)
	if err != nil {
		return errors.Wrap(err, "failed to load seller account id")
	}

	buyerAccountId, err := q.GetCreateAccountID(buyer)
	if err != nil {
		return errors.Wrap(err, "failed to load buyer account id")
	}
	soldAssetId, err := q.GetCreateAssetID(trade.AssetSold)
	if err != nil {
		return errors.Wrap(err, "failed to get sold asset id")
	}

	boughtAssetId, err := q.GetCreateAssetID(trade.AssetBought)
	if err != nil {
		return errors.Wrap(err, "failed to get bought asset id")
	}
	var baseAssetId, counterAssetId int64
	var baseAccountId, counterAccountId int64
	var baseAmount, counterAmount xdr.Int64

	//map seller and buyer to base and counter based on ordering of ids
	if soldAssetId < boughtAssetId {
		baseAccountId, baseAssetId, baseAmount, counterAccountId, counterAssetId, counterAmount =
			sellerAccountId, soldAssetId, trade.AmountSold, buyerAccountId, boughtAssetId, trade.AmountBought
	} else {
		baseAccountId, baseAssetId, baseAmount, counterAccountId, counterAssetId, counterAmount =
			buyerAccountId, boughtAssetId, trade.AmountBought, sellerAccountId, soldAssetId, trade.AmountSold
	}

	ingest.builders[TradesTableName].Values(
		opid,
		order,
		time.Unix(ledgerClosedAt, 0).UTC(),
		trade.OfferId,
		baseAccountId,
		baseAssetId,
		baseAmount,
		counterAccountId,
		counterAssetId,
		counterAmount,
		soldAssetId < boughtAssetId,
	)
	return nil
}

// Transaction ingests the provided transaction data into a new row in the
// `history_transactions` table
func (ingest *Ingestion) Transaction(
	successful bool,
	id int64,
	tx *core.Transaction,
	fee *core.TransactionFee,
) error {
	// Enquote empty signatures
	signatures := tx.Base64Signatures()

	return ingest.builders[TransactionsTableName].Values(
		id,
		tx.TransactionHash,
		tx.LedgerSequence,
		tx.Index,
		tx.SourceAddress(),
		tx.Sequence(),
		tx.MaxFee(),
		tx.FeeCharged(),
		len(tx.Envelope.Tx.Operations),
		tx.EnvelopeXDR(),
		tx.ResultXDR(),
		tx.ResultMetaXDR(),
		fee.ChangesXDR(),
		sqx.StringArray(signatures),
		ingest.formatTimeBounds(tx.Envelope.Tx.TimeBounds),
		tx.MemoType(),
		tx.Memo(),
		time.Now().UTC(),
		time.Now().UTC(),
		successful,
	)
}

// TransactionParticipants ingests the provided account ids as participants of
// transaction with id `tx`, creating a new row in the
// `history_transaction_participants` table.
func (ingest *Ingestion) TransactionParticipants(tx int64, aids []xdr.AccountId) {
	for _, aid := range aids {
		ingest.builders[TransactionParticipantsTableName].Values(tx, Address(aid.Address()))
	}
}

func (ingest *Ingestion) createInsertBuilders() {
	ingest.builders = make(map[TableName]*BatchInsertBuilder)

	ingest.builders[LedgersTableName] = &BatchInsertBuilder{
		TableName: LedgersTableName,
		Columns: []string{
			"importer_version",
			"id",
			"sequence",
			"ledger_hash",
			"previous_ledger_hash",
			"total_coins",
			"fee_pool",
			"base_fee",
			"base_reserve",
			"max_tx_set_size",
			"closed_at",
			"created_at",
			"updated_at",
			"transaction_count",
			"successful_transaction_count",
			"failed_transaction_count",
			"operation_count",
			"protocol_version",
			"ledger_header",
		},
	}

	ingest.builders[TransactionsTableName] = &BatchInsertBuilder{
		TableName: TransactionsTableName,
		Columns: []string{
			"id",
			"transaction_hash",
			"ledger_sequence",
			"application_order",
			"account",
			"account_sequence",
			"max_fee",
			"fee_charged",
			"operation_count",
			"tx_envelope",
			"tx_result",
			"tx_meta",
			"tx_fee_meta",
			"signatures",
			"time_bounds",
			"memo_type",
			"memo",
			"created_at",
			"updated_at",
			"successful",
		},
	}

	ingest.builders[TransactionParticipantsTableName] = &BatchInsertBuilder{
		TableName: TransactionParticipantsTableName,
		Columns: []string{
			"history_transaction_id",
			"history_account_id",
		},
	}

	ingest.builders[OperationsTableName] = &BatchInsertBuilder{
		TableName: OperationsTableName,
		Columns: []string{
			"id",
			"transaction_id",
			"application_order",
			"source_account",
			"type",
			"details",
		},
	}

	ingest.builders[OperationParticipantsTableName] = &BatchInsertBuilder{
		TableName: OperationParticipantsTableName,
		Columns: []string{
			"history_operation_id",
			"history_account_id",
		},
	}

	ingest.builders[EffectsTableName] = &BatchInsertBuilder{
		TableName: EffectsTableName,
		Columns: []string{
			"history_account_id",
			"history_operation_id",
			"\"order\"",
			"type",
			"details",
		},
	}

	ingest.builders[TradesTableName] = &BatchInsertBuilder{
		TableName: TradesTableName,
		Columns: []string{
			"history_operation_id",
			"\"order\"",
			"ledger_closed_at",
			"offer_id",
			"base_account_id",
			"base_asset_id",
			"base_amount",
			"counter_account_id",
			"counter_asset_id",
			"counter_amount",
			"base_is_seller",
		},
	}
}

func (ingest *Ingestion) commit() error {
	err := ingest.DB.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ingest *Ingestion) formatTimeBounds(bounds *xdr.TimeBounds) interface{} {
	if bounds == nil {
		return nil
	}

	if bounds.MaxTime == 0 {
		return sq.Expr("int8range(?,?)", bounds.MinTime, nil)
	}

	maxTime := bounds.MaxTime
	if maxTime > math.MaxInt64 {
		maxTime = math.MaxInt64
	}

	return sq.Expr("int8range(?,?)", bounds.MinTime, maxTime)
}
