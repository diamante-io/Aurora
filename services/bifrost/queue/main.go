package queue

type AssetCode string

const (
	AssetCodeBTC AssetCode = "BTC"
	AssetCodeETH AssetCode = "ETH"
)

type Transaction struct {
	TransactionID string
	AssetCode     AssetCode
	// CRITICAL REQUIREMENT: Amount in the base unit of currency.
	// For 10 satoshi this should be equal 0.0000001
	// For 1 BTC      this should be equal 1.0000000
	// For 1 Finney   this should be equal 0.0010000
	// For 1 ETH      this should be equal 1.0000000
	// Currently, the length of Amount string shouldn't be longer than 17 characters.
	Amount           string
	HcNetPublicKey string
}

// Queue implements transactions queue.
// The queue must not allow duplicates (including history) or must implement deduplication
// interval so it should not allow duplicate entries for 5 minutes since the first
// entry with the same ID was added.
// This is a critical requirement! Otherwise ETH/BTC may be sent twice to HcNet account.
// If you don't know what to do, use default AWS SQS FIFO queue or DB queue.
type Queue interface {
	// QueueAdd inserts the element to this queue. If element already exists in a queue, it should
	// return nil.
	QueueAdd(tx Transaction) error
	// QueuePool receives and removes the head of this queue. Returns nil if no elements found.
	QueuePool() (*Transaction, error)
}

type SQSFiFo struct{}
