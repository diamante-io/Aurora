package ingest

import (
	"github.com/hcnet/go/exp/ingest/adapters"
	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/support/errors"
)

var _ Session = &SingleLedgerSession{}

func (s *SingleLedgerSession) Run() error {
	s.setRunningState(true)
	defer s.setRunningState(false)
	s.shutdown = make(chan bool)

	historyAdapter := adapters.MakeHistoryArchiveAdapter(s.Archive)

	var err error
	sequence := s.LedgerSequence
	if sequence == 0 {
		sequence, err = historyAdapter.GetLatestLedgerSequence()
		if err != nil {
			return errors.Wrap(err, "Error getting the latest ledger sequence")
		}
	}

	err = s.processState(historyAdapter, sequence)
	if err != nil {
		return errors.Wrap(err, "processState errored")
	}

	s.standardSession.latestSuccessfullyProcessedLedger = sequence
	return nil
}

func (s *SingleLedgerSession) Resume(ledgerSequence uint32) error {
	panic("Not possible to resume SingleLedgerSession")
}

func (s *SingleLedgerSession) processState(historyAdapter *adapters.HistoryArchiveAdapter, sequence uint32) error {
	var tempSet io.TempSet = &io.MemoryTempSet{}
	if s.TempSet != nil {
		tempSet = s.TempSet
	}

	stateReader, err := historyAdapter.GetState(sequence, tempSet)
	if err != nil {
		return errors.Wrap(err, "Error getting state from history archive")
	}
	if s.StateReporter != nil {
		s.StateReporter.OnStartState(sequence)
		stateReader = reporterStateReader{stateReader, s.StateReporter}
	}

	errChan := s.StatePipeline.Process(stateReader)
	select {
	case err := <-errChan:
		if err != nil {
			if s.StateReporter != nil {
				s.StateReporter.OnEndState(err, false)
			}
			return errors.Wrap(err, "State pipeline errored")
		}
	case <-s.shutdown:
		if s.StateReporter != nil {
			s.StateReporter.OnEndState(nil, true)
		}
		s.StatePipeline.Shutdown()
	}

	if s.StateReporter != nil {
		s.StateReporter.OnEndState(nil, false)
	}
	return nil
}
