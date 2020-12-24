package processors

import (
	"context"
	stdio "io"
	"testing"

	"github.com/diamnet/go/exp/ingest/io"
	"github.com/diamnet/go/exp/orderbook"
	"github.com/diamnet/go/xdr"
)

func TestProcessOrderBookState(t *testing.T) {
	reader := &io.MockStateReader{}
	writer := &io.MockStateWriter{}
	graph := orderbook.NewOrderBookGraph()
	processor := OrderbookProcessor{graph}

	reader.On("Read").Return(xdr.LedgerEntryChange{}, stdio.EOF).Once()
	reader.On("Close").Return(nil).Once()
	writer.On("Close").Return(nil).Once()
	if err := processor.ProcessState(context.Background(), nil, reader, writer); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	writer.AssertExpectations(t)
	if err := graph.Apply(); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !graph.IsEmpty() {
		t.Fatal("expected graph to be empty")
	}

	reader.On("Read").Return(
		xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
			State: &xdr.LedgerEntry{
				Data: xdr.LedgerEntryData{
					Type: xdr.LedgerEntryTypeOffer,
					Offer: &xdr.OfferEntry{
						OfferId: xdr.Int64(1),
						Price:   xdr.Price{1, 2},
					},
				},
			},
		},
		nil,
	).Once()
	reader.On("Read").Return(
		xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
			State: &xdr.LedgerEntry{
				Data: xdr.LedgerEntryData{
					Type: xdr.LedgerEntryTypeOffer,
					Offer: &xdr.OfferEntry{
						OfferId: xdr.Int64(2),
						Price:   xdr.Price{1, 2},
					},
				},
			},
		},
		nil,
	).Once()
	reader.On("Read").Return(
		xdr.LedgerEntryChange{
			Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
			State: &xdr.LedgerEntry{
				Data: xdr.LedgerEntryData{
					Type: xdr.LedgerEntryTypeOffer,
					Offer: &xdr.OfferEntry{
						OfferId: xdr.Int64(3),
						Price:   xdr.Price{1, 2},
					},
				},
			},
		},
		nil,
	).Once()
	reader.On("Read").Return(xdr.LedgerEntryChange{}, stdio.EOF).Once()
	reader.On("Close").Return(nil).Once()
	writer.On("Close").Return(nil).Once()

	if err := processor.ProcessState(context.Background(), nil, reader, writer); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	writer.AssertExpectations(t)
	reader.AssertExpectations(t)
	if err := graph.Apply(); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expectedOffers := map[xdr.Int64]bool{
		xdr.Int64(1): true,
		xdr.Int64(2): true,
		xdr.Int64(3): true,
	}

	offers := graph.Offers()
	for _, offer := range offers {
		if !expectedOffers[offer.OfferId] {
			t.Fatalf("unexpected offer id %v", offer.OfferId)
		}
		delete(expectedOffers, offer.OfferId)
	}
	if len(expectedOffers) != 0 {
		t.Fatal("expected offers does not match offers in graph")
	}
}

func TestProcessOrderBookLedger(t *testing.T) {
	reader := &io.MockLedgerReader{}
	writer := &io.MockLedgerWriter{}
	graph := orderbook.NewOrderBookGraph()
	processor := OrderbookProcessor{graph}

	reader.On("Read").Return(io.LedgerTransaction{}, stdio.EOF).Once()
	reader.On("Close").Return(nil).Once()
	writer.On("Close").Return(nil).Once()
	if err := processor.ProcessLedger(context.Background(), nil, reader, writer); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	writer.AssertExpectations(t)
	reader.AssertExpectations(t)
	if err := graph.Apply(); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !graph.IsEmpty() {
		t.Fatal("expected graph to be empty")
	}

	// should be ignored because it's not an offer type
	reader.On("Read").
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

	// should be ignored because transaction was not successful
	reader.On("Read").
		Return(io.LedgerTransaction{
			Result: xdr.TransactionResultPair{
				Result: xdr.TransactionResult{
					Result: xdr.TransactionResultResult{
						Code: xdr.TransactionResultCodeTxFailed,
					},
				},
			},
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						// State
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
							Created: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(6),
										Price:   xdr.Price{1, 2},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	// add offer
	reader.On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						// State
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
							Created: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(1),
										Price:   xdr.Price{1, 2},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	// add another 2 offers
	reader.On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
							Created: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(2),
										Price:   xdr.Price{1, 3},
									},
								},
							},
						},
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
							Created: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(3),
										Price:   xdr.Price{3, 1},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	// update an offer
	reader.On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						// State
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
							State: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(2),
										Price:   xdr.Price{1, 3},
									},
								},
							},
						},
						// Updated
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryUpdated,
							Updated: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(2),
										Price:   xdr.Price{1, 6},
									},
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	reader.On("Read").
		Return(io.LedgerTransaction{
			Meta: createTransactionMeta([]xdr.OperationMeta{
				xdr.OperationMeta{
					Changes: []xdr.LedgerEntryChange{
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryState,
							State: &xdr.LedgerEntry{
								Data: xdr.LedgerEntryData{
									Type: xdr.LedgerEntryTypeOffer,
									Offer: &xdr.OfferEntry{
										OfferId: xdr.Int64(3),
										Price:   xdr.Price{3, 1},
									},
								},
							},
						},
						xdr.LedgerEntryChange{
							Type: xdr.LedgerEntryChangeTypeLedgerEntryRemoved,
							Removed: &xdr.LedgerKey{
								Type: xdr.LedgerEntryTypeOffer,
								Offer: &xdr.LedgerKeyOffer{
									OfferId: xdr.Int64(3),
								},
							},
						},
					},
				},
			}),
		}, nil).Once()

	reader.On("Read").
		Return(io.LedgerTransaction{}, stdio.EOF).Once()

	reader.On("Close").Return(nil).Once()
	writer.On("Close").Return(nil).Once()

	if err := processor.ProcessLedger(context.Background(), nil, reader, writer); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	writer.AssertExpectations(t)
	reader.AssertExpectations(t)
	if err := graph.Apply(); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expectedOffers := map[xdr.Int64]xdr.Price{
		xdr.Int64(1): xdr.Price{1, 2},
		xdr.Int64(2): xdr.Price{1, 6},
	}

	offers := graph.Offers()
	for _, offer := range offers {
		if price, ok := expectedOffers[offer.OfferId]; !ok {
			t.Fatalf("unexpected offer id %v", offer.OfferId)
		} else if offer.Price != price {
			t.Fatalf("unexpected offer price %v for offer with id %v", offer.Price, offer.OfferId)
		}
		delete(expectedOffers, offer.OfferId)
	}
	if len(expectedOffers) != 0 {
		t.Fatal("expected offers does not match offers in graph")
	}
}
