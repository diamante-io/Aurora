package hcnet

import (
	"strconv"

	"github.com/hcnet/go/build"
	"github.com/hcnet/go/clients/aurora"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
)

func (ac *AccountConfigurator) createAccountTransaction(destination string) error {
	transaction, err := ac.buildTransaction(
		ac.signerPublicKey,
		ac.SignerSecretKey,
		build.CreateAccount(
			build.SourceAccount{ac.DistributionPublicKey},
			build.Destination{destination},
			build.NativeAmount{ac.StartingBalance},
		),
	)
	if err != nil {
		return errors.Wrap(err, "Error building transaction")
	}

	err = ac.submitTransaction(transaction)
	if err != nil {
		return errors.Wrap(err, "Error submitting a transaction")
	}

	return nil
}

// configureAccountTransaction is using a signer on an user accounts to configure the account.
func (ac *AccountConfigurator) configureAccountTransaction(destination, intermediateAssetCode, amount string, needsAuthorize bool) error {
	mutators := []build.TransactionMutator{
		build.Trust(intermediateAssetCode, ac.IssuerPublicKey),
		build.Trust(ac.TokenAssetCode, ac.IssuerPublicKey),
	}

	if needsAuthorize {
		mutators = append(
			mutators,
			// Chain token received (BTC/ETH)
			build.AllowTrust(
				build.SourceAccount{ac.IssuerPublicKey},
				build.Trustor{destination},
				build.AllowTrustAsset{intermediateAssetCode},
				build.Authorize{true},
			),
			// Destination token
			build.AllowTrust(
				build.SourceAccount{ac.IssuerPublicKey},
				build.Trustor{destination},
				build.AllowTrustAsset{ac.TokenAssetCode},
				build.Authorize{true},
			),
		)
	}

	var tokenPrice string
	switch intermediateAssetCode {
	case "BTC":
		tokenPrice = ac.TokenPriceBTC
	case "ETH":
		tokenPrice = ac.TokenPriceETH
	default:
		return errors.Errorf("Invalid intermediateAssetCode: $%s", intermediateAssetCode)
	}

	mutators = append(
		mutators,
		// Send BTC/ETH
		build.Payment(
			build.SourceAccount{ac.DistributionPublicKey},
			build.Destination{destination},
			build.CreditAmount{
				Code:   intermediateAssetCode,
				Issuer: ac.IssuerPublicKey,
				Amount: amount,
			},
		),
		// Exchange BTC/ETH => token
		build.CreateOffer(
			build.Rate{
				Selling: build.CreditAsset(intermediateAssetCode, ac.IssuerPublicKey),
				Buying:  build.CreditAsset(ac.TokenAssetCode, ac.IssuerPublicKey),
				Price:   build.Price(tokenPrice),
			},
			build.Amount(amount),
		),
	)

	transaction, err := ac.buildTransaction(destination, ac.SignerSecretKey, mutators...)
	if err != nil {
		return errors.Wrap(err, "Error building a transaction")
	}

	err = ac.submitTransaction(transaction)
	if err != nil {
		return errors.Wrap(err, "Error submitting a transaction")
	}

	return nil
}

// removeTemporarySigner is removing temporary signer from an account.
func (ac *AccountConfigurator) removeTemporarySigner(destination string) error {
	// Remove signer
	mutators := []build.TransactionMutator{
		build.SetOptions(
			build.MasterWeight(1),
			build.RemoveSigner(ac.signerPublicKey),
		),
	}

	transaction, err := ac.buildTransaction(destination, ac.SignerSecretKey, mutators...)
	if err != nil {
		return errors.Wrap(err, "Error building a transaction")
	}

	err = ac.submitTransaction(transaction)
	if err != nil {
		return errors.Wrap(err, "Error submitting a transaction")
	}

	return nil
}

// buildUnlockAccountTransaction creates and returns unlock account transaction.
func (ac *AccountConfigurator) buildUnlockAccountTransaction(source string) (string, error) {
	// Remove signer
	mutators := []build.TransactionMutator{
		build.Timebounds{
			MinTime: ac.LockUnixTimestamp,
		},
		build.SetOptions(
			build.MasterWeight(1),
			build.RemoveSigner(ac.signerPublicKey),
		),
	}

	return ac.buildTransaction(source, ac.SignerSecretKey, mutators...)
}

func (ac *AccountConfigurator) buildTransaction(source string, signer string, mutators ...build.TransactionMutator) (string, error) {
	muts := []build.TransactionMutator{
		build.SourceAccount{source},
		build.Network{ac.NetworkPassphrase},
	}

	if source == ac.signerPublicKey {
		muts = append(muts, build.Sequence{ac.getSignerSequence()})
	} else {
		muts = append(muts, build.AutoSequence{ac.Aurora})
	}

	muts = append(muts, mutators...)
	tx, err := build.Transaction(muts...)
	if err != nil {
		return "", err
	}
	txe, err := tx.Sign(signer)
	if err != nil {
		return "", err
	}
	return txe.Base64()
}

func (ac *AccountConfigurator) submitTransaction(transaction string) error {
	localLog := log.WithField("tx", transaction)
	localLog.Info("Submitting transaction")

	_, err := ac.Aurora.SubmitTransaction(transaction)
	if err != nil {
		fields := log.F{"err": err}
		if err2, ok := err.(*aurora.Error); ok {
			fields["result"] = string(err2.Problem.Extras["result_xdr"])
			ac.updateSignerSequence()
		}
		localLog.WithFields(fields).Error("Error submitting transaction")
		return errors.Wrap(err, "Error submitting transaction")
	}

	localLog.Info("Transaction successfully submitted")
	return nil
}

func (ac *AccountConfigurator) updateSignerSequence() error {
	ac.signerSequenceMutex.Lock()
	defer ac.signerSequenceMutex.Unlock()

	account, err := ac.Aurora.LoadAccount(ac.signerPublicKey)
	if err != nil {
		err = errors.Wrap(err, "Error loading issuing account")
		ac.log.Error(err)
		return err
	}

	ac.signerSequence, err = strconv.ParseUint(account.Sequence, 10, 64)
	if err != nil {
		err = errors.Wrap(err, "Invalid DistributionPublicKey sequence")
		ac.log.Error(err)
		return err
	}

	return nil
}

func (ac *AccountConfigurator) getSignerSequence() uint64 {
	ac.signerSequenceMutex.Lock()
	defer ac.signerSequenceMutex.Unlock()
	ac.signerSequence++
	sequence := ac.signerSequence
	return sequence
}
