package submitter

import (
	"fmt"
	"testing"
	"time"

	hc "github.com/diamnet/go/clients/auroraclient"
	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/bridge/internal/db"
	"github.com/diamnet/go/services/bridge/internal/mocks"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/txnbuild"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionSubmitter(t *testing.T) {
	var mockAurora = new(hc.MockClient)
	var mockDatabase = new(mocks.MockDatabase)
	mocks.PredefinedTime = time.Now()
	seed := "SDZT3EJZ7FZRYNTLOZ7VH6G5UYBFO2IO3Q5PGONMILPCZU3AL7QNZHTE"
	accountID := "GCLOMB72ODBFUGK4E2BK7VMR3RNZ5WSTMEOGNA2YUVHFR3WMH2XBAB6H"
	transactionSubmitter := NewTransactionSubmitter(mockAurora, mockDatabase, "Test SDF Network ; September 2015", mocks.Now)

	// When seed is invalid
	_, err := transactionSubmitter.LoadAccount("invalidSeed")
	assert.NotNil(t, err)

	// When there is an error loading account
	mockAurora.On(
		"AccountDetail",
		mock.AnythingOfType("auroraclient.AccountRequest"),
	).Return(
		hProtocol.Account{},
		errors.New("Account not found"),
	).Once()

	_, err = transactionSubmitter.LoadAccount(seed)
	assert.NotNil(t, err)
	mockAurora.AssertExpectations(t)

	// successfully loads an account
	transactionSubmitter = NewTransactionSubmitter(mockAurora, mockDatabase, "Test SDF Network ; September 2015", mocks.Now)

	mockAurora.On(
		"AccountDetail",
		mock.AnythingOfType("auroraclient.AccountRequest"),
	).Return(
		hProtocol.Account{
			ID:        accountID,
			AccountID: accountID,
			Sequence:  "10372672437354496",
		},
		nil,
	).Once()

	account, err := transactionSubmitter.LoadAccount(seed)
	assert.Nil(t, err)
	assert.Equal(t, account.Keypair.Address(), accountID)
	assert.Equal(t, account.Seed, seed)
	assert.Equal(t, account.SequenceNumber, uint64(10372672437354496))
	mockAurora.AssertExpectations(t)

	// Submit transaction - Error response from aurora
	transactionSubmitter = NewTransactionSubmitter(mockAurora, mockDatabase, "Test SDF Network ; September 2015", mocks.Now)

	mockAurora.On(
		"AccountDetail",
		mock.AnythingOfType("auroraclient.AccountRequest"),
	).Return(
		hProtocol.Account{
			ID:        accountID,
			AccountID: accountID,
			Sequence:  "10372672437354496",
		},
		nil,
	).Once()

	err = transactionSubmitter.InitAccount(seed)
	assert.Nil(t, err)

	txB64 := "AAAAAJbmB/pwwloZXCaCr9WR3Fue2lNhHGaDWKVOWO7MPq4QAAAAZAAk2eQAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAABAAAAAHdv1hoGkOgiXF0LRkRaHa7m0GfXIuWT0ZcaajT1ldQgAAAAAAAAAAA7msoAAAAAAAAAAAHMPq4QAAAAQMk5tSJngfsKsfYxK5VqfFCSwgqGatSnp54Lm+WVrMD5wNVFMaHHflIJrzUDS0+/uTeh6lzpIRHRYRUOTAfKpAc="

	// Persist sending transaction
	mockDatabase.On(
		"InsertSentTransaction",
		mock.AnythingOfType("*db.SentTransaction"),
	).Return(nil).Once().Run(func(args mock.Arguments) {
		transaction := args.Get(0).(*db.SentTransaction)
		assert.Equal(t, "b2ce8447092ccfbb486aa9d38251a19df7df8df16c0c59edc55fee4d9727626e", transaction.TransactionID)
		assert.Equal(t, "sending", string(transaction.Status))
		assert.Equal(t, "GCLOMB72ODBFUGK4E2BK7VMR3RNZ5WSTMEOGNA2YUVHFR3WMH2XBAB6H", transaction.Source)
		assert.Equal(t, mocks.PredefinedTime, transaction.SubmittedAt)
		assert.Equal(t, txB64, transaction.EnvelopeXdr)
	})

	// Persist failure
	mockDatabase.On(
		"UpdateSentTransaction",
		mock.AnythingOfType("*db.SentTransaction"),
	).Return(nil).Once().Run(func(args mock.Arguments) {
		transaction := args.Get(0).(*db.SentTransaction)
		assert.Equal(t, "b2ce8447092ccfbb486aa9d38251a19df7df8df16c0c59edc55fee4d9727626e", transaction.TransactionID)
		fmt.Println("txStatus: ", transaction.Status)
		assert.Equal(t, "failure", string(transaction.Status))
		assert.Equal(t, "GCLOMB72ODBFUGK4E2BK7VMR3RNZ5WSTMEOGNA2YUVHFR3WMH2XBAB6H", transaction.Source)
		assert.Equal(t, mocks.PredefinedTime, transaction.SubmittedAt)
		assert.Equal(t, txB64, transaction.EnvelopeXdr)
	})

	mockAurora.On("SubmitTransactionXDR", txB64).Return(
		hProtocol.TransactionSuccess{
			Ledger: 0,
			Result: "AAAAAAAAAGT/////AAAAAQAAAAAAAAAB////+wAAAAA=", // no_destination

		},
		errors.New("tx failed"),
	).Once()

	txnOp := &txnbuild.Payment{
		Destination: "GB3W7VQ2A2IOQIS4LUFUMRC2DWXONUDH24ROLE6RS4NGUNHVSXKCABOM",
		Amount:      "100",
		Asset:       txnbuild.NativeAsset{},
	}

	_, err = transactionSubmitter.SubmitTransaction((*string)(nil), seed, []txnbuild.Operation{txnOp}, nil)
	assert.Nil(t, err)
	mockAurora.AssertExpectations(t)

	// Submit transaction - success response
	transactionSubmitter = NewTransactionSubmitter(mockAurora, mockDatabase, "Test SDF Network ; September 2015", mocks.Now)

	mockAurora.On(
		"AccountDetail",
		mock.AnythingOfType("auroraclient.AccountRequest"),
	).Return(
		hProtocol.Account{
			ID:        accountID,
			AccountID: accountID,
			Sequence:  "10372672437354496",
		},
		nil,
	).Once()

	err = transactionSubmitter.InitAccount(seed)
	assert.Nil(t, err)

	txB64 = "AAAAAJbmB/pwwloZXCaCr9WR3Fue2lNhHGaDWKVOWO7MPq4QAAAAZAAk2eQAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAHdv1hoGkOgiXF0LRkRaHa7m0GfXIuWT0ZcaajT1ldQgAAAAADuaygAAAAAAAAAAAcw+rhAAAABAPgwRbiJH9d4zukMq8ULwe88YCLniYFbq9YgryxS+VmYIJ7N6KKbsRMWi2LDMovRY2I6f3GG8eBHUh0JCwTbdCg=="

	// Persist sending transaction
	mockDatabase.On(
		"InsertSentTransaction",
		mock.AnythingOfType("*db.SentTransaction"),
	).Return(nil).Once().Run(func(args mock.Arguments) {
		transaction := args.Get(0).(*db.SentTransaction)
		assert.Equal(t, "90c541957386b0325e66cc1308cfdfcffc5d95fe30210ade896f50a168839a13", transaction.TransactionID)
		assert.Equal(t, "sending", string(transaction.Status))
		assert.Equal(t, "GCLOMB72ODBFUGK4E2BK7VMR3RNZ5WSTMEOGNA2YUVHFR3WMH2XBAB6H", transaction.Source)
		assert.Equal(t, mocks.PredefinedTime, transaction.SubmittedAt)
		assert.Equal(t, txB64, transaction.EnvelopeXdr)
	})

	// Persist success
	mockDatabase.On(
		"UpdateSentTransaction",
		mock.AnythingOfType("*db.SentTransaction"),
	).Return(nil).Once().Run(func(args mock.Arguments) {
		transaction := args.Get(0).(*db.SentTransaction)
		assert.Equal(t, "90c541957386b0325e66cc1308cfdfcffc5d95fe30210ade896f50a168839a13", transaction.TransactionID)
		assert.Equal(t, "success", string(transaction.Status))
		assert.Equal(t, "GCLOMB72ODBFUGK4E2BK7VMR3RNZ5WSTMEOGNA2YUVHFR3WMH2XBAB6H", transaction.Source)
		assert.Equal(t, mocks.PredefinedTime, transaction.SubmittedAt)
		assert.Equal(t, txB64, transaction.EnvelopeXdr)
	})

	mockAurora.On("SubmitTransactionXDR", txB64).Return(
		hProtocol.TransactionSuccess{
			Ledger: int32(123),
			Result: "AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAA=",
		},
		nil,
	).Once()

	txnOp2 := &txnbuild.CreateAccount{
		Destination: "GB3W7VQ2A2IOQIS4LUFUMRC2DWXONUDH24ROLE6RS4NGUNHVSXKCABOM",
		Amount:      "100",
	}

	_, err = transactionSubmitter.SubmitTransaction((*string)(nil), seed, []txnbuild.Operation{txnOp2}, nil)
	assert.Nil(t, err)
	mockAurora.AssertExpectations(t)

}
