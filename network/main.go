// Package network contains functions that deal with hcnet network passphrases
// and IDs.
package network

import (
	"bytes"

	"strings"

	"github.com/hcnet/go/hash"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

const (
	// PublicNetworkPassphrase is the pass phrase used for every transaction intended for the public hcnet network
	PublicNetworkPassphrase = "Public Global HcNet Network ; September 2015"
	// TestNetworkPassphrase is the pass phrase used for every transaction intended for the SDF-run test network
	TestNetworkPassphrase = "Test SDF Network ; September 2015"
)

// ID returns the network ID derived from the provided passphrase.  This value
// also happens to be the raw (i.e. not strkey encoded) secret key for the root
// account of the network.
func ID(passphrase string) [32]byte {
	return hash.Hash([]byte(passphrase))
}

// HashTransaction derives the network specific hash for the provided
// transaction using the network identified by the supplied passphrase.  The
// resulting hash is the value that can be signed by hcnet secret key to
// authorize the transaction identified by the hash to hcnet validators.
func HashTransaction(tx *xdr.Transaction, passphrase string) ([32]byte, error) {
	var txBytes bytes.Buffer

	if strings.TrimSpace(passphrase) == "" {
		return [32]byte{}, errors.New("empty network passphrase")
	}

	id := ID(passphrase)
	_, err := txBytes.Write(id[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "fprint network id failed")
	}

	_, err = xdr.Marshal(&txBytes, xdr.EnvelopeTypeEnvelopeTypeTx)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "marshal type failed")
	}

	_, err = xdr.Marshal(&txBytes, tx)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "marshal tx failed")
	}

	return hash.Hash(txBytes.Bytes()), nil
}
