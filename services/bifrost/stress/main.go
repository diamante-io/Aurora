package stress

import (
	"math/big"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/diamnet/go/clients/aurora"
	"github.com/diamnet/go/support/log"
)

const satsInBtc = 100000000

var (
	ten      = big.NewInt(10)
	eighteen = big.NewInt(18)
	// weiInEth = 10^18
	weiInEth = new(big.Int).Exp(ten, eighteen, nil)
)

// RandomBitcoinClient implements bitcoin.Client and generates random bitcoin transactions.
type RandomBitcoinClient struct {
	currentBlockNumber  int64
	heightHash          map[int64]*chainhash.Hash
	hashBlock           map[*chainhash.Hash]*wire.MsgBlock
	userAddresses       []string
	userAddressesLock   sync.Mutex
	firstBlockGenerated chan bool
	log                 *log.Entry
}

// RandomEthereumClient implements ethereum.Client and generates random ethereum transactions.
type RandomEthereumClient struct {
	currentBlockNumber  int64
	blocks              map[int64]*types.Block
	userAddresses       []string
	userAddressesLock   sync.Mutex
	firstBlockGenerated chan bool
	log                 *log.Entry
}

type UserState int

const (
	PendingUserState UserState = iota
	GeneratedAddressUserState
	AccountCreatedUserState
	TrustLinesCreatedUserState
	ReceivedPaymentUserState
)

// Users is responsible for imitating user interactions:
// * Request a new bifrost address by calling /generate-bitcoin-address or /generate-ethereum-address.
// * Add a generate address to RandomBitcoinClient or RandomEthereumClient so it generates a new transaction
//   and puts it in a future block.
// * Once account is funded, create a trustline.
// * Wait for BTC/ETH payment.
type Users struct {
	Aurora           aurora.ClientInterface
	NetworkPassphrase string
	UsersPerSecond    int
	BifrostPorts      []int
	IssuerPublicKey   string

	users     map[string]*User // public key => User
	usersLock sync.Mutex
	log       *log.Entry
}

type User struct {
	State           UserState
	AccountCreated  chan bool
	PaymentReceived chan bool
}
