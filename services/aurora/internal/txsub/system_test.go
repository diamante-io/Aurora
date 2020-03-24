package txsub

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hcnet/go/build"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/services/aurora/internal/txsub/sequence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SystemTestSuite struct {
	suite.Suite
	ctx       context.Context
	submitter *MockSubmitter
	results   *MockResultProvider
	sequences *MockSequenceProvider
	system    *System
	noResults Result
	successTx Result
	badSeq    SubmissionResult
}

func (suite *SystemTestSuite) SetupTest() {
	suite.ctx = test.Context()
	suite.submitter = &MockSubmitter{}
	suite.results = &MockResultProvider{}
	suite.sequences = &MockSequenceProvider{}

	suite.system = &System{
		Pending:           NewDefaultSubmissionList(),
		Submitter:         suite.submitter,
		Results:           suite.results,
		Sequences:         suite.sequences,
		SubmissionQueue:   sequence.NewManager(),
		NetworkPassphrase: build.TestNetwork.Passphrase,
	}

	suite.noResults = Result{Err: ErrNoResults}
	suite.successTx = Result{
		Hash:           "2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d",
		LedgerSequence: 2,
		EnvelopeXDR:    "AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAArqN6LeOagjxMaUP96Bzfs9e0corNZXzBWJkFoK7kvkwAAAAAO5rKAAAAAAAAAAABVvwF9wAAAECDzqvkQBQoNAJifPRXDoLhvtycT3lFPCQ51gkdsFHaBNWw05S/VhW0Xgkr0CBPE4NaFV2Kmcs3ZwLmib4TRrML",
		ResultXDR:      "I3Tpk0m57326ml2zM5t4/ajzR3exrzO6RorVwN+UbU0AAAAAAAAAZAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAA==",
	}

	suite.badSeq = SubmissionResult{
		Err: ErrBadSequence,
	}

	suite.sequences.On("Get", []string{"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"}).
		Return(map[string]uint64{"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H": 0}, nil).
		Once()
}

// Returns the result provided by the ResultProvider.
func (suite *SystemTestSuite) TestSubmit_Basic() {
	suite.results.Results = []Result{suite.successTx}
	r := <-suite.system.Submit(suite.ctx, suite.successTx.EnvelopeXDR)

	assert.Nil(suite.T(), r.Err)
	assert.Equal(suite.T(), suite.successTx.Hash, r.Hash)
	assert.False(suite.T(), suite.submitter.WasSubmittedTo)
}

// Returns the error from submission if no result is found by hash and the suite.submitter returns an error.
func (suite *SystemTestSuite) TestSubmit_NotFoundError() {
	suite.submitter.R.Err = errors.New("busted for some reason")
	r := <-suite.system.Submit(suite.ctx, suite.successTx.EnvelopeXDR)

	assert.NotNil(suite.T(), r.Err)
	assert.True(suite.T(), suite.submitter.WasSubmittedTo)
	assert.Equal(suite.T(), int64(0), suite.system.Metrics.SuccessfulSubmissionsMeter.Count())
	assert.Equal(suite.T(), int64(1), suite.system.Metrics.FailedSubmissionsMeter.Count())
	assert.Equal(suite.T(), int64(1), suite.system.Metrics.SubmissionTimer.Count())
}

// If the error is bad_seq and the result at the transaction's sequence number is for the same hash, return result.
func (suite *SystemTestSuite) TestSubmit_BadSeq() {
	suite.submitter.R = suite.badSeq
	suite.results.Results = []Result{suite.noResults, suite.successTx}

	r := <-suite.system.Submit(suite.ctx, suite.successTx.EnvelopeXDR)

	assert.Nil(suite.T(), r.Err)
	assert.Equal(suite.T(), suite.successTx.Hash, r.Hash)
	assert.True(suite.T(), suite.submitter.WasSubmittedTo)
}

// If error is bad_seq and no result is found, return error.
func (suite *SystemTestSuite) TestSubmit_BadSeqNotFound() {
	suite.submitter.R = suite.badSeq
	suite.submitter.R = suite.badSeq
	r := <-suite.system.Submit(suite.ctx, suite.successTx.EnvelopeXDR)

	assert.NotNil(suite.T(), r.Err)
	assert.True(suite.T(), suite.submitter.WasSubmittedTo)
}

// If no result found and no error submitting, add to open transaction list.
func (suite *SystemTestSuite) TestSubmit_OpenTransactionList() {
	_ = suite.system.Submit(suite.ctx, suite.successTx.EnvelopeXDR)
	pending := suite.system.Pending.Pending(suite.ctx)
	assert.Equal(suite.T(), 1, len(pending))
	assert.Equal(suite.T(), suite.successTx.Hash, pending[0])
	assert.Equal(suite.T(), int64(1), suite.system.Metrics.SuccessfulSubmissionsMeter.Count())
	assert.Equal(suite.T(), int64(0), suite.system.Metrics.FailedSubmissionsMeter.Count())
	assert.Equal(suite.T(), int64(1), suite.system.Metrics.SubmissionTimer.Count())
}

// Tick should be a no-op if there are no open submissions.
func (suite *SystemTestSuite) TestTick_Noop() {
	suite.system.Tick(suite.ctx)
}

// TestTick_Deadlock is a regression test for Tick() deadlock: if for any reason
// call to Tick() takes more time and another Tick() is called.
// This test starts two go routines: both calling Tick() but the call to
// `sys.Sequences.Get(addys)` is delayed by 1 second. It allows to simulate two
// calls to `Tick()` executed at the same time.
func (suite *SystemTestSuite) TestTick_Deadlock() {
	secondDone := make(chan bool, 1)
	testDone := make(chan bool)

	go func() {
		select {
		case <-secondDone:
			// OK!
		case <-time.After(5 * time.Second):
			assert.Fail(suite.T(), "Timeout, likely a deadlock in Tick()")
		}

		testDone <- true
	}()

	// Start first Tick
	suite.system.SubmissionQueue.Push("address", 0)
	// Configure suite.sequences to return after 1 second in a first call
	suite.sequences.On("Get", []string{"address"}).After(time.Second).Return(map[string]uint64{}, nil)

	go func() {
		fmt.Println("Starting first Tick()")
		suite.system.Tick(suite.ctx)
		fmt.Println("Finished first Tick()")
	}()

	go func() {
		// Start second Tick - should be deadlocked if mutex is not Unlock()'ed.
		fmt.Println("Starting second Tick()")
		suite.system.Tick(suite.ctx)
		fmt.Println("Finished second Tick()")
		secondDone <- true
	}()

	<-testDone
}

// Test that Tick finishes any available transactions,
func (suite *SystemTestSuite) TestTick_FinishesTransactions() {
	l := make(chan Result, 1)
	suite.system.Pending.Add(suite.ctx, suite.successTx.Hash, l)
	suite.system.Tick(suite.ctx)
	assert.Equal(suite.T(), 0, len(l))
	assert.Equal(suite.T(), 1, len(suite.system.Pending.Pending(suite.ctx)))

	suite.results.Results = []Result{suite.successTx}
	suite.system.Tick(suite.ctx)

	assert.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), 0, len(suite.system.Pending.Pending(suite.ctx)))
}

// Test that Tick removes old submissions that have timed out.
func (suite *SystemTestSuite) TestTick_RemovesStaleSubmissions() {
	l := make(chan Result, 1)
	suite.system.SubmissionTimeout = 100 * time.Millisecond
	suite.system.Pending.Add(suite.ctx, suite.successTx.Hash, l)
	<-time.After(101 * time.Millisecond)
	suite.system.Tick(suite.ctx)

	assert.Equal(suite.T(), 0, len(suite.system.Pending.Pending(suite.ctx)))
	assert.Equal(suite.T(), 1, len(l))
	<-l
	select {
	case _, stillOpen := <-l:
		assert.False(suite.T(), stillOpen)
	default:
		panic("could not read from listener")
	}
}

func TestSystemTestSuite(t *testing.T) {
	suite.Run(t, new(SystemTestSuite))
}
