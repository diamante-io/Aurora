package bitcoin

import (
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/log"
	"github.com/tyler-smith/go-bip32"
)

var (
	eight = big.NewInt(8)
	ten   = big.NewInt(10)
	// satInBtc = 10^8
	satInBtc = new(big.Rat).SetInt(new(big.Int).Exp(ten, eight, nil))
)

// Listener listens for transactions using bitcoin-core RPC. It calls TransactionHandler for each new
// transactions. It will reprocess the block if TransactionHandler returns error. It will
// start from the block number returned from Storage.GetBitcoinBlockToProcess or the latest block
// if it returned 0. Transactions can be processed more than once, it's TransactionHandler
// responsibility to ignore duplicates.
// Listener tracks only P2PKH payments.
// You can run multiple Listeners if Storage is implemented correctly.
type Listener struct {
	Enabled            bool
	Client             Client  `inject:""`
	Storage            Storage `inject:""`
	TransactionHandler TransactionHandler
	Testnet            bool

	chainParams *chaincfg.Params
	log         *log.Entry
}

type Client interface {
	GetBlockCount() (int64, error)
	GetBlockHash(blockHeight int64) (*chainhash.Hash, error)
	GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error)
}

// Storage is an interface that must be implemented by an object using
// persistent storage.
type Storage interface {
	// GetBitcoinBlockToProcess gets the number of Bitcoin block to process. `0` means the
	// processing should start from the current block.
	GetBitcoinBlockToProcess() (uint64, error)
	// SaveLastProcessedBitcoinBlock should update the number of the last processed Bitcoin
	// block. It should only update the block if block > current block in atomic transaction.
	SaveLastProcessedBitcoinBlock(block uint64) error
}

type TransactionHandler func(transaction Transaction) error

type Transaction struct {
	Hash       string
	TxOutIndex int
	// Value in sats
	ValueSat int64
	To       string
}

type AddressGenerator struct {
	masterPublicKey *bip32.Key
	chainParams     *chaincfg.Params
}

func BtcToSat(btc string) (int64, error) {
	valueRat := new(big.Rat)
	_, ok := valueRat.SetString(btc)
	if !ok {
		return 0, errors.New("Could not convert to *big.Rat")
	}

	// Calculate value in satoshi
	valueRat.Mul(valueRat, satInBtc)

	// Ensure denominator is equal `1`
	if valueRat.Denom().Cmp(big.NewInt(1)) != 0 {
		return 0, errors.New("Invalid precision, is value smaller than 1 satoshi?")
	}

	return valueRat.Num().Int64(), nil
}
