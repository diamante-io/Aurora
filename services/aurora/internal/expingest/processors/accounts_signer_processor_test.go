package processors

import (
	"context"
	stdio "io"
	"testing"

	"github.com/diamnet/go/exp/ingest/io"
	supportPipeline "github.com/diamnet/go/exp/support/pipeline"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/suite"
)

func TestAccountsSignerProcessorTestSuiteState(t *testing.T) {
	suite.Run(t, new(AccountsSignerProcessorTestSuiteState))
}

type AccountsSignerProcessorTestSuiteState struct {
	suite.Suite
	processor       *DatabaseProcessor
	mockQ           *history.MockQSigners
	mockStateReader *io.MockStateReader
	mockStateWriter *io.MockStateWriter
}

func (s *AccountsSignerProcessorTestSuiteState) SetupTest() {
	s.mockQ = &history.MockQSigners{}
	s.mockStateReader = &io.MockStateReader{}
	s.mockStateWriter = &io.MockStateWriter{}

	s.processor = &DatabaseProcessor{
		Action:   AccountsForSigner,
		HistoryQ: s.mockQ,
	}

	// Reader and Writer should be always closed and once
	s.mockStateReader.
		On("Close").
		Return(nil).Once()

	s.mockStateWriter.
		On("Close").
		Return(nil).Once()
}

func (s *AccountsSignerProcessorTestSuiteState) TearDownTest() {
	s.mockQ.AssertExpectations(s.T())
	s.mockStateReader.AssertExpectations(s.T())
	s.mockStateWriter.AssertExpectations(s.T())
}

func (s *AccountsSignerProcessorTestSuiteState) TestNoEntries() {
	s.mockStateReader.
		On("Read").
		Return(xdr.LedgerEntryChange{}, stdio.EOF).Once()

	err := s.processor.ProcessState(
		context.Background(),
		&supportPipeline.Store{},
		s.mockStateReader,
		s.mockStateWriter,
	)

	s.Assert().NoError(err)
}

func (s *AccountsSignerProcessorTestSuiteState) TestInvalidEntry() {
	s.mockStateReader.
		On("Read").
		Return(xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
		}, nil).Once()

	err := s.processor.ProcessState(
		context.Background(),
		&supportPipeline.Store{},
		s.mockStateReader,
		s.mockStateWriter,
	)

	s.Assert().EqualError(err, "DatabaseProcessor requires LedgerEntryChangeTypeLedgerEntryState changes only")
}

func (s *AccountsSignerProcessorTestSuiteState) TestCreatesSigners() {
	s.mockStateReader.
		On("Read").
		Return(xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
			State: &xdr.LedgerEntry{
				Data: xdr.LedgerEntryData{
					Type: xdr.LedgerEntryTypeAccount,
					Account: &xdr.AccountEntry{
						AccountId:  xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
						Thresholds: [4]byte{1, 1, 1, 1},
					},
				},
			},
		}, nil).Once()

	s.mockQ.
		On(
			"CreateAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			int32(1),
		).
		Return(nil).Once()

	s.mockStateReader.
		On("Read").
		Return(xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
			State: &xdr.LedgerEntry{
				Data: xdr.LedgerEntryData{
					Type: xdr.LedgerEntryTypeAccount,
					Account: &xdr.AccountEntry{
						AccountId: xdr.MustAddress("GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX"),
						Signers: []xdr.Signer{
							xdr.Signer{
								Key:    xdr.MustSigner("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
								Weight: 10,
							},
						},
					},
				},
			},
		}, nil).Once()

	s.mockQ.
		On(
			"CreateAccountSigner",
			"GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			int32(10),
		).
		Return(nil).Once()

	s.mockStateReader.
		On("Read").
		Return(xdr.LedgerEntryChange{}, stdio.EOF).Once()

	err := s.processor.ProcessState(
		context.Background(),
		&supportPipeline.Store{},
		s.mockStateReader,
		s.mockStateWriter,
	)

	s.Assert().NoError(err)
}

func TestAccountsSignerProcessorTestSuiteLedger(t *testing.T) {
	suite.Run(t, new(AccountsSignerProcessorTestSuiteLedger))
}

type AccountsSignerProcessorTestSuiteLedger struct {
	suite.Suite
	processor        *DatabaseProcessor
	mockQ            *history.MockQSigners
	mockLedgerReader *io.MockLedgerReader
	mockLedgerWriter *io.MockLedgerWriter
}

func (s *AccountsSignerProcessorTestSuiteLedger) SetupTest() {
	s.mockQ = &history.MockQSigners{}
	s.mockLedgerReader = &io.MockLedgerReader{}
	s.mockLedgerWriter = &io.MockLedgerWriter{}

	s.processor = &DatabaseProcessor{
		Action:   AccountsForSigner,
		HistoryQ: s.mockQ,
	}

	// Reader and Writer should be always closed and once
	s.mockLedgerReader.
		On("Close").
		Return(nil).Once()

	s.mockLedgerWriter.
		On("Close").
		Return(nil).Once()
}

func (s *AccountsSignerProcessorTestSuiteLedger) TearDownTest() {
	s.mockQ.AssertExpectations(s.T())
	s.mockLedgerReader.AssertExpectations(s.T())
	s.mockLedgerWriter.AssertExpectations(s.T())
}

func (s *AccountsSignerProcessorTestSuiteLedger) TestNoTransactions() {
	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	err := s.processor.ProcessLedger(
		context.Background(),
		&supportPipeline.Store{},
		s.mockLedgerReader,
		s.mockLedgerWriter,
	)

	s.Assert().NoError(err)
}

func (s *AccountsSignerProcessorTestSuiteLedger) TestNewAccount() {
	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
							Created: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId:  xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Thresholds: [4]byte{1, 1, 1, 1},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	s.mockQ.
		On(
			"CreateAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			int32(1),
		).
		Return(nil).Once()

	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	err := s.processor.ProcessLedger(
		context.Background(),
		&supportPipeline.Store{},
		s.mockLedgerReader,
		s.mockLedgerWriter,
	)

	s.Assert().NoError(err)
}

func (s *AccountsSignerProcessorTestSuiteLedger) TestNewSigner() {
	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						// State
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
							State: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Signers: []xdr.Signer{
											xdr.Signer{
												Key:    xdr.MustSigner("GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV"),
												Weight: 10,
											},
										},
									},
								},
							},
						},
						// Updated
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryUpdated,
							Updated: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Signers: []xdr.Signer{
											xdr.Signer{
												Key:    xdr.MustSigner("GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV"),
												Weight: 10,
											},
											xdr.Signer{
												Key:    xdr.MustSigner("GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS"),
												Weight: 15,
											},
										},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	// Remove old signer
	s.mockQ.
		On(
			"RemoveAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV",
		).
		Return(nil).Once()

	// Create new and old signer
	s.mockQ.
		On(
			"CreateAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV",
			int32(10),
		).
		Return(nil).Once()

	s.mockQ.
		On(
			"CreateAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS",
			int32(15),
		).
		Return(nil).Once()

	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	err := s.processor.ProcessLedger(
		context.Background(),
		&supportPipeline.Store{},
		s.mockLedgerReader,
		s.mockLedgerWriter,
	)

	s.Assert().NoError(err)
}

func (s *AccountsSignerProcessorTestSuiteLedger) TestSignerRemoved() {
	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						// State
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
							State: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Signers: []xdr.Signer{
											xdr.Signer{
												Key:    xdr.MustSigner("GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV"),
												Weight: 10,
											},
											xdr.Signer{
												Key:    xdr.MustSigner("GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS"),
												Weight: 15,
											},
										},
									},
								},
							},
						},
						// Updated
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryUpdated,
							Updated: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Signers: []xdr.Signer{
											xdr.Signer{
												Key:    xdr.MustSigner("GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS"),
												Weight: 15,
											},
										},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	// Remove old signers
	s.mockQ.
		On(
			"RemoveAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCBBDQLCTNASZJ3MTKAOYEOWRGSHDFAJVI7VPZUOP7KXNHYR3HP2BUKV",
		).
		Return(nil).Once()

	s.mockQ.
		On(
			"RemoveAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS",
		).
		Return(nil).Once()

	// Create new signer
	s.mockQ.
		On(
			"CreateAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GCAHY6JSXQFKWKP6R7U5JPXDVNV4DJWOWRFLY3Y6YPBF64QRL4BPFDNS",
			int32(15),
		).
		Return(nil).Once()

	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	err := s.processor.ProcessLedger(
		context.Background(),
		&supportPipeline.Store{},
		s.mockLedgerReader,
		s.mockLedgerWriter,
	)

	s.Assert().NoError(err)
}

func (s *AccountsSignerProcessorTestSuiteLedger) TestRemoveAccount() {
	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
							State: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeAccount,
									Account: &xdr.AccountEntry{
										AccountId:  xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
										Thresholds: [4]byte{1, 1, 1, 1},
									},
								},
							},
						},
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryRemoved,
							Removed: &xdr.LedgerKey{
								Type: xdr.LedgerEntryTypeAccount,
								Account: &xdr.LedgerKeyAccount{
									AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	s.mockQ.
		On(
			"RemoveAccountSigner",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
			"GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
		).
		Return(nil).Once()

	s.mockLedgerReader.
		On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	err := s.processor.ProcessLedger(
		context.Background(),
		&supportPipeline.Store{},
		s.mockLedgerReader,
		s.mockLedgerWriter,
	)

	s.Assert().NoError(err)
}

func createTransactionMeta(opMeta []xdr.OperationMeta) xdr.TransactionMeta {
	return xdr.TransactionMeta{
		V: 1,
		V1: &xdr.TransactionMetaV1{
			Operations: opMeta,
		},
	}
}
