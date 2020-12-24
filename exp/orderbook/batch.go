package orderbook

import (
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

const (
	_ = iota
	// the operationType enum values start at 1 because when constructing a
	// orderBookOperation struct, the operationType field should always be specified
	// explicity. if the operationType enum values started at 0 then it would be
	// possible to create a valid orderBookOperation struct without specifying
	// the operationType field
	addOfferOperationType    = iota
	removeOfferOperationType = iota
)

type orderBookOperation struct {
	operationType int
	offerID       xdr.Int64
	offer         *xdr.OfferEntry
}

type orderBookBatchedUpdates struct {
	operations []orderBookOperation
	orderbook  *OrderBookGraph
	committed  bool
}

// addOffer will queue an operation to add the given offer to the order book
func (tx *orderBookBatchedUpdates) addOffer(offer xdr.OfferEntry) *orderBookBatchedUpdates {
	tx.operations = append(tx.operations, orderBookOperation{
		operationType: addOfferOperationType,
		offerID:       offer.OfferId,
		offer:         &offer,
	})

	return tx
}

// removeOffer will queue an operation to remove the given offer from the order book
func (tx *orderBookBatchedUpdates) removeOffer(offerID xdr.Int64) *orderBookBatchedUpdates {
	tx.operations = append(tx.operations, orderBookOperation{
		operationType: removeOfferOperationType,
		offerID:       offerID,
	})

	return tx
}

// apply will attempt to apply all the updates in the batch to the order book
func (tx *orderBookBatchedUpdates) apply() error {
	tx.orderbook.lock.Lock()
	defer tx.orderbook.lock.Unlock()

	if tx.committed {
		// This should never happen
		panic(errBatchAlreadyApplied)
	}
	tx.committed = true

	for _, operation := range tx.operations {
		switch operation.operationType {
		case addOfferOperationType:
			if err := tx.orderbook.add(*operation.offer); err != nil {
				panic(errors.Wrap(err, "could not apply update in batch"))
			}
		case removeOfferOperationType:
			if err := tx.orderbook.remove(operation.offerID); err != nil {
				panic(errors.Wrap(err, "could not apply update in batch"))
			}
		default:
			panic(errors.New("invalid operation type"))
		}
	}

	return nil
}
