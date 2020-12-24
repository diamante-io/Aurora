---
title: Basic Examples
---

- [Creating a payment Transaction](#creating-and-submitting-a-payment-transaction)


## Create an account

The first account on TestNet needs to be created by calling friendbot, a helper service that will create and fund the
provided account address. However, on the public network, you need an initial, funded account before you can create further accounts. Typically a wallet or exchange can create an initial account for you.

In this TestNet example, we first get an account funded from friendbot, and then demonstrate the `create account`
operation to set up a second account.

```go
package main

import (
	"log"

	"github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/network"
	"github.com/diamnet/go/txnbuild"
)

func main() {
	// Generate a new randomly generated address
	pair, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Seed 0:", pair.Seed())
	log.Println("Address 0:", pair.Address())

	// Create and fund the address on TestNet, using friendbot
	client := auroraclient.DefaultTestNetClient
	client.Fund(pair.Address())

	// Get information about the account we just created
	accountRequest := auroraclient.AccountRequest{AccountID: pair.Address()}
	hAccount0, err := client.AccountDetail(accountRequest)
	if err != nil {
		log.Fatal(err)
	}

	// Generate a second randomly generated address
	kp1, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Seed 1:", kp1.Seed())
	log.Println("Address 1:", kp1.Address())

	// Construct the operation
	createAccountOp := txnbuild.CreateAccount{
		Destination: kp1.Address(),
		Amount:      "10",
	}

	// Construct the transaction that will carry the operation
	tx := txnbuild.Transaction{
		SourceAccount: &hAccount0,
		Operations:    []txnbuild.Operation{&createAccountOp},
		Timebounds:    txnbuild.NewTimeout(300),
		Network:       network.TestNetworkPassphrase,
	}

	// Sign the transaction, serialise it to XDR, and base 64 encode it
	txeBase64, err := tx.BuildSignEncode(pair)
	log.Println("Transaction base64: ", txeBase64)

	// Submit the transaction
	resp, err := client.SubmitTransactionXDR(txeBase64)
	if err != nil {
		hError := err.(*auroraclient.Error)
		log.Fatal("Error submitting transaction:", hError)
	}

	log.Println("\nTransaction response: ", resp)
}
```
