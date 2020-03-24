// Skip this test file in Go <1.8 because it's using http.Server.Shutdown
// +build go1.8

package server

import (
	"math/big"
	"testing"
	"time"

	"github.com/hcnet/go/services/bifrost/database"
	"github.com/hcnet/go/services/bifrost/ethereum"
	"github.com/hcnet/go/services/bifrost/queue"
	"github.com/hcnet/go/services/bifrost/sse"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var weiInEth = new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(18), nil)

type EthereumRailTestSuite struct {
	suite.Suite
	Server        *Server
	MockDatabase  *database.MockDatabase
	MockQueue     *queue.MockQueue
	MockSSEServer *sse.MockServer
}

func (suite *EthereumRailTestSuite) SetupTest() {
	suite.MockDatabase = &database.MockDatabase{}
	suite.MockQueue = &queue.MockQueue{}
	suite.MockSSEServer = &sse.MockServer{}

	suite.Server = &Server{
		Database:          suite.MockDatabase,
		TransactionsQueue: suite.MockQueue,
		SSEServer:         suite.MockSSEServer,
		minimumValueWei:   big.NewInt(1000000000000000000), // 1 ETH
	}
	suite.Server.initLogger()
}

func (suite *EthereumRailTestSuite) TearDownTest() {
	suite.MockDatabase.AssertExpectations(suite.T())
	suite.MockQueue.AssertExpectations(suite.T())
	suite.MockSSEServer.AssertExpectations(suite.T())
}

func (suite *EthereumRailTestSuite) TestInvalidValue() {
	transaction := ethereum.Transaction{
		Hash:     "0x0a190d17ba0405bce37fafd3a7a7bef51264ea4083ffae3b2de90ed61ee5264e",
		ValueWei: big.NewInt(500000000000000000), // 0.5 ETH
		To:       "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
	}
	suite.MockDatabase.AssertNotCalled(suite.T(), "AddProcessedTransaction")
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewEthereumTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *EthereumRailTestSuite) TestAssociationNotExist() {
	transaction := ethereum.Transaction{
		Hash:     "0x0a190d17ba0405bce37fafd3a7a7bef51264ea4083ffae3b2de90ed61ee5264e",
		ValueWei: weiInEth,
		To:       "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainEthereum, "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D").
		Return(nil, nil)
	suite.MockDatabase.AssertNotCalled(suite.T(), "AddProcessedTransaction")
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewEthereumTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *EthereumRailTestSuite) TestAssociationAlreadyProcessed() {
	transaction := ethereum.Transaction{
		Hash:     "0x0a190d17ba0405bce37fafd3a7a7bef51264ea4083ffae3b2de90ed61ee5264e",
		ValueWei: weiInEth,
		To:       "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
	}
	association := &database.AddressAssociation{
		Chain:            database.ChainEthereum,
		AddressIndex:     1,
		Address:          "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
		HcNetPublicKey: "GDULKYRRVOMASFMXBYD4BYFRSHAKQDREEVVP2TMH2CER3DW2KATIOASB",
		CreatedAt:        time.Now(),
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainEthereum, transaction.To).
		Return(association, nil)
	suite.MockDatabase.
		On("AddProcessedTransaction", database.ChainEthereum, transaction.Hash, transaction.To).
		Return(true, nil)
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewEthereumTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *EthereumRailTestSuite) TestAssociationSuccess() {
	transaction := ethereum.Transaction{
		Hash:     "0x0a190d17ba0405bce37fafd3a7a7bef51264ea4083ffae3b2de90ed61ee5264e",
		ValueWei: weiInEth,
		To:       "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
	}
	association := &database.AddressAssociation{
		Chain:            database.ChainEthereum,
		AddressIndex:     1,
		Address:          "0x80D3ee1268DC1A2d1b9E73D49050083E75Ef7c2D",
		HcNetPublicKey: "GDULKYRRVOMASFMXBYD4BYFRSHAKQDREEVVP2TMH2CER3DW2KATIOASB",
		CreatedAt:        time.Now(),
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainEthereum, transaction.To).
		Return(association, nil)
	suite.MockDatabase.
		On("AddProcessedTransaction", database.ChainEthereum, transaction.Hash, transaction.To).
		Return(false, nil)
	suite.MockQueue.
		On("QueueAdd", mock.AnythingOfType("queue.Transaction")).
		Return(nil).
		Run(func(args mock.Arguments) {
			queueTransaction := args.Get(0).(queue.Transaction)
			suite.Assert().Equal(transaction.Hash, queueTransaction.TransactionID)
			suite.Assert().Equal("ETH", string(queue.AssetCodeETH))
			suite.Assert().Equal(queue.AssetCodeETH, queueTransaction.AssetCode)
			suite.Assert().Equal("1.0000000", queueTransaction.Amount)
			suite.Assert().Equal(association.HcNetPublicKey, queueTransaction.HcNetPublicKey)
		})
	suite.MockSSEServer.
		On("BroadcastEvent", transaction.To, sse.TransactionReceivedAddressEvent, []byte(nil))
	err := suite.Server.onNewEthereumTransaction(transaction)
	suite.Require().NoError(err)
}

func TestEthereumRailTestSuite(t *testing.T) {
	suite.Run(t, new(EthereumRailTestSuite))
}
