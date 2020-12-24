// Skip this test file in Go <1.8 because it's using http.Server.Shutdown
// +build go1.8

package server

import (
	"testing"
	"time"

	"github.com/diamnet/go/services/bifrost/bitcoin"
	"github.com/diamnet/go/services/bifrost/database"
	"github.com/diamnet/go/services/bifrost/queue"
	"github.com/diamnet/go/services/bifrost/sse"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BitcoinRailTestSuite struct {
	suite.Suite
	Server        *Server
	MockDatabase  *database.MockDatabase
	MockQueue     *queue.MockQueue
	MockSSEServer *sse.MockServer
}

func (suite *BitcoinRailTestSuite) SetupTest() {
	suite.MockDatabase = &database.MockDatabase{}
	suite.MockQueue = &queue.MockQueue{}
	suite.MockSSEServer = &sse.MockServer{}

	suite.Server = &Server{
		Database:          suite.MockDatabase,
		TransactionsQueue: suite.MockQueue,
		SSEServer:         suite.MockSSEServer,
		minimumValueSat:   100000000, // 1 BTC
	}
	suite.Server.initLogger()
}

func (suite *BitcoinRailTestSuite) TearDownTest() {
	suite.MockDatabase.AssertExpectations(suite.T())
	suite.MockQueue.AssertExpectations(suite.T())
	suite.MockSSEServer.AssertExpectations(suite.T())
}

func (suite *BitcoinRailTestSuite) TestInvalidValue() {
	transaction := bitcoin.Transaction{
		Hash:       "109fa1c369680c2f27643fdd160620d010851a376d25b9b00ef71afe789ea6ed",
		TxOutIndex: 0,
		ValueSat:   50000000, // 0.5 BTC
		To:         "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
	}
	suite.MockDatabase.AssertNotCalled(suite.T(), "AddProcessedTransaction")
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewBitcoinTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *BitcoinRailTestSuite) TestAssociationNotExist() {
	transaction := bitcoin.Transaction{
		Hash:       "109fa1c369680c2f27643fdd160620d010851a376d25b9b00ef71afe789ea6ed",
		TxOutIndex: 0,
		ValueSat:   100000000,
		To:         "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainBitcoin, "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf").
		Return(nil, nil)
	suite.MockDatabase.AssertNotCalled(suite.T(), "AddProcessedTransaction")
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewBitcoinTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *BitcoinRailTestSuite) TestAssociationAlreadyProcessed() {
	transaction := bitcoin.Transaction{
		Hash:       "109fa1c369680c2f27643fdd160620d010851a376d25b9b00ef71afe789ea6ed",
		TxOutIndex: 0,
		ValueSat:   100000000,
		To:         "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
	}
	association := &database.AddressAssociation{
		Chain:            database.ChainBitcoin,
		AddressIndex:     1,
		Address:          "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
		DiamNetPublicKey: "GDULKYRRVOMASFMXBYD4BYFRSHAKQDREEVVP2TMH2CER3DW2KATIOASB",
		CreatedAt:        time.Now(),
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainBitcoin, transaction.To).
		Return(association, nil)
	suite.MockDatabase.
		On("AddProcessedTransaction", database.ChainBitcoin, transaction.Hash, transaction.To).
		Return(true, nil)
	suite.MockQueue.AssertNotCalled(suite.T(), "QueueAdd")
	err := suite.Server.onNewBitcoinTransaction(transaction)
	suite.Require().NoError(err)
}

func (suite *BitcoinRailTestSuite) TestAssociationSuccess() {
	transaction := bitcoin.Transaction{
		Hash:       "109fa1c369680c2f27643fdd160620d010851a376d25b9b00ef71afe789ea6ed",
		TxOutIndex: 0,
		ValueSat:   100000000,
		To:         "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
	}
	association := &database.AddressAssociation{
		Chain:            database.ChainBitcoin,
		AddressIndex:     1,
		Address:          "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf",
		DiamNetPublicKey: "GDULKYRRVOMASFMXBYD4BYFRSHAKQDREEVVP2TMH2CER3DW2KATIOASB",
		CreatedAt:        time.Now(),
	}
	suite.MockDatabase.
		On("GetAssociationByChainAddress", database.ChainBitcoin, transaction.To).
		Return(association, nil)
	suite.MockDatabase.
		On("AddProcessedTransaction", database.ChainBitcoin, transaction.Hash, transaction.To).
		Return(false, nil)
	suite.MockQueue.
		On("QueueAdd", mock.AnythingOfType("queue.Transaction")).
		Return(nil).
		Run(func(args mock.Arguments) {
			queueTransaction := args.Get(0).(queue.Transaction)
			suite.Assert().Equal(transaction.Hash, queueTransaction.TransactionID)
			suite.Assert().Equal("BTC", string(queue.AssetCodeBTC))
			suite.Assert().Equal(queue.AssetCodeBTC, queueTransaction.AssetCode)
			suite.Assert().Equal("1.0000000", queueTransaction.Amount)
			suite.Assert().Equal(association.DiamNetPublicKey, queueTransaction.DiamNetPublicKey)
		})
	suite.MockSSEServer.
		On("BroadcastEvent", transaction.To, sse.TransactionReceivedAddressEvent, []byte(nil))
	err := suite.Server.onNewBitcoinTransaction(transaction)
	suite.Require().NoError(err)
}

func TestBitcoinRailTestSuite(t *testing.T) {
	suite.Run(t, new(BitcoinRailTestSuite))
}
