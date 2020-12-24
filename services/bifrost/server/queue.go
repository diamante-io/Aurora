package server

import (
	"time"
)

// poolTransactionsQueue pools transactions queue which contains only processed and
// validated transactions and sends it to DiamNetAccountConfigurator for account configuration.
func (s *Server) poolTransactionsQueue() {
	s.log.Info("Started pooling transactions queue")

	for {
		transaction, err := s.TransactionsQueue.QueuePool()
		if err != nil {
			s.log.WithField("err", err).Error("Error pooling transactions queue")
			time.Sleep(time.Second)
			continue
		}

		if transaction == nil {
			time.Sleep(time.Second)
			continue
		}

		s.log.WithField("transaction", transaction).Info("Received transaction from transactions queue")
		go s.DiamNetAccountConfigurator.ConfigureAccount(
			transaction.DiamNetPublicKey,
			string(transaction.AssetCode),
			transaction.Amount,
		)
	}
}
