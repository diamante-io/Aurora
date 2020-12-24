package internal

import (
	"strconv"

	"github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// Account implements the `txnbuild.Account` interface.
type Account struct {
	AccountID string
	Sequence  xdr.SequenceNumber
}

// GetAccountID returns the Account ID.
func (a Account) GetAccountID() string {
	return a.AccountID
}

// IncrementSequenceNumber increments the internal record of the
// account's sequence number by 1.
func (a Account) IncrementSequenceNumber() (xdr.SequenceNumber, error) {
	a.Sequence++
	return a.Sequence, nil
}

// RefreshSequenceNumber gets an Account's correct in-memory sequence number from Aurora.
func (a *Account) RefreshSequenceNumber(hclient *auroraclient.Client) error {
	accountRequest := auroraclient.AccountRequest{AccountID: a.GetAccountID()}
	accountDetail, err := hclient.AccountDetail(accountRequest)
	if err != nil {
		return errors.Wrap(err, "getting account detail")
	}
	seq, err := strconv.ParseInt(accountDetail.Sequence, 10, 64)
	if err != nil {
		return errors.Wrap(err, "parsing account seqnum")
	}
	a.Sequence = xdr.SequenceNumber(seq)
	return nil
}
