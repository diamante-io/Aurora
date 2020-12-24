package stress

import (
	"errors"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/haltingstate/secp256k1-go"
	"github.com/diamnet/go/services/bifrost/common"
	"github.com/diamnet/go/support/log"
)

func (c *RandomBitcoinClient) Start(addresses <-chan string) {
	c.log = common.CreateLogger("RandomBitcoinClient")
	rand.Seed(time.Now().Unix())
	c.currentBlockNumber = 0
	c.heightHash = map[int64]*chainhash.Hash{}
	c.hashBlock = map[*chainhash.Hash]*wire.MsgBlock{}
	c.firstBlockGenerated = make(chan bool)
	go c.generateBlocks()
	go c.addUserAddresses(addresses)
	go c.logStats()
	<-c.firstBlockGenerated
}

func (g *RandomBitcoinClient) logStats() {
	for {
		g.log.WithField("addresses", len(g.userAddresses)).Info("Stats")
		time.Sleep(15 * time.Second)
	}
}

func (g *RandomBitcoinClient) addUserAddresses(addresses <-chan string) {
	for {
		address := <-addresses
		g.userAddressesLock.Lock()
		g.userAddresses = append(g.userAddresses, address)
		g.userAddressesLock.Unlock()
	}
}

func (c *RandomBitcoinClient) GetBlockCount() (int64, error) {
	return c.currentBlockNumber, nil
}

// This should always return testnet genesis block hash for `0` so production
// bitcoin.Listener genesis block check will fail.
func (c *RandomBitcoinClient) GetBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	if blockHeight == 0 {
		return chaincfg.TestNet3Params.GenesisHash, nil
	}

	if blockHeight > c.currentBlockNumber {
		return nil, errors.New("Block height out of range")
	}
	return c.heightHash[blockHeight], nil
}

func (c *RandomBitcoinClient) GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error) {
	block, ok := c.hashBlock[blockHash]
	if !ok {
		return nil, errors.New("Block cannot be found")
	}
	return block, nil
}

func (c *RandomBitcoinClient) generateBlocks() {
	for {
		block := &wire.MsgBlock{
			Header: wire.BlockHeader{
				// We just want hashes to be different
				Version:   int32(c.currentBlockNumber),
				Timestamp: time.Now(),
			},
		}

		// Generate 50-200 txs
		transactionsCount := 50 + rand.Int()%150
		for i := 0; i < transactionsCount; i++ {
			pkscript, err := txscript.PayToAddrScript(c.randomAddress())
			if err != nil {
				panic(err)
			}

			tx := &wire.MsgTx{
				TxOut: []*wire.TxOut{
					{
						Value:    c.randomAmount(),
						PkScript: pkscript,
					},
				},
			}
			block.AddTransaction(tx)
		}

		nextBlockNumber := c.currentBlockNumber + 1

		blockHash := block.BlockHash()
		c.hashBlock[&blockHash] = block
		c.heightHash[nextBlockNumber] = &blockHash

		c.currentBlockNumber = nextBlockNumber

		if c.currentBlockNumber == 1 {
			c.firstBlockGenerated <- true
		}

		c.log.WithFields(log.F{"blockNumber": nextBlockNumber, "txs": transactionsCount}).Info("Generated block")
		// Stress tests, we want results faster than 1 block / 10 minutes.
		time.Sleep(10 * time.Second)
	}
}

func (g *RandomBitcoinClient) randomAddress() btcutil.Address {
	g.userAddressesLock.Lock()
	defer g.userAddressesLock.Unlock()

	var err error
	var address btcutil.Address

	if len(g.userAddresses) > 0 {
		address, err = btcutil.DecodeAddress(g.userAddresses[0], &chaincfg.TestNet3Params)
		if err != nil {
			panic(err)
		}
		g.userAddresses = g.userAddresses[1:]
	} else {
		pubKey, _ := secp256k1.GenerateKeyPair()
		address, err = btcutil.NewAddressPubKey(pubKey, &chaincfg.TestNet3Params)
		if err != nil {
			panic(err)
		}
	}

	return address
}

// randomAmount generates random amount between [0, 100) BTC
func (g *RandomBitcoinClient) randomAmount() int64 {
	return rand.Int63n(100 * satsInBtc)
}
