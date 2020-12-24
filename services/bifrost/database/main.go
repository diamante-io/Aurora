package database

import (
	"time"

	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
)

type Chain string

// Scan implements database/sql.Scanner interface
func (s *Chain) Scan(src interface{}) error {
	value, ok := src.([]byte)
	if !ok {
		return errors.New("Cannot convert value to Chain")
	}
	*s = Chain(value)
	return nil
}

const (
	SchemaVersion uint32 = 2

	ChainBitcoin  Chain = "bitcoin"
	ChainEthereum Chain = "ethereum"
)

type Database interface {
	// CreateAddressAssociation creates Bitcoin/Ethereum-DiamNet association. `addressIndex`
	// is the chain (Bitcoin/Ethereum) address derivation index (BIP-32).
	CreateAddressAssociation(chain Chain, diamnetAddress, address string, addressIndex uint32) error
	// GetAssociationByChainAddress searches for previously saved Bitcoin/Ethereum-DiamNet association.
	// Should return nil if not found.
	GetAssociationByChainAddress(chain Chain, address string) (*AddressAssociation, error)
	// GetAssociationByDiamNetPublicKey searches for previously saved Bitcoin/Ethereum-DiamNet association.
	// Should return nil if not found.
	GetAssociationByDiamNetPublicKey(diamnetPublicKey string) (*AddressAssociation, error)
	// AddProcessedTransaction adds a transaction to database as processed. This
	// should return `true` and no error if transaction processing has already started/finished.
	AddProcessedTransaction(chain Chain, transactionID, receivingAddress string) (alreadyProcessing bool, err error)
	// IncrementAddressIndex returns the current value of index used for `chain` key
	// derivation and then increments it. This operation must be atomic so this function
	// should never return the same value more than once.
	IncrementAddressIndex(chain Chain) (uint32, error)

	// ResetBlockCounters changes last processed bitcoin and ethereum block to default value.
	// Used in stress tests.
	ResetBlockCounters() error

	// AddRecoveryTransaction inserts recovery account ID and transaction envelope
	AddRecoveryTransaction(sourceAccount string, txEnvelope string) error
}

type PostgresDatabase struct {
	session *db.Session
}

type AddressAssociation struct {
	// Chain is the name of the payment origin chain
	Chain Chain `db:"chain"`
	// BIP-44
	AddressIndex     uint32    `db:"address_index"`
	Address          string    `db:"address"`
	DiamNetPublicKey string    `db:"diamnet_public_key"`
	CreatedAt        time.Time `db:"created_at"`
}
