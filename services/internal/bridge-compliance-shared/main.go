package shared

import (
	"bytes"

	"github.com/diamnet/go/hash"
	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/txnbuild"
	"github.com/diamnet/go/xdr"
)

// BuildTransaction is used in compliance server. The sequence number in built transaction will be equal 0!
func BuildTransaction(accountID, networkPassphrase string, operation []txnbuild.Operation, memo txnbuild.Memo) (string, error) {

	// Sequence number is set to -1 here because txnbuild auto increments by 1.
	tx := txnbuild.Transaction{
		SourceAccount: &txnbuild.SimpleAccount{AccountID: accountID, Sequence: int64(-1)},
		Operations:    operation,
		Timebounds:    txnbuild.NewInfiniteTimeout(),
		Network:       networkPassphrase,
		Memo:          memo,
	}

	err := tx.Build()
	if err != nil {
		return "", errors.Wrap(err, "unable to build transaction")
	}

	err = tx.Sign()
	if err != nil {
		return "", errors.Wrap(err, "unable to build transaction")
	}

	txeB64, err := tx.Base64()
	if err != nil {
		return "", errors.Wrap(err, "unable to encode transaction envelope")
	}

	var txXDR xdr.TransactionEnvelope
	err = xdr.SafeUnmarshalBase64(txeB64, &txXDR)
	if err != nil {
		return "", errors.Wrap(err, "unable to decode transaction envelope")
	}
	txB64, err := xdr.MarshalBase64(txXDR.Tx)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode transaction")
	}

	return txB64, err
}

// TransactionHash returns transaction hash for a given Transaction based on the network
func TransactionHash(tx *xdr.Transaction, networkPassphrase string) ([32]byte, error) {
	var txBytes bytes.Buffer

	h := hash.Hash([]byte(networkPassphrase))
	_, err := txBytes.Write(h[:])
	if err != nil {
		return [32]byte{}, err
	}

	_, err = xdr.Marshal(&txBytes, xdr.EnvelopeTypeEnvelopeTypeTx)
	if err != nil {
		return [32]byte{}, err
	}

	_, err = xdr.Marshal(&txBytes, tx)
	if err != nil {
		return [32]byte{}, err
	}

	return hash.Hash(txBytes.Bytes()), nil
}

// IsValidAccountID returns true if account ID is valid
func IsValidAccountID(accountID string) bool {
	_, err := keypair.Parse(accountID)
	if err != nil {
		return false
	}

	if accountID[0] != 'G' {
		return false
	}

	return true
}

// IsValidSecret returns true if secret is valid
func IsValidSecret(secret string) bool {
	_, err := keypair.Parse(secret)
	if err != nil {
		return false
	}

	if secret[0] != 'S' {
		return false
	}

	return true
}

// IsValidAssetCode returns true if asset code is valid
func IsValidAssetCode(code string) bool {
	if len(code) < 1 || len(code) > 12 {
		return false
	}
	return true
}
