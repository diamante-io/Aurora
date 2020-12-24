package stress

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"time"

	ethereumCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/diamnet/go/services/bifrost/common"
	"github.com/diamnet/go/support/log"
)

func (c *RandomEthereumClient) Start(addresses <-chan string) {
	c.log = common.CreateLogger("RandomEthereumClient")
	rand.Seed(time.Now().Unix())
	c.currentBlockNumber = 0
	c.blocks = map[int64]*types.Block{}
	c.firstBlockGenerated = make(chan bool)
	go c.generateBlocks()
	go c.addUserAddresses(addresses)
	go c.logStats()
	c.log.Info("Waiting for the first block...")
	<-c.firstBlockGenerated
}

func (g *RandomEthereumClient) logStats() {
	for {
		g.log.WithField("addresses", len(g.userAddresses)).Info("Stats")
		time.Sleep(15 * time.Second)
	}
}

func (g *RandomEthereumClient) addUserAddresses(addresses <-chan string) {
	for {
		address := <-addresses
		g.userAddressesLock.Lock()
		g.userAddresses = append(g.userAddresses, address)
		g.userAddressesLock.Unlock()
	}
}

// This should always return testnet ID so production ethereum.Listener
// genesis block check will fail.
func (c *RandomEthereumClient) NetworkID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(3), nil
}

func (c *RandomEthereumClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	if number == nil {
		number = big.NewInt(c.currentBlockNumber)
	}

	block, ok := c.blocks[number.Int64()]
	if !ok {
		return nil, errors.New("not found")
	}
	return block, nil
}

func (g *RandomEthereumClient) generateBlocks() {
	for {
		// Generate 50-200 txs
		transactionsCount := 50 + rand.Int()%150
		transactions := make([]*types.Transaction, 0, transactionsCount)
		for i := 0; i < transactionsCount; i++ {
			tx := types.NewTransaction(
				uint64(i),
				g.randomAddress(),
				g.randomAmount(),
				big.NewInt(1),
				big.NewInt(2),
				[]byte{0, 0, 0, 0},
			)
			transactions = append(transactions, tx)
		}

		newBlockNumber := g.currentBlockNumber + 1

		header := &types.Header{
			Number: big.NewInt(newBlockNumber),
			Time:   big.NewInt(time.Now().Unix()),
		}

		g.blocks[newBlockNumber] = types.NewBlock(header, transactions, []*types.Header{}, []*types.Receipt{})
		g.currentBlockNumber = newBlockNumber

		if g.currentBlockNumber == 1 {
			g.firstBlockGenerated <- true
		}

		g.log.WithFields(log.F{"blockNumber": newBlockNumber, "txs": transactionsCount}).Info("Generated block")
		time.Sleep(10 * time.Second)
	}
}

func (g *RandomEthereumClient) randomAddress() ethereumCommon.Address {
	g.userAddressesLock.Lock()
	defer g.userAddressesLock.Unlock()

	var address ethereumCommon.Address
	if len(g.userAddresses) > 0 {
		address = ethereumCommon.HexToAddress(g.userAddresses[0])
		g.userAddresses = g.userAddresses[1:]
	} else {
		rand.Read(address[:])
	}
	return address
}

// randomAmount generates random amount between [1, 50) ETH
func (g *RandomEthereumClient) randomAmount() *big.Int {
	eth := big.NewInt(1 + rand.Int63n(49))
	amount := new(big.Int)
	amount.Mul(eth, weiInEth)
	return amount
}
