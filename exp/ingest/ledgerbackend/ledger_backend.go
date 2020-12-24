package ledgerbackend

import "github.com/diamnet/go/xdr"

// LedgerBackend represents the interface to a ledger data store.
type LedgerBackend interface {
	GetLatestLedgerSequence() (sequence uint32, err error)
	// The first returned value is false when the ledger does not exist in a backend.
	GetLedger(sequence uint32) (bool, LedgerCloseMeta, error)
	Close() error
}

// session is the interface needed to access a persistent database session.
// TODO can't use this until we add Close() to the existing db.Session object
type session interface {
	GetRaw(dest interface{}, query string, args ...interface{}) error
	SelectRaw(dest interface{}, query string, args ...interface{}) error
	Close() error
}

// LedgerCloseMeta is the information needed to reconstruct the history of transactions in a given ledger.
type LedgerCloseMeta struct {
	LedgerHeader          xdr.LedgerHeaderHistoryEntry
	TransactionEnvelope   []xdr.TransactionEnvelope
	TransactionResult     []xdr.TransactionResultPair
	TransactionMeta       []xdr.TransactionMeta
	TransactionFeeChanges []xdr.LedgerEntryChanges
}

// ledgerHeaderHistory is a helper struct used to unmarshall header fields from a diamnet-core DB.
type ledgerHeaderHistory struct {
	Hash   xdr.Hash         `db:"ledgerhash"`
	Header xdr.LedgerHeader `db:"data"`
}

// ledgerHeader holds a row of data from the diamnet-core `ledgerheaders` table.
type ledgerHeader struct {
	LedgerHash     string           `db:"ledgerhash"`
	PrevHash       string           `db:"prevhash"`
	BucketListHash string           `db:"bucketlisthash"`
	CloseTime      int64            `db:"closetime"`
	LedgerSeq      uint32           `db:"ledgerseq"`
	Data           xdr.LedgerHeader `db:"data"`
}

// txHistory holds a row of data from the diamnet-core `txhistory` table.
type txHistory struct {
	TXID      string                    `db:"txid"`
	LedgerSeq uint32                    `db:"ledgerseq"`
	TXIndex   uint32                    `db:"txindex"`
	TXBody    xdr.TransactionEnvelope   `db:"txbody"`
	TXResult  xdr.TransactionResultPair `db:"txresult"`
	TXMeta    xdr.TransactionMeta       `db:"txmeta"`
}

// txFeeHistory holds a row of data from the diamnet-core `txfeehistory` table.
type txFeeHistory struct {
	TXID      string                 `db:"txid"`
	LedgerSeq uint32                 `db:"ledgerseq"`
	TXIndex   uint32                 `db:"txindex"`
	TXChanges xdr.LedgerEntryChanges `db:"txchanges"`
}

// scpHistory holds a row of data from the diamnet-core `scphistory` table.
// type scpHistory struct {
// 	NodeID    string `db:"nodeid"`
// 	LedgerSeq uint32 `db:"ledgerseq"`
// 	Envelope  string `db:"envelope"`
// }

// upgradeHistory holds a row of data from the diamnet-core `upgradehistory` table.
// type upgradeHistory struct {
// 	LedgerSeq    uint32 `db:"ledgerseq"`
// 	UpgradeIndex uint32 `db:"upgradeindex"`
// 	Upgrade      string `db:"upgrade"`
// 	Changes      string `db:"changes"`
// }
