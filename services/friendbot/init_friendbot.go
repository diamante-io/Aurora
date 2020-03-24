package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/services/friendbot/internal"
	"github.com/hcnet/go/strkey"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/txnbuild"
)

func initFriendbot(
	friendbotSecret string,
	networkPassphrase string,
	auroraURL string,
	startingBalance string,
	numMinions int,
) (*internal.Bot, error) {
	if friendbotSecret == "" || networkPassphrase == "" || auroraURL == "" || startingBalance == "" || numMinions < 0 {
		return nil, errors.New("invalid input param(s)")
	}

	// Guarantee that friendbotSecret is a seed, if not blank.
	strkey.MustDecode(strkey.VersionByteSeed, friendbotSecret)

	hclient := &auroraclient.Client{
		AuroraURL: auroraURL,
		HTTP:       http.DefaultClient,
		AppName:    "friendbot",
	}

	botKP, err := keypair.Parse(friendbotSecret)
	if err != nil {
		return nil, errors.Wrap(err, "parsing bot keypair")
	}

	// Casting from the interface type will work, since we
	// already confirmed that friendbotSecret is a seed.
	botKeypair := botKP.(*keypair.Full)
	botAccount := internal.Account{AccountID: botKeypair.Address()}
	minionBalance := "101.00"
	if numMinions == 0 {
		numMinions = 1000
	}
	log.Printf("Found all valid params, now creating %d minions", numMinions)
	minions, err := createMinionAccounts(botAccount, botKeypair, networkPassphrase, startingBalance, minionBalance, numMinions, hclient)
	if err != nil && len(minions) == 0 {
		return nil, errors.Wrap(err, "creating minion accounts")
	}
	log.Printf("Adding %d minions to friendbot", len(minions))
	return &internal.Bot{Minions: minions}, nil
}

func createMinionAccounts(botAccount internal.Account, botKeypair *keypair.Full, networkPassphrase, newAccountBalance, minionBalance string, numMinions int, hclient *auroraclient.Client) ([]internal.Minion, error) {
	var minions []internal.Minion
	numRemainingMinions := numMinions
	minionBatchSize := 100
	for numRemainingMinions > 0 {
		var (
			newMinions []internal.Minion
			ops        []txnbuild.Operation
		)
		// Refresh the sequence number before submitting a new transaction.
		rerr := botAccount.RefreshSequenceNumber(hclient)
		if rerr != nil {
			return minions, errors.Wrap(rerr, "refreshing bot seqnum")
		}
		// The tx will create min(numRemainingMinions, 100) Minion accounts.
		numCreateMinions := minionBatchSize
		if numRemainingMinions < minionBatchSize {
			numCreateMinions = numRemainingMinions
		}
		log.Printf("Creating %d new minion accounts", numCreateMinions)
		for i := 0; i < numCreateMinions; i++ {
			minionKeypair, err := keypair.Random()
			if err != nil {
				return minions, errors.Wrap(err, "making keypair")
			}
			newMinions = append(newMinions, internal.Minion{
				Account:           internal.Account{AccountID: minionKeypair.Address()},
				Keypair:           minionKeypair,
				BotAccount:        botAccount,
				BotKeypair:        botKeypair,
				Aurora:           hclient,
				Network:           networkPassphrase,
				StartingBalance:   newAccountBalance,
				SubmitTransaction: internal.SubmitTransaction,
			})

			ops = append(ops, &txnbuild.CreateAccount{
				Destination: minionKeypair.Address(),
				Amount:      minionBalance,
			})
		}

		// Build and submit batched account creation tx.
		txn := txnbuild.Transaction{
			SourceAccount: botAccount,
			Operations:    ops,
			Timebounds:    txnbuild.NewTimeout(300),
			Network:       networkPassphrase,
		}
		txe, err := txn.BuildSignEncode(botKeypair)
		if err != nil {
			return minions, errors.Wrap(err, "making create accounts tx")
		}
		resp, err := hclient.SubmitTransactionXDR(txe)
		if err != nil {
			log.Println(resp)
			switch e := err.(type) {
			case *auroraclient.Error:
				problemString := fmt.Sprintf("Problem[Type=%s, Title=%s, Status=%d, Detail=%s, Extras=%v]", e.Problem.Type, e.Problem.Title, e.Problem.Status, e.Problem.Detail, e.Problem.Extras)
				return minions, errors.Wrap(errors.Wrap(e, problemString), "submitting create accounts tx")
			}
			return minions, errors.Wrap(err, "submitting create accounts tx")
		}

		// Process successful create accounts tx.
		numRemainingMinions -= numCreateMinions
		minions = append(minions, newMinions...)
		log.Printf("Submitted create accounts tx for %d minions successfully", numCreateMinions)
	}
	return minions, nil
}
