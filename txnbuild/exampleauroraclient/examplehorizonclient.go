// Package exampleauroraclient provides a dummy client for use with the GoDoc examples.
package exampleauroraclient

import (
	hProtocol "github.com/hcnet/go/protocols/aurora"
)

// AccountRequest is a simple mock
type AccountRequest struct {
	AccountID string
}

// Client is a simple mock
type Client struct {
}

// DefaultTestNetClient is a simple mock
var DefaultTestNetClient = Client{}

// AccountDetail returns a minimal, static Account object
func (client *Client) AccountDetail(req AccountRequest) (hProtocol.Account, error) {
	return hProtocol.Account{
		AccountID: req.AccountID,
		Sequence:  "3556091187167235",
	}, nil
}

// FeeStats returns mock network fee information
func (client *Client) FeeStats() (hProtocol.FeeStats, error) {
	return hProtocol.FeeStats{
		LastLedger:          22606298,
		LastLedgerBaseFee:   100,
		LedgerCapacityUsage: 0.97,
		MinAcceptedFee:      100,
		ModeAcceptedFee:     200,
		P10AcceptedFee:      250,
		P20AcceptedFee:      300,
		P30AcceptedFee:      350,
		P40AcceptedFee:      500,
		P50AcceptedFee:      600,
		P60AcceptedFee:      700,
		P70AcceptedFee:      800,
		P80AcceptedFee:      900,
		P90AcceptedFee:      2000,
		P95AcceptedFee:      3000,
		P99AcceptedFee:      5000,
	}, nil
}
