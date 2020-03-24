package aurora

import (
	"fmt"
	"net/http"

	"github.com/hcnet/go/services/aurora/internal/actions"
	"github.com/hcnet/go/services/aurora/internal/operationfeestats"
	"github.com/hcnet/go/support/render/hal"
	"github.com/hcnet/go/support/render/problem"
)

// This file contains the actions:
//
// OperationFeeStatsAction: stats representing current state of network fees

var _ actions.JSONer = (*OperationFeeStatsAction)(nil)

// OperationFeeStatsAction renders a few useful statistics that describe the
// current state of operation fees on the network.
type OperationFeeStatsAction struct {
	Action
	FeeMin              int64
	FeeMode             int64
	FeeP10              int64
	FeeP20              int64
	FeeP30              int64
	FeeP40              int64
	FeeP50              int64
	FeeP60              int64
	FeeP70              int64
	FeeP80              int64
	FeeP90              int64
	FeeP95              int64
	FeeP99              int64
	LedgerCapacityUsage string
	LastBaseFee         int64
	LastLedger          int64
}

// JSON is a method for actions.JSON
func (action *OperationFeeStatsAction) JSON() error {
	if !action.App.config.IngestFailedTransactions {
		// If Aurora is not ingesting failed transaction it does not make sense to display
		// operation fee stats because they will be incorrect.
		p := problem.P{
			Type:   "endpoint_not_available",
			Title:  "Endpoint Not Available",
			Status: http.StatusNotImplemented,
			Detail: "/operation_fee_stats is unavailable when Aurora is not ingesting failed " +
				"transactions. Set `INGEST_FAILED_TRANSACTIONS=true` to start ingesting them.",
		}
		problem.Render(action.R.Context(), action.W, p)
		return nil
	}

	action.Do(
		action.loadRecords,
		func() {
			hal.Render(action.W, map[string]string{
				"min_accepted_fee":      fmt.Sprint(action.FeeMin),
				"mode_accepted_fee":     fmt.Sprint(action.FeeMode),
				"p10_accepted_fee":      fmt.Sprint(action.FeeP10),
				"p20_accepted_fee":      fmt.Sprint(action.FeeP20),
				"p30_accepted_fee":      fmt.Sprint(action.FeeP30),
				"p40_accepted_fee":      fmt.Sprint(action.FeeP40),
				"p50_accepted_fee":      fmt.Sprint(action.FeeP50),
				"p60_accepted_fee":      fmt.Sprint(action.FeeP60),
				"p70_accepted_fee":      fmt.Sprint(action.FeeP70),
				"p80_accepted_fee":      fmt.Sprint(action.FeeP80),
				"p90_accepted_fee":      fmt.Sprint(action.FeeP90),
				"p95_accepted_fee":      fmt.Sprint(action.FeeP95),
				"p99_accepted_fee":      fmt.Sprint(action.FeeP99),
				"ledger_capacity_usage": action.LedgerCapacityUsage,
				"last_ledger_base_fee":  fmt.Sprint(action.LastBaseFee),
				"last_ledger":           fmt.Sprint(action.LastLedger),
			})
		},
	)
	return action.Err
}

func (action *OperationFeeStatsAction) loadRecords() {
	cur := operationfeestats.CurrentState()
	action.FeeMin = cur.FeeMin
	action.FeeMode = cur.FeeMode
	action.LastBaseFee = cur.LastBaseFee
	action.LastLedger = cur.LastLedger
	action.LedgerCapacityUsage = cur.LedgerCapacityUsage
	action.FeeP10 = cur.FeeP10
	action.FeeP20 = cur.FeeP20
	action.FeeP30 = cur.FeeP30
	action.FeeP40 = cur.FeeP40
	action.FeeP50 = cur.FeeP50
	action.FeeP60 = cur.FeeP60
	action.FeeP70 = cur.FeeP70
	action.FeeP80 = cur.FeeP80
	action.FeeP90 = cur.FeeP90
	action.FeeP95 = cur.FeeP95
	action.FeeP99 = cur.FeeP99
}
