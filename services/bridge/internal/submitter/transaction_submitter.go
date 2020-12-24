package submitter

import (
	"database/sql"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	hc "github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/keypair"
	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/bridge/internal/db"
	shared "github.com/diamnet/go/services/internal/bridge-compliance-shared"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/txnbuild"
	"github.com/diamnet/go/xdr"
)

// TransactionSubmitterInterface helps mocking TransactionSubmitter
type TransactionSubmitterInterface interface {
	SubmitTransaction(paymentID *string, seed string, operation []txnbuild.Operation, memo txnbuild.Memo) (response hProtocol.TransactionSuccess, err error)
	SignAndSubmitRawTransaction(paymentID *string, seed string, tx *xdr.Transaction) (response hProtocol.TransactionSuccess, err error)
}

// TransactionSubmitter submits transactions to DiamNet Network
type TransactionSubmitter struct {
	Aurora       hc.ClientInterface
	Accounts      map[string]*Account // seed => *Account
	AccountsMutex sync.Mutex
	Database      db.Database
	Network       string
	log           *logrus.Entry
	now           func() time.Time
}

// Account represents account used to signing and sending transactions
type Account struct {
	Keypair        keypair.KP
	Seed           string
	SequenceNumber uint64
	Mutex          sync.Mutex
}

// NewTransactionSubmitter creates a new TransactionSubmitter
func NewTransactionSubmitter(
	aurora hc.ClientInterface,
	database db.Database,
	networkPassphrase string,
	now func() time.Time,
) (ts TransactionSubmitter) {
	ts.Aurora = aurora
	ts.Database = database
	ts.Accounts = make(map[string]*Account)
	ts.Network = networkPassphrase
	ts.log = logrus.WithFields(logrus.Fields{
		"service": "TransactionSubmitter",
	})
	ts.now = now
	return
}

// LoadAccount loads current state of DiamNet account and creates a map entry if it didn't exist
func (ts *TransactionSubmitter) LoadAccount(seed string) (*Account, error) {
	ts.AccountsMutex.Lock()

	account, exist := ts.Accounts[seed]
	if exist {
		ts.AccountsMutex.Unlock()
		return account, nil
	}

	kp, err := keypair.Parse(seed)
	if err != nil {
		ts.log.Print("Invalid seed")
		ts.AccountsMutex.Unlock()
		return nil, err
	}

	ts.Accounts[seed] = &Account{
		Seed:    seed,
		Keypair: kp,
	}
	ts.AccountsMutex.Unlock()

	// Load account sequence number
	ts.Accounts[seed].Mutex.Lock()
	defer ts.Accounts[seed].Mutex.Unlock()

	if ts.Accounts[seed].SequenceNumber != 0 {
		return ts.Accounts[seed], nil
	}

	accountRequest := hc.AccountRequest{AccountID: ts.Accounts[seed].Keypair.Address()}
	accountResponse, err := ts.Aurora.AccountDetail(accountRequest)
	if err != nil {
		return nil, err
	}

	ts.Accounts[seed].SequenceNumber, err = strconv.ParseUint(accountResponse.Sequence, 10, 64)
	if err != nil {
		return nil, err
	}

	return ts.Accounts[seed], nil
}

// InitAccount loads an account and returns error if it fails
func (ts *TransactionSubmitter) InitAccount(seed string) (err error) {
	_, err = ts.LoadAccount(seed)
	return
}

// SignAndSubmitRawTransaction will:
// - update sequence number of the transaction to the current one,
// - sign it,
// - submit it to the network.
func (ts *TransactionSubmitter) SignAndSubmitRawTransaction(paymentID *string, seed string, tx *xdr.Transaction) (response hProtocol.TransactionSuccess, err error) {
	account, err := ts.LoadAccount(seed)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Error loading account")
		return
	}

	account.Mutex.Lock()
	account.SequenceNumber++
	tx.SeqNum = xdr.SequenceNumber(account.SequenceNumber)
	account.Mutex.Unlock()

	hash, err := shared.TransactionHash(tx, ts.Network)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Error calculating transaction hash")
		return
	}

	sig, err := account.Keypair.SignDecorated(hash[:])
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Error signing a transaction")
		return
	}

	envelopeXdr := xdr.TransactionEnvelope{
		Tx:         *tx,
		Signatures: []xdr.DecoratedSignature{sig},
	}

	txeB64, err := xdr.MarshalBase64(envelopeXdr)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Cannot encode transaction envelope")
		return
	}

	transactionHashBytes, err := shared.TransactionHash(tx, ts.Network)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Warn("Error calculating tx hash")
		return
	}

	var herr *hc.Error
	response, err = ts.SubmitAndSave(paymentID, account.Keypair.Address(), txeB64, hex.EncodeToString(transactionHashBytes[:]))
	if err != nil {
		var isAuroraError bool
		herr, isAuroraError = err.(*hc.Error)
		if !isAuroraError {
			ts.log.WithFields(logrus.Fields{"err": err}).Error("Error submitting transaction ", err)
			return
		}
	}

	// Sync sequence number
	if herr != nil {
		codes, rerr := herr.ResultCodes()
		if rerr != nil {
			return response, herr
		}

		if codes.TransactionCode != "tx_bad_seq" {
			return response, herr
		}

		account.Mutex.Lock()
		ts.log.Print("Syncing sequence number for ", account.Keypair.Address())

		accountRequest := hc.AccountRequest{AccountID: account.Keypair.Address()}
		accountResponse, err := ts.Aurora.AccountDetail(accountRequest)
		if err != nil {
			ts.log.Error("Error updating sequence number ", err)
		} else {
			account.SequenceNumber, _ = strconv.ParseUint(accountResponse.Sequence, 10, 64)
		}
		account.Mutex.Unlock()

		return response, herr
	}
	return
}

// SubmitTransaction builds and submits transaction to DiamNet network
func (ts *TransactionSubmitter) SubmitTransaction(paymentID *string, seed string, operation []txnbuild.Operation, memo txnbuild.Memo) (hProtocol.TransactionSuccess, error) {
	account, err := ts.LoadAccount(seed)
	if err != nil {
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "Error loading an account")
	}

	tx := txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: account.Keypair.Address(), Sequence: int64(account.SequenceNumber)},
		Operations:    operation,
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       ts.Network,
		Memo:          memo,
	}

	err = tx.Build()
	if err != nil {
		ts.log.Error("Unable to build transaction")
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "unable to build transaction")
	}

	kp, err := keypair.Parse(seed)
	if err != nil {
		ts.log.Error("Unable to convert seed to keypair")
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "unable to convert seed to keypair")
	}

	err = tx.Sign(kp.(*keypair.Full))
	if err != nil {
		ts.log.Error("Unable to sign transaction")
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "unable to sign transaction")
	}

	txe, err := tx.Base64()
	if err != nil {
		ts.log.Error("Unable to encode transaction")
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "unable to encode transaction")
	}

	txHashBytes, err := tx.Hash()
	if err != nil {
		ts.log.Error("Unable to get transaction hash")
		return hProtocol.TransactionSuccess{}, errors.Wrap(err, "unable to get transaction hash")
	}

	return ts.SubmitAndSave(paymentID, tx.SourceAccount.GetAccountID(), txe, hex.EncodeToString(txHashBytes[:]))
}

// SubmitAndSave sumbits a transaction to aurora and saves the details in the bridge server database.
func (ts *TransactionSubmitter) SubmitAndSave(paymentID *string, sourceAccount, txeB64, txHash string) (response hProtocol.TransactionSuccess, err error) {
	nullPaymentID := sql.NullString{Valid: false}
	if paymentID != nil {
		nullPaymentID = sql.NullString{
			String: *paymentID,
			Valid:  true,
		}
	}

	sentTransaction := &db.SentTransaction{
		PaymentID:     nullPaymentID,
		TransactionID: txHash,
		Status:        db.SentTransactionStatusSending,
		Source:        sourceAccount,
		SubmittedAt:   ts.now(),
		EnvelopeXdr:   txeB64,
	}

	err = ts.Database.InsertSentTransaction(sentTransaction)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Error inserting sent transaction")
		return
	}

	ts.log.WithFields(logrus.Fields{"tx": txeB64}).Info("Submitting transaction")

	var herr *hc.Error
	response, err = ts.Aurora.SubmitTransactionXDR(txeB64)
	if err == nil {
		sentTransaction.Status = db.SentTransactionStatusSuccess
		sentTransaction.Ledger = &response.Ledger
		now := time.Now()
		sentTransaction.SucceededAt = &now
	} else {
		var isAuroraError bool
		herr, isAuroraError = err.(*hc.Error)
		if !isAuroraError {
			ts.log.WithFields(logrus.Fields{"err": err}).Error("Error submitting transaction ", err)
		} else {
			var result string
			result, err = herr.ResultString()
			if err != nil {
				result = errors.Wrap(err, "Error getting tx result").Error()
			}
			sentTransaction.ResultXdr = &result
		}
		sentTransaction.Status = db.SentTransactionStatusFailure
	}

	err = ts.Database.UpdateSentTransaction(sentTransaction)
	if err != nil {
		ts.log.WithFields(logrus.Fields{"err": err}).Error("Error updating sent transaction")
		return
	}

	return
}
