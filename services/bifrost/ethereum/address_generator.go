package ethereum

import (
	ethereumCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/haltingstate/secp256k1-go"
	"github.com/hcnet/go/support/errors"
	"github.com/tyler-smith/go-bip32"
)

func NewAddressGenerator(masterPublicKeyString string) (*AddressGenerator, error) {
	deserializedMasterPublicKey, err := bip32.B58Deserialize(masterPublicKeyString)
	if err != nil {
		return nil, errors.Wrap(err, "Error deserializing master public key")
	}

	if deserializedMasterPublicKey.IsPrivate {
		return nil, errors.New("Key is not a master public key")
	}

	return &AddressGenerator{deserializedMasterPublicKey}, nil
}

func (g *AddressGenerator) Generate(index uint32) (string, error) {
	if g.masterPublicKey == nil {
		return "", errors.New("No master public key set")
	}

	accountKey, err := g.masterPublicKey.NewChildKey(index)
	if err != nil {
		return "", errors.Wrap(err, "Error creating new child key")
	}

	uncompressed := secp256k1.UncompressPubkey(accountKey.Key)
	uncompressed = uncompressed[1:]

	keccak := crypto.Keccak256(uncompressed)
	address := ethereumCommon.BytesToAddress(keccak[12:]).Hex() // Encode lower 160 bits/20 bytes
	return address, nil
}
