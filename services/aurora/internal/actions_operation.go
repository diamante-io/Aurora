package aurora

import (
	"fmt"
	"strings"

	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/render/sse"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/services/aurora/internal/toid"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/hal"
	supportProblem "github.com/diamnet/go/support/render/problem"
	"github.com/diamnet/go/xdr"
)

// This file contains the actions:
//
// OperationIndexAction: pages of operations
// OperationShowAction: single operation by id

// Interface verifications
var _ actions.JSONer = (*OperationIndexAction)(nil)
var _ actions.EventStreamer = (*OperationIndexAction)(nil)

const (
	joinTransactions = "transactions"
)

// OperationIndexAction renders a page of operations resources, identified by
// a normal page query and optionally filtered by an account, ledger, or
// transaction.
type OperationIndexAction struct {
	Action
	LedgerFilter        int32
	AccountFilter       string
	TransactionFilter   string
	PagingParams        db2.PageQuery
	OperationRecords    []history.Operation
	TransactionRecords  []history.Transaction
	Ledgers             *history.LedgerCache
	Page                hal.Page
	IncludeFailed       bool
	IncludeTransactions bool
	OnlyPayments        bool
}

// JSON is a method for actions.JSON
func (action *OperationIndexAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
		action.loadRecords,
		action.loadLedgers,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

// SSE is a method for actions.SSE
func (action *OperationIndexAction) SSE(stream *sse.Stream) error {
	action.Setup(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
	)
	action.Do(
		action.loadRecords,
		action.loadLedgers,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			operationRecords := action.OperationRecords[stream.SentCount():]
			var transactionRecords []history.Transaction
			if action.IncludeTransactions {
				transactionRecords = action.TransactionRecords[stream.SentCount():]
			}
			for i, operationRecord := range operationRecords {
				ledger, found := action.Ledgers.Records[operationRecord.LedgerSequence()]
				if !found {
					action.Err = errors.New(fmt.Sprintf("could not find ledger data for sequence %d", operationRecord.LedgerSequence()))
					return
				}

				var transactionRecord *history.Transaction
				if action.IncludeTransactions {
					transactionRecord = &transactionRecords[i]
				}

				res, err := resourceadapter.NewOperation(action.R.Context(), operationRecord, transactionRecord, ledger)
				if err != nil {
					action.Err = err
					return
				}

				stream.Send(sse.Event{
					ID:   res.PagingToken(),
					Data: res,
				})
			}
		},
	)

	return action.Err
}

func parseJoinField(action *actions.Base) (map[string]bool, error) {
	join := action.GetString("join")
	validJoins := map[string]bool{}
	if join != "" {
		for _, part := range strings.Split(join, ",") {
			if part == joinTransactions {
				validJoins[joinTransactions] = true
			} else {
				return nil, supportProblem.MakeInvalidFieldProblem(
					"join",
					fmt.Errorf("it is not possible to join '%s'", part),
				)
			}
		}
	}
	return validJoins, nil
}

func (action *OperationIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.AccountFilter = action.GetAddress("account_id")
	action.LedgerFilter = action.GetInt32("ledger_id")
	action.TransactionFilter = action.GetStringFromURLParam("tx_id")
	action.PagingParams = action.GetPageQuery()
	action.IncludeFailed = action.GetBool("include_failed")
	parsed, err := parseJoinField(&action.Action.Base)
	if err != nil {
		action.Err = err
		return
	}
	action.IncludeTransactions = parsed[joinTransactions]

	filters, err := countNonEmpty(
		action.AccountFilter,
		action.LedgerFilter,
		action.TransactionFilter,
	)

	if err != nil {
		action.Err = errors.Wrap(err, "Error in countNonEmpty")
		return
	}

	if filters > 1 {
		action.Err = supportProblem.BadRequest
		return
	}

	// Double check TransactionFilter as it's used to determine if failed txs should be returned
	if action.TransactionFilter != "" && !isValidTransactionHash(action.TransactionFilter) {
		action.Err = supportProblem.MakeInvalidFieldProblem("tx_id", errors.New("Invalid transaction hash"))
		return
	}

	if action.IncludeFailed && !action.App.config.IngestFailedTransactions {
		err := errors.New("`include_failed` parameter is unavailable when Aurora is not ingesting failed " +
			"transactions. Set `INGEST_FAILED_TRANSACTIONS=true` to start ingesting them.")
		action.Err = supportProblem.MakeInvalidFieldProblem("include_failed", err)
		return
	}
}

func validateTransactionForOperation(transaction history.Transaction, operation history.Operation) error {
	if transaction.ID != operation.TransactionID {
		return errors.Errorf(
			"transaction id %v does not match transaction id in operation %v",
			transaction.ID,
			operation.TransactionID,
		)
	}
	if transaction.TransactionHash != operation.TransactionHash {
		return errors.Errorf(
			"transaction hash %v does not match transaction hash in operation %v",
			transaction.TransactionHash,
			operation.TransactionHash,
		)
	}
	if transaction.TxResult != operation.TxResult {
		return errors.Errorf(
			"transaction result %v does not match transaction result in operation %v",
			transaction.TxResult,
			operation.TxResult,
		)
	}
	if transaction.IsSuccessful() != operation.IsTransactionSuccessful() {
		return errors.Errorf(
			"transaction successful flag %v does not match transaction successful flag in operation %v",
			transaction.IsSuccessful(),
			operation.IsTransactionSuccessful(),
		)
	}

	return nil
}

func (action *OperationIndexAction) loadRecords() {
	q := action.HistoryQ()
	ops := q.Operations()

	switch {
	case action.AccountFilter != "":
		ops.ForAccount(action.AccountFilter)
	case action.LedgerFilter > 0:
		ops.ForLedger(action.LedgerFilter)
	case action.TransactionFilter != "":
		ops.ForTransaction(action.TransactionFilter)
	}

	// When querying operations for transaction return both successful
	// and failed operations. We assume that because user is querying
	// this specific transactions, she knows it's status.
	if action.TransactionFilter != "" || action.IncludeFailed {
		ops.IncludeFailed()
	}

	if action.IncludeTransactions {
		ops.IncludeTransactions()
	}

	if action.OnlyPayments {
		ops.OnlyPayments()
	}

	action.OperationRecords, action.TransactionRecords, action.Err = ops.Page(action.PagingParams).Fetch()
	if action.Err != nil {
		return
	}

	if action.IncludeTransactions && len(action.TransactionRecords) != len(action.OperationRecords) {
		action.Err = errors.New("number of transactions doesn't match number of operations")
		return
	}

	for i, o := range action.OperationRecords {
		if !action.IncludeFailed && action.TransactionFilter == "" {
			if !o.IsTransactionSuccessful() {
				action.Err = errors.Errorf("Corrupted data! `include_failed=false` but returned transaction in /operations is failed: %s", o.TransactionHash)
				return
			}

			var resultXDR xdr.TransactionResult
			action.Err = xdr.SafeUnmarshalBase64(o.TxResult, &resultXDR)
			if action.Err != nil {
				return
			}

			if resultXDR.Result.Code != xdr.TransactionResultCodeTxSuccess {
				action.Err = errors.Errorf("Corrupted data! `include_failed=false` but returned transaction /operations is failed: %s %s", o.TransactionHash, o.TxResult)
				return
			}
		}
		if action.IncludeTransactions {
			transaction := action.TransactionRecords[i]
			action.Err = validateTransactionForOperation(transaction, o)
			if action.Err != nil {
				return
			}
		}
	}
}

// loadLedgers populates the ledger cache for this action
func (action *OperationIndexAction) loadLedgers() {
	action.Ledgers = &history.LedgerCache{}
	for _, op := range action.OperationRecords {
		action.Ledgers.Queue(op.LedgerSequence())
	}
	action.Err = action.Ledgers.Load(action.HistoryQ())
}

func (action *OperationIndexAction) loadPage() {
	for i, operationRecord := range action.OperationRecords {
		ledger, found := action.Ledgers.Records[operationRecord.LedgerSequence()]
		if !found {
			msg := fmt.Sprintf("could not find ledger data for sequence %d", operationRecord.LedgerSequence())
			action.Err = errors.New(msg)
			return
		}

		var transactionRecord *history.Transaction
		if action.IncludeTransactions {
			transactionRecord = &action.TransactionRecords[i]
		}

		var res hal.Pageable
		res, action.Err = resourceadapter.NewOperation(action.R.Context(), operationRecord, transactionRecord, ledger)
		if action.Err != nil {
			return
		}
		action.Page.Add(res)
	}

	action.Page.FullURL = action.FullURL()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}

// Interface verification
var _ actions.JSONer = (*OperationShowAction)(nil)

// OperationShowAction renders a page of operation resources.
type OperationShowAction struct {
	Action
	ID                  int64
	OperationRecord     history.Operation
	TransactionRecord   *history.Transaction
	Ledger              history.Ledger
	IncludeTransactions bool
	Resource            interface{}
}

func (action *OperationShowAction) loadParams() {
	action.ID = action.GetInt64("id")
	parsed, err := parseJoinField(&action.Action.Base)
	if err != nil {
		action.Err = err
		return
	}
	action.IncludeTransactions = parsed[joinTransactions]
}

func (action *OperationShowAction) loadRecord() {
	action.OperationRecord, action.TransactionRecord, action.Err = action.HistoryQ().OperationByID(
		action.IncludeTransactions, action.ID,
	)
	if action.Err != nil {
		return
	}

	if action.IncludeTransactions {
		if action.TransactionRecord == nil {
			action.Err = errors.Errorf("could not find transaction for operation %v", action.ID)
			return
		}
		action.Err = validateTransactionForOperation(*action.TransactionRecord, action.OperationRecord)
	}
}

func (action *OperationShowAction) loadLedger() {
	action.Err = action.HistoryQ().LedgerBySequence(&action.Ledger, action.OperationRecord.LedgerSequence())
}

func (action *OperationShowAction) loadResource() {
	action.Resource, action.Err = resourceadapter.NewOperation(
		action.R.Context(),
		action.OperationRecord,
		action.TransactionRecord,
		action.Ledger,
	)
}

// JSON is a method for actions.JSON
func (action *OperationShowAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.verifyWithinHistory,
		action.loadRecord,
		action.loadLedger,
		action.loadResource,
		func() { hal.Render(action.W, action.Resource) },
	)
	return action.Err
}

func (action *OperationShowAction) verifyWithinHistory() {
	parsed := toid.Parse(action.ID)
	if parsed.LedgerSequence < ledger.CurrentState().HistoryElder {
		action.Err = &problem.BeforeHistory
	}
}
