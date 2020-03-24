package server

import (
	"encoding/json"

	"github.com/hcnet/go/services/bifrost/sse"
)

func (s *Server) onHcNetAccountCreated(destination string) {
	association, err := s.Database.GetAssociationByHcNetPublicKey(destination)
	if err != nil {
		s.log.WithField("err", err).Error("Error getting association")
		return
	}

	if association == nil {
		s.log.WithField("hcnetPublicKey", destination).Error("Association not found")
		return
	}

	s.SSEServer.BroadcastEvent(association.Address, sse.AccountCreatedAddressEvent, nil)
}

func (s *Server) onExchanged(destination string) {
	association, err := s.Database.GetAssociationByHcNetPublicKey(destination)
	if err != nil {
		s.log.WithField("err", err).Error("Error getting association")
		return
	}

	if association == nil {
		s.log.WithField("hcnetPublicKey", destination).Error("Association not found")
		return
	}

	s.SSEServer.BroadcastEvent(association.Address, sse.ExchangedEvent, nil)
}

func (s *Server) OnExchangedTimelocked(destination, transaction string) {
	association, err := s.Database.GetAssociationByHcNetPublicKey(destination)
	if err != nil {
		s.log.WithField("err", err).Error("Error getting association")
		return
	}

	if association == nil {
		s.log.WithField("hcnetPublicKey", destination).Error("Association not found")
		return
	}

	// Save tx to database
	err = s.Database.AddRecoveryTransaction(destination, transaction)
	if err != nil {
		s.log.WithField("err", err).Error("Error saving unlock transaction to DB")
		return
	}

	data := map[string]string{
		"transaction": transaction,
	}

	j, err := json.Marshal(data)
	if err != nil {
		s.log.WithField("data", data).Error("Error marshalling json")
	}

	s.SSEServer.BroadcastEvent(association.Address, sse.ExchangedTimelockedEvent, j)
}
