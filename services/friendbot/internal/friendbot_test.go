package internal

import (
	"sync"
	"testing"

	"github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/keypair"
	hProtocol "github.com/diamnet/go/protocols/aurora"
	"github.com/stretchr/testify/assert"
)

func TestFriendbot_Pay(t *testing.T) {
	mockSubmitTransaction := func(minion *Minion, hclient *auroraclient.Client, tx string) (*hProtocol.TransactionSuccess, error) {
		// Instead of submitting the tx, we emulate a success.
		txSuccess := hProtocol.TransactionSuccess{Env: tx}
		return &txSuccess, nil
	}

	// Public key: GD25B4QI6KWVDWXDW25CIM7EKR6A6PBSWE2RCNSAC4NJQDQJXZJYMMKR
	botSeed := "SCWNLYELENPBXN46FHYXETT5LJCYBZD5VUQQVW4KZPHFO2YTQJUWT4D5"
	botKeypair, err := keypair.Parse(botSeed)
	if !assert.NoError(t, err) {
		return
	}
	botAccount := Account{AccountID: botKeypair.Address()}

	// Public key: GD4AGPPDFFHKK3Z2X4XZDRXX6GZQKP4FMLVQ5T55NDEYGG3GIP7BQUHM
	minionSeed := "SDTNSEERJPJFUE2LSDNYBFHYGVTPIWY7TU2IOJZQQGLWO2THTGB7NU5A"
	minionKeypair, err := keypair.Parse(minionSeed)
	if !assert.NoError(t, err) {
		return
	}

	minion := Minion{
		Account: Account{
			AccountID: minionKeypair.Address(),
			Sequence:  1,
		},
		Keypair:           minionKeypair.(*keypair.Full),
		BotAccount:        botAccount,
		BotKeypair:        botKeypair.(*keypair.Full),
		Network:           "Test SDF Network ; September 2015",
		StartingBalance:   "10000.00",
		SubmitTransaction: mockSubmitTransaction,
	}
	fb := &Bot{Minions: []Minion{minion}}

	recipientAddress := "GDJIN6W6PLTPKLLM57UW65ZH4BITUXUMYQHIMAZFYXF45PZVAWDBI77Z"
	txSuccess, err := fb.Pay(recipientAddress)
	if !assert.NoError(t, err) {
		return
	}
	expectedTxn := "AAAAAPgDPeMpTqVvOr8vkcb38bMFP4Vi6w7PvWjJgxtmQ/4YAAAAZAAAAAAAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAA9dDyCPKtUdrjtrokM+RUfA88MrE1ETZAFxqYDgm+U4YAAAAAAAAAANKG+t565vUtbO/pb3cn4FE6XozEDoYDJcXLzr81BYYUAAAAF0h26AAAAAAAAAAAAmZD/hgAAABANEsSWMNVgAudOT2YNx5AR3k+uNDITctQCOy0jJNYfm39M/3T0XrpOAR8EUozFIoXp+Rrtm49xKzjSLHgCiYSCgm+U4YAAABA9Iazzw7Be5vPtRPqcWG+EXjsRB9o6yaIiw6SODNSuYGjKklBOYwxuB6LHSR1t8epLvn6J58ml1cs0UOt4afGAQ=="
	assert.Equal(t, expectedTxn, txSuccess.Env)

	// Don't assert on tx values below, since the completion order is unknown.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, err := fb.Pay(recipientAddress)
		assert.NoError(t, err)
		wg.Done()
	}()
	go func() {
		_, err := fb.Pay(recipientAddress)
		assert.NoError(t, err)
		wg.Done()
	}()
	wg.Wait()
}
