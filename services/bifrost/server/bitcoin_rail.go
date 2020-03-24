package server

import (
	"github.com/hcnet/go/services/bifrost/bitcoin"
	"github.com/hcnet/go/services/bifrost/database"
	"github.com/hcnet/go/services/bifrost/queue"
	"github.com/hcnet/go/services/bifrost/sse"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
)

// onNewBitcoinTransaction checks if transaction is valid and adds it to
// the transactions queue for HcNetAccountConfigurator to consume.
//
// Transaction added to transactions queue should be in a format described in
// queue.Transaction (especialy amounts). Pooling service should not have to deal with any
// conversions.
//
// This is very unlikely but it's possible that a single transaction will have more than
// one output going to bifrost account. Then only first output will be processed.
// Because it's very unlikely that this will happen and it's not a security issue this
// will be fixed in a future release.
func (s *Server) onNewBitcoinTransaction(transaction bitcoin.Transaction) error {
	localLog := s.log.WithFields(log.F{"transaction": transaction, "rail": "bitcoin"})
	localLog.Debug("Processing transaction")

	// Let's check if tx is valid first.

	// Check if value is above minimum required
	if transaction.ValueSat < s.minimumValueSat {
		localLog.Debug("Value is below minimum required amount, skipping")
		return nil
	}

	addressAssociation, err := s.Database.GetAssociationByChainAddress(database.ChainBitcoin, transaction.To)
	if err != nil {
		return errors.Wrap(err, "Error getting association")
	}

	if addressAssociation == nil {
		localLog.Debug("Associated address not found, skipping")
		return nil
	}

	// Add transaction as processing.
	processed, err := s.Database.AddProcessedTransaction(database.ChainBitcoin, transaction.Hash, transaction.To)
	if err != nil {
		return err
	}

	if processed {
		localLog.Debug("Transaction already processed, skipping")
		return nil
	}

	// Add tx to the processing queue
	queueTx := queue.Transaction{
		TransactionID: transaction.Hash,
		AssetCode:     queue.AssetCodeBTC,
		// Amount in the base unit of currency.
		Amount:           transaction.ValueToHcNet(),
		HcNetPublicKey: addressAssociation.HcNetPublicKey,
	}

	err = s.TransactionsQueue.QueueAdd(queueTx)
	if err != nil {
		return errors.Wrap(err, "Error adding transaction to the processing queue")
	}
	localLog.Info("Transaction added to transaction queue")

	// Broadcast event to address stream
	s.SSEServer.BroadcastEvent(transaction.To, sse.TransactionReceivedAddressEvent, nil)
	localLog.Info("Transaction processed successfully")
	return nil
}
