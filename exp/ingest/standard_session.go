package ingest

func (s *standardSession) setRunningState(newState bool) {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.running && newState {
		panic("Session is running...")
	}
	s.running = newState
}

func (s *standardSession) Shutdown() {
	close(s.shutdown)
}

func (s *standardSession) QueryLock() {
	s.rwLock.RLock()
}

func (s *standardSession) QueryUnlock() {
	s.rwLock.RUnlock()
}

func (s *standardSession) UpdateLock() {
	s.rwLock.Lock()
}

func (s *standardSession) UpdateUnlock() {
	s.rwLock.Unlock()
}

func (s *standardSession) GetLatestSuccessfullyProcessedLedger() uint32 {
	return s.latestSuccessfullyProcessedLedger
}
