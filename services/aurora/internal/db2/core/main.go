// Package core contains database record definitions useable for
// reading rows from a DiamNet Core db
package core

import (
	"strconv"

	"github.com/guregu/null"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// Account is a row of data from the `accounts` table
type Account struct {
	Accountid          string
	Balance            xdr.Int64
	Seqnum             string
	Numsubentries      int32
	Inflationdest      null.String
	HomeDomain         null.String
	Thresholds         xdr.Thresholds
	Flags              xdr.AccountFlags
	LastModified       uint32
	BuyingLiabilities  xdr.Int64 `db:"buyingliabilities"`
	SellingLiabilities xdr.Int64 `db:"sellingliabilities"`
}

// AccountData is a row of data from the `accountdata` table
type AccountData struct {
	Accountid string
	Key       string `db:"dataname"`
	Value     string `db:"datavalue"`
}

// LedgerHeader is row of data from the `ledgerheaders` table
type LedgerHeader struct {
	LedgerHash     string           `db:"ledgerhash"`
	PrevHash       string           `db:"prevhash"`
	BucketListHash string           `db:"bucketlisthash"`
	CloseTime      int64            `db:"closetime"`
	Sequence       uint32           `db:"ledgerseq"`
	Data           xdr.LedgerHeader `db:"data"`
}

// Offer is row of data from the `offers` table from diamnet-core
type Offer struct {
	SellerID string `db:"sellerid"`
	OfferID  int64  `db:"offerid"`

	SellingAsset xdr.Asset `db:"sellingasset"`
	BuyingAsset  xdr.Asset `db:"buyingasset"`

	Amount       xdr.Int64 `db:"amount"`
	Pricen       int32     `db:"pricen"`
	Priced       int32     `db:"priced"`
	Price        float64   `db:"price"`
	Flags        int32     `db:"flags"`
	Lastmodified int32     `db:"lastmodified"`
}

// get returns Offer. Useful in `internalOffer` context, when Offer is embedded.
func (o Offer) get() Offer {
	return o
}

// internalOffer is row of data from the `offers` table from diamnet-core used
// internally only to support schema <=8.
type internalOffer struct {
	Offer

	// Schema v8 fields, for compatibility only.
	SellingAssetType xdr.AssetType `db:"sellingassettype"`
	SellingAssetCode null.String   `db:"sellingassetcode"`
	SellingIssuer    null.String   `db:"sellingissuer"`

	BuyingAssetType xdr.AssetType `db:"buyingassettype"`
	BuyingAssetCode null.String   `db:"buyingassetcode"`
	BuyingIssuer    null.String   `db:"buyingissuer"`
}

// OrderBookSummaryPriceLevel is a collapsed view of multiple offers at the same price that
// contains the summed amount from all the member offers. Used by OrderBookSummary
type OrderBookSummaryPriceLevel struct {
	Type string `db:"type"`
	PriceLevel
}

// OrderBookSummary is a summary of a set of offers for a given base and
// counter currency
type OrderBookSummary []OrderBookSummaryPriceLevel

// Q is a helper struct on which to hang common queries against a diamnet
// core database.
type Q struct {
	*db.Session
}

// PriceLevel represents an aggregation of offers to trade at a certain
// price.
type PriceLevel struct {
	Pricen int32   `db:"pricen"`
	Priced int32   `db:"priced"`
	Pricef float64 `db:"pricef"`
	Amount string  `db:"amount"`
}

// SequenceProvider implements `txsub.SequenceProvider`
type SequenceProvider struct {
	Q *Q
}

// Signer is a row of data from the `signers` table from diamnet-core
type Signer struct {
	Accountid string
	Publickey string
	Weight    int32
}

// Transaction is row of data from the `txhistory` table from diamnet-core
type Transaction struct {
	TransactionHash string                    `db:"txid"`
	LedgerSequence  int32                     `db:"ledgerseq"`
	Index           int32                     `db:"txindex"`
	Envelope        xdr.TransactionEnvelope   `db:"txbody"`
	Result          xdr.TransactionResultPair `db:"txresult"`
	ResultMeta      xdr.TransactionMeta       `db:"txmeta"`
}

// TransactionFee is row of data from the `txfeehistory` table from diamnet-core
type TransactionFee struct {
	TransactionHash string                 `db:"txid"`
	LedgerSequence  int32                  `db:"ledgerseq"`
	Index           int32                  `db:"txindex"`
	Changes         xdr.LedgerEntryChanges `db:"txchanges"`
}

// Trustline is a row of data from the `trustlines` table from diamnet-core
type Trustline struct {
	Accountid          string
	Assettype          xdr.AssetType
	Issuer             string
	Assetcode          string
	Tlimit             xdr.Int64
	Balance            xdr.Int64
	Flags              int32
	LastModified       uint32
	BuyingLiabilities  xdr.Int64 `db:"buyingliabilities"`
	SellingLiabilities xdr.Int64 `db:"sellingliabilities"`
}

// AssetFromDB produces an xdr.Asset by combining the constituent type, code and
// issuer, as often retrieved from the DB in 3 separate columns.
func AssetFromDB(typ xdr.AssetType, code string, issuer string) (result xdr.Asset, err error) {
	switch typ {
	case xdr.AssetTypeAssetTypeNative:
		result, err = xdr.NewAsset(xdr.AssetTypeAssetTypeNative, nil)
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		var (
			an      xdr.AssetAlphaNum4
			decoded []byte
			pkey    xdr.Uint256
		)

		copy(an.AssetCode[:], []byte(code))
		decoded, err = strkey.Decode(strkey.VersionByteAccountID, issuer)
		if err != nil {
			return
		}

		copy(pkey[:], decoded)
		an.Issuer, err = xdr.NewAccountId(xdr.PublicKeyTypePublicKeyTypeEd25519, pkey)
		if err != nil {
			return
		}
		result, err = xdr.NewAsset(xdr.AssetTypeAssetTypeCreditAlphanum4, an)
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		var (
			an      xdr.AssetAlphaNum12
			decoded []byte
			pkey    xdr.Uint256
		)

		copy(an.AssetCode[:], []byte(code))
		decoded, err = strkey.Decode(strkey.VersionByteAccountID, issuer)
		if err != nil {
			return
		}

		copy(pkey[:], decoded)
		an.Issuer, err = xdr.NewAccountId(xdr.PublicKeyTypePublicKeyTypeEd25519, pkey)
		if err != nil {
			return
		}
		result, err = xdr.NewAsset(xdr.AssetTypeAssetTypeCreditAlphanum12, an)
	}

	return
}

// ElderLedger represents the oldest "ingestable" ledger known to the
// diamnet-core database this ingestion system is communicating with.  Aurora,
// which wants to operate on a contiguous range of ledger data (i.e. free from
// gaps) uses the elder ledger to start importing in the case of an empty
// database.  NOTE:  This current query used is correct, but slow.  Please keep
// this query out of latency sensitive or frequently trafficked code paths.
func (q *Q) ElderLedger(dest *int32) error {
	err := q.GetRaw(dest, `
		SELECT COALESCE(ledgerseq, 0)

		FROM (
			SELECT 
				ledgerseq,
				LAG(ledgerseq, 1) OVER ( ORDER BY ledgerseq) as prev
			FROM ledgerheaders
		) seqs

		WHERE COALESCE(prev, -1) < ledgerseq - 1
		ORDER BY ledgerseq DESC
		LIMIT 1;
	`)

	return err
}

// LatestLedger loads the latest known ledger
func (q *Q) LatestLedger(dest interface{}) error {
	return q.GetRaw(dest, `SELECT COALESCE(MAX(ledgerseq), 0) FROM ledgerheaders`)
}

// SchemaVersion returns Core DB schema version
func (q *Q) SchemaVersion() (int, error) {
	var version string
	err := q.GetRaw(&version, `SELECT state FROM storestate WHERE statename = 'databaseschema'`)
	if err != nil {
		return 0, errors.Wrap(err, "Error getting 'databaseschema'")
	}

	return strconv.Atoi(version)
}
