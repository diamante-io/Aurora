package aurora

import (
	"encoding/json"
	"testing"
)

func TestOperationFeeTestsActions_Show(t *testing.T) {
	testCases := []struct {
		scenario            string
		lastbasefee         string
		min                 string
		mode                string
		p10                 string
		p20                 string
		p30                 string
		p40                 string
		p50                 string
		p60                 string
		p70                 string
		p80                 string
		p90                 string
		p95                 string
		p99                 string
		ledgerCapacityUsage string
	}{
		// happy path
		{
			"operation_fee_stats_1",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"0.04",
		},
		// no transactions in last 5 ledgers
		{
			"operation_fee_stats_2",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"100",
			"0.00",
		},
		// transactions with varying fees
		{
			"operation_fee_stats_3",
			"100",
			"200",
			"400",
			"260", // p10
			"320",
			"380",
			"400",
			"400",
			"400",
			"400",
			"400",
			"400",
			"400",
			"400",
			"0.03",
		},
	}

	for _, kase := range testCases {
		t.Run("/fee_stats", func(t *testing.T) {
			ht := StartHTTPTest(t, kase.scenario)
			defer ht.Finish()

			// Update max_tx_set_size on ledgers
			_, err := ht.AuroraSession().ExecRaw("UPDATE history_ledgers SET max_tx_set_size = 50")
			ht.Require.NoError(err)

			ht.App.UpdateOperationFeeStatsState()

			w := ht.Get("/fee_stats")

			if ht.Assert.Equal(200, w.Code) {
				var result map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &result)
				ht.Require.NoError(err)
				ht.Assert.Equal(kase.min, result["min_accepted_fee"], "min")
				ht.Assert.Equal(kase.mode, result["mode_accepted_fee"], "mode")
				ht.Assert.Equal(kase.lastbasefee, result["last_ledger_base_fee"], "base_fee")
				ht.Assert.Equal(kase.p10, result["p10_accepted_fee"], "p10")
				ht.Assert.Equal(kase.p20, result["p20_accepted_fee"], "p20")
				ht.Assert.Equal(kase.p30, result["p30_accepted_fee"], "p30")
				ht.Assert.Equal(kase.p40, result["p40_accepted_fee"], "p40")
				ht.Assert.Equal(kase.p50, result["p50_accepted_fee"], "p50")
				ht.Assert.Equal(kase.p60, result["p60_accepted_fee"], "p60")
				ht.Assert.Equal(kase.p70, result["p70_accepted_fee"], "p70")
				ht.Assert.Equal(kase.p80, result["p80_accepted_fee"], "p80")
				ht.Assert.Equal(kase.p90, result["p90_accepted_fee"], "p90")
				ht.Assert.Equal(kase.p95, result["p95_accepted_fee"], "p95")
				ht.Assert.Equal(kase.p99, result["p99_accepted_fee"], "p99")
				ht.Assert.Equal(kase.ledgerCapacityUsage, result["ledger_capacity_usage"], "ledger_capacity_usage")
			}
		})
	}
}

// TestOperationFeeTestsActions_ShowMultiOp tests fee stats in case transactions contain multiple operations.
// In such case, since protocol v11, we should use number of operations as the indicator of ledger capacity usage.
func TestOperationFeeTestsActions_ShowMultiOp(t *testing.T) {
	ht := StartHTTPTest(t, "operation_fee_stats_3")
	defer ht.Finish()

	// Update max_tx_set_size on ledgers
	_, err := ht.AuroraSession().ExecRaw("UPDATE history_ledgers SET max_tx_set_size = 50")
	ht.Require.NoError(err)

	// Update number of ops on each transaction
	_, err = ht.AuroraSession().ExecRaw("UPDATE history_transactions SET operation_count = operation_count * 2")
	ht.Require.NoError(err)

	ht.App.UpdateOperationFeeStatsState()

	w := ht.Get("/fee_stats")

	if ht.Assert.Equal(200, w.Code) {
		var result map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &result)
		ht.Require.NoError(err)
		ht.Assert.Equal("100", result["min_accepted_fee"], "min")
		ht.Assert.Equal("200", result["mode_accepted_fee"], "mode")
		ht.Assert.Equal("100", result["last_ledger_base_fee"], "base_fee")
		ht.Assert.Equal("130", result["p10_accepted_fee"], "p10")
		ht.Assert.Equal("160", result["p20_accepted_fee"], "p20")
		ht.Assert.Equal("190", result["p30_accepted_fee"], "p30")
		ht.Assert.Equal("200", result["p40_accepted_fee"], "p40")
		ht.Assert.Equal("200", result["p50_accepted_fee"], "p50")
		ht.Assert.Equal("200", result["p60_accepted_fee"], "p60")
		ht.Assert.Equal("200", result["p70_accepted_fee"], "p70")
		ht.Assert.Equal("200", result["p80_accepted_fee"], "p80")
		ht.Assert.Equal("200", result["p90_accepted_fee"], "p90")
		ht.Assert.Equal("200", result["p95_accepted_fee"], "p95")
		ht.Assert.Equal("200", result["p99_accepted_fee"], "p99")
		ht.Assert.Equal("0.06", result["ledger_capacity_usage"], "ledger_capacity_usage")
	}
}
