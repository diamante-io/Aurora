package internal

import (
	"fmt"

	"github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/keypair"
	hProtocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/txnbuild"
)

const createAccountAlreadyExistXDR = "AAAAAAAAAGT/////AAAAAQAAAAAAAAAA/////AAAAAA="

var ErrAccountExists error = errors.New(fmt.Sprintf("createAccountAlreadyExist (%s)", createAccountAlreadyExistXDR))

// Minion contains a HcNet channel account and Go channels to communicate with friendbot.
type Minion struct {
	Account           Account
	Keypair           *keypair.Full
	BotAccount        txnbuild.Account
	BotKeypair        *keypair.Full
	Aurora           *auroraclient.Client
	Network           string
	StartingBalance   string
	SubmitTransaction func(minion *Minion, hclient *auroraclient.Client, tx string) (*hProtocol.TransactionSuccess, error)

	// Uninitialized.
	forceRefreshSequence bool
}

// Run reads a payment destination address and an output channel. It attempts
// to pay that address and submits the result to the channel.
func (minion *Minion) Run(destAddress string, resultChan chan SubmitResult) {
	err := minion.checkSequenceRefresh(minion.Aurora)
	if err != nil {
		resultChan <- SubmitResult{
			maybeTransactionSuccess: nil,
			maybeErr:                errors.Wrap(err, "checking minion seq"),
		}
	}
	txStr, err := minion.makeTx(destAddress)
	if err != nil {
		resultChan <- SubmitResult{
			maybeTransactionSuccess: nil,
			maybeErr:                errors.Wrap(err, "making payment tx"),
		}
	}
	succ, err := minion.SubmitTransaction(minion, minion.Aurora, txStr)
	resultChan <- SubmitResult{
		maybeTransactionSuccess: succ,
		maybeErr:                errors.Wrap(err, "submitting tx to minion"),
	}
}

// SubmitTransaction should be passed to the Minion.
func SubmitTransaction(minion *Minion, hclient *auroraclient.Client, tx string) (*hProtocol.TransactionSuccess, error) {
	result, err := hclient.SubmitTransactionXDR(tx)
	if err != nil {
		errStr := "submitting tx to aurora"
		switch e := err.(type) {
		case *auroraclient.Error:
			minion.checkHandleBadSequence(e)
			resStr, resErr := e.ResultString()
			if resErr != nil {
				errStr += ": error getting aurora error code: " + resErr.Error()
			} else if resStr == createAccountAlreadyExistXDR {
				return nil, errors.Wrap(ErrAccountExists, errStr)
			} else {
				errStr += ": aurora error string: " + resStr
			}
			return nil, errors.New(errStr)
		}
		return nil, errors.Wrap(err, errStr)
	}
	return &result, nil
}

// Establishes the minion's initial sequence number, if needed.
func (minion *Minion) checkSequenceRefresh(hclient *auroraclient.Client) error {
	if minion.Account.Sequence != 0 && !minion.forceRefreshSequence {
		return nil
	}
	err := minion.Account.RefreshSequenceNumber(hclient)
	if err != nil {
		return errors.Wrap(err, "refreshing minion seqnum")
	}
	minion.forceRefreshSequence = false
	return nil
}

func (minion *Minion) checkHandleBadSequence(err *auroraclient.Error) {
	resCode, e := err.ResultCodes()
	isTxBadSeqCode := e == nil && resCode.TransactionCode == "tx_bad_seq"
	if !isTxBadSeqCode {
		return
	}
	minion.forceRefreshSequence = true
}

func (minion *Minion) makeTx(destAddress string) (string, error) {
	createAccountOp := txnbuild.CreateAccount{
		Destination:   destAddress,
		SourceAccount: minion.BotAccount,
		Amount:        minion.StartingBalance,
	}
	txn := txnbuild.Transaction{
		SourceAccount: minion.Account,
		Operations:    []txnbuild.Operation{&createAccountOp},
		Network:       minion.Network,
		Timebounds:    txnbuild.NewInfiniteTimeout(),
	}
	txe, err := txn.BuildSignEncode(minion.Keypair, minion.BotKeypair)
	if err != nil {
		return "", errors.Wrap(err, "making account payment tx")
	}
	// Increment the in-memory sequence number, since the tx will be submitted.
	_, err = minion.Account.IncrementSequenceNumber()
	if err != nil {
		return "", errors.Wrap(err, "incrementing minion seq")
	}
	return txe, err
}
