package stress

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hcnet/go/build"
	"github.com/hcnet/go/clients/aurora"
	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/services/bifrost/common"
	"github.com/hcnet/go/services/bifrost/server"
)

func (u *Users) Start(accounts chan<- server.GenerateAddressResponse) {
	u.log = common.CreateLogger("Users")
	rand.Seed(time.Now().Unix())
	u.users = map[string]*User{}

	go func() {
		cursor := aurora.Cursor("now")
		err := u.Aurora.StreamPayments(context.Background(), u.IssuerPublicKey, &cursor, u.onNewPayment)
		if err != nil {
			panic(err)
		}
	}()

	for {
		for i := 0; i < u.UsersPerSecond; i++ {
			go func() {
				kp, err := keypair.Random()
				if err != nil {
					panic(err)
				}

				u.usersLock.Lock()
				u.users[kp.Address()] = &User{
					State:           PendingUserState,
					AccountCreated:  make(chan bool),
					PaymentReceived: make(chan bool),
				}
				u.usersLock.Unlock()

				accounts <- u.newUser(kp)
			}()
		}
		u.printStatus()
		time.Sleep(time.Second)
	}
}

func (u *Users) onNewPayment(payment aurora.Payment) {
	var destination string

	switch payment.Type {
	case "create_account":
		destination = payment.Account
	case "payment":
		destination = payment.To
	default:
		return
	}

	u.usersLock.Lock()
	user := u.users[destination]
	u.usersLock.Unlock()
	if user == nil {
		return
	}

	switch payment.Type {
	case "create_account":
		user.AccountCreated <- true
	case "payment":
		user.PaymentReceived <- true
	}
}

// newUser generates a new user interaction.
func (u *Users) newUser(kp *keypair.Full) server.GenerateAddressResponse {
	u.usersLock.Lock()
	user := u.users[kp.Address()]
	u.usersLock.Unlock()

	randomPort := u.BifrostPorts[rand.Int()%len(u.BifrostPorts)]
	randomCoin := []string{"bitcoin", "ethereum"}[rand.Int()%2]

	params := url.Values{}
	params.Add("hcnet_public_key", kp.Address())
	req, err := http.PostForm(
		fmt.Sprintf("http://localhost:%d/generate-%s-address", randomPort, randomCoin),
		params,
	)
	if err != nil {
		panic(err)
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var response server.GenerateAddressResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	u.updateUserState(kp.Address(), GeneratedAddressUserState)

	// Return generated address and continue interactions asynchronously
	go func() {
		// Wait for account to be created.
		<-user.AccountCreated

		account, exists, err := u.getAccount(kp.Address())
		if err != nil {
			panic(err)
		}

		if !exists {
			panic("disappeared")
		}

		u.updateUserState(kp.Address(), AccountCreatedUserState)

		sequence, err := strconv.ParseUint(account.Sequence, 10, 64)
		if err != nil {
			panic(err)
		}

		// Create trust lines
		sequence++
		tx, err := build.Transaction(
			build.SourceAccount{kp.Address()},
			build.Sequence{sequence},
			build.Network{u.NetworkPassphrase},
			build.Trust("BTC", u.IssuerPublicKey),
			build.Trust("ETH", u.IssuerPublicKey),
		)
		if err != nil {
			panic(err)
		}

		txe, err := tx.Sign(kp.Seed())
		if err != nil {
			panic(err)
		}

		txeB64, err := txe.Base64()
		if err != nil {
			panic(err)
		}

		_, err = u.Aurora.SubmitTransaction(txeB64)
		if err != nil {
			fmt.Println(txeB64)
			panic(err)
		}

		u.updateUserState(kp.Address(), TrustLinesCreatedUserState)

		// Wait for (BTC/ETH) payment
		<-user.PaymentReceived

		var assetCode, assetBalance string
		account, exists, err = u.getAccount(kp.Address())
		if err != nil {
			panic(err)
		}

		if !exists {
			panic("disappeared")
		}

		btcBalance := account.GetCreditBalance("BTC", u.IssuerPublicKey)
		ethBalance := account.GetCreditBalance("ETH", u.IssuerPublicKey)

		btcBalanceRat, ok := new(big.Rat).SetString(btcBalance)
		if !ok {
			panic("Error BTC balance: " + btcBalance)
		}
		ethBalanceRat, ok := new(big.Rat).SetString(ethBalance)
		if !ok {
			panic("Error ETH balance: " + ethBalance)
		}

		if btcBalanceRat.Sign() != 0 {
			assetCode = "BTC"
			assetBalance = btcBalance
		} else if ethBalanceRat.Sign() != 0 {
			assetCode = "ETH"
			assetBalance = ethBalance
		}

		u.updateUserState(kp.Address(), ReceivedPaymentUserState)

		// Merge account so we don't need to fund issuing account over and over again.
		sequence++
		tx, err = build.Transaction(
			build.SourceAccount{kp.Address()},
			build.Sequence{sequence},
			build.Network{u.NetworkPassphrase},
			build.Payment(
				build.Destination{u.IssuerPublicKey},
				build.CreditAmount{
					Code:   assetCode,
					Issuer: u.IssuerPublicKey,
					Amount: assetBalance,
				},
			),
			build.RemoveTrust("BTC", u.IssuerPublicKey),
			build.RemoveTrust("ETH", u.IssuerPublicKey),
			build.AccountMerge(
				build.Destination{u.IssuerPublicKey},
			),
		)
		if err != nil {
			panic(err)
		}

		txe, err = tx.Sign(kp.Seed())
		if err != nil {
			panic(err)
		}

		txeB64, err = txe.Base64()
		if err != nil {
			panic(err)
		}

		_, err = u.Aurora.SubmitTransaction(txeB64)
		if err != nil {
			if herr, ok := err.(*aurora.Error); ok {
				fmt.Println(herr.Problem)
			}
			panic(err)
		}
	}()

	return response
}

func (u *Users) printStatus() {
	u.usersLock.Lock()
	defer u.usersLock.Unlock()

	counters := map[UserState]int{}
	total := 0
	for _, user := range u.users {
		counters[user.State]++
		total++
	}

	u.log.Info(
		fmt.Sprintf(
			"Stress test status: total: %d, pending: %d, generated_address: %d, account_created: %d, trust_lines_created: %d, received_payment: %d",
			total,
			counters[PendingUserState],
			counters[GeneratedAddressUserState],
			counters[AccountCreatedUserState],
			counters[TrustLinesCreatedUserState],
			counters[ReceivedPaymentUserState],
		),
	)
}

func (u *Users) updateUserState(publicKey string, state UserState) {
	u.usersLock.Lock()
	defer u.usersLock.Unlock()
	user := u.users[publicKey]
	user.State = state
}

func (u *Users) getAccount(account string) (aurora.Account, bool, error) {
	var hAccount aurora.Account
	hAccount, err := u.Aurora.LoadAccount(account)
	if err != nil {
		if err2, ok := err.(*aurora.Error); ok && err2.Response.StatusCode == http.StatusNotFound {
			return hAccount, false, nil
		}
		return hAccount, false, err
	}

	return hAccount, true, nil
}
