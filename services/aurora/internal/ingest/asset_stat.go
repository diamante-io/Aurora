package ingest

import (
	"database/sql"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

func (assetStats *AssetStats) init() {
	assetStats.batchInsertBuilder = &BatchInsertBuilder{
		TableName: AssetStatsTableName,
		Columns: []string{
			"id",
			"amount",
			"num_accounts",
			"flags",
			"toml",
		},
	}

	assetStats.toUpdate = make(map[string]xdr.Asset)
}

func (assetStats *AssetStats) handlePaymentOp(paymentOp *xdr.PaymentOp, sourceAccount *xdr.AccountId) error {
	err := assetStats.updateIfAssetIssuerInvolved(paymentOp.Asset, *sourceAccount)
	if err != nil {
		return err
	}
	return assetStats.updateIfAssetIssuerInvolved(paymentOp.Asset, paymentOp.Destination)
}

func defaultSourceAccount(sourceAccount *xdr.AccountId, defaultAccount *xdr.AccountId) *xdr.AccountId {
	if sourceAccount != nil {
		return sourceAccount
	}
	return defaultAccount
}

// AddAllAssets adds all assets to update list. Used in initialization of stats.
func (assetStats *AssetStats) AddAllAssetsFromCore() (int, error) {
	assetStats.initOnce.Do(assetStats.init)

	var assets []xdr.Asset
	coreQ := &core.Q{Session: assetStats.CoreSession}
	err := coreQ.AllAssets(&assets)
	if err != nil {
		return 0, errors.Wrap(err, "Error getting all assets")
	}

	for _, asset := range assets {
		assetStats.add(asset)
	}

	return len(assets), nil
}

func (assetStats *AssetStats) add(asset xdr.Asset) {
	assetStats.toUpdate[asset.String()] = asset
}

// IngestOperation updates the assetsModified using the passed in operation
func (assetStats *AssetStats) IngestOperation(op *xdr.Operation, source *xdr.AccountId) error {
	assetStats.initOnce.Do(assetStats.init)

	body := op.Body
	sourceAccount := defaultSourceAccount(op.SourceAccount, source)
	switch body.Type {
	// TODO NNS 2 need to fix GetCreateAssetID call when adding assets from account
	// case xdr.OperationTypeSetOptions:
	// 	assetStats.addAssetsFromAccount(coreQ, sourceAccount)
	case xdr.OperationTypePayment:
		// payments is the only operation where we currently perform the optimization of checking against the issuer
		return assetStats.handlePaymentOp(body.PaymentOp, sourceAccount)
	case xdr.OperationTypePathPayment:
		// if this gets expensive then we can limit it to only include those assets that includes the issuer
		assetStats.add(body.PathPaymentOp.DestAsset)
		assetStats.add(body.PathPaymentOp.SendAsset)
		for _, asset := range body.PathPaymentOp.Path {
			assetStats.add(asset)
		}
	case xdr.OperationTypeManageBuyOffer:
		// if this gets expensive then we can limit it to only include those assets that includes the issuer
		assetStats.add(body.ManageBuyOfferOp.Buying)
		assetStats.add(body.ManageBuyOfferOp.Selling)
	case xdr.OperationTypeManageSellOffer:
		// if this gets expensive then we can limit it to only include those assets that includes the issuer
		assetStats.add(body.ManageSellOfferOp.Buying)
		assetStats.add(body.ManageSellOfferOp.Selling)
	case xdr.OperationTypeCreatePassiveSellOffer:
		// if this gets expensive then we can limit it to only include those assets that includes the issuer
		assetStats.add(body.CreatePassiveSellOfferOp.Buying)
		assetStats.add(body.CreatePassiveSellOfferOp.Selling)
	case xdr.OperationTypeChangeTrust:
		assetStats.add(body.ChangeTrustOp.Line)
	case xdr.OperationTypeAllowTrust:
		asset := body.AllowTrustOp.Asset.ToAsset(*sourceAccount)
		assetStats.add(asset)
	}

	return nil
}

// UpdateAssetStats updates the db with the latest asset stats for the assets that were modified
func (assetStats *AssetStats) UpdateAssetStats() error {
	assetStats.initOnce.Do(assetStats.init)

	hasValue := false
	for _, asset := range assetStats.toUpdate {
		assetStat, err := assetStats.computeAssetStat(&asset)
		if err != nil {
			return err
		}

		if assetStat != nil {
			hasValue = true
			assetStats.batchInsertBuilder.Values(
				assetStat.ID,
				assetStat.Amount,
				assetStat.NumAccounts,
				assetStat.Flags,
				assetStat.Toml,
			)
		}
	}

	if hasValue {
		// perform a delete first since upsert is not supported if postgres < 9.5
		err := errors.Wrap(assetStats.deleteRows(assetStats.HistorySession), "Error deleting asset_stats row")
		if err != nil {
			return err
		}

		// can perform a direct upsert if postgres > 9.4
		// is.Ingestion.assetStats = is.Ingestion.assetStats.
		// 	Suffix("ON CONFLICT (id) DO UPDATE SET (amount, num_accounts, flags, toml) = (excluded.amount, excluded.num_accounts, excluded.flags, excluded.toml)")
		return errors.Wrap(assetStats.batchInsertBuilder.Exec(assetStats.HistorySession), "Error inserting asset_stats row")
	}

	return nil
}

// func (assetStats *AssetStats) addAssetsFromAccount(coreQ *core.Q, account *xdr.AccountId) {
// 	if account == nil {
// 		return
// 	}

// 	var assets []xdr.Asset
// 	coreQ.AssetsForAddress(&assets, account.Address())

// 	for _, asset := range assets {
// 		if asset.Type != xdr.AssetTypeAssetTypeNative {
// 			assetStats.add(asset)
// 		}
// 	}
// }

func (assetStats *AssetStats) deleteRows(session *db.Session) error {
	if len(assetStats.toUpdate) == 0 {
		return nil
	}

	assets := make([]xdr.Asset, 0, len(assetStats.toUpdate))
	for _, asset := range assetStats.toUpdate {
		assets = append(assets, asset)
	}

	historyQ := history.Q{Session: assetStats.HistorySession}
	ids, err := historyQ.GetAssetIDs(assets)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	deleteStmt := sq.Delete("asset_stats").Where(sq.Eq{"id": ids})
	_, err = session.Exec(deleteStmt)
	return err
}

func (assetStats *AssetStats) updateIfAssetIssuerInvolved(asset xdr.Asset, account xdr.AccountId) error {
	var assetType, assetCode, assetIssuer string
	err := asset.Extract(&assetType, &assetCode, &assetIssuer)
	if err != nil {
		return err
	}

	if assetIssuer == account.Address() {
		assetStats.add(asset)
	}
	return nil
}

func (assetStats *AssetStats) computeAssetStat(asset *xdr.Asset) (*history.AssetStat, error) {
	if asset.Type == xdr.AssetTypeAssetTypeNative {
		return nil, nil
	}

	historyQ := &history.Q{Session: assetStats.HistorySession}
	assetID, err := historyQ.GetCreateAssetID(*asset)
	if err != nil {
		return nil, errors.Wrap(err, "historyQ.GetCreateAssetID error")
	}

	var assetType xdr.AssetType
	var assetCode, assetIssuer string
	err = asset.Extract(&assetType, &assetCode, &assetIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "asset.Extract error")
	}

	numAccounts, amount, err := statTrustlinesInfo(assetStats.CoreSession, assetType, assetCode, assetIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "statTrustlinesInfo error")
	}

	flags, toml, err := statAccountInfo(assetStats.CoreSession, assetIssuer)
	if err != nil {
		return nil, errors.Wrap(err, "statAccountInfo error")
	}

	return &history.AssetStat{
		ID:          assetID,
		Amount:      amount,
		NumAccounts: numAccounts,
		Flags:       flags,
		Toml:        toml,
	}, nil
}

// statTrustlinesInfo fetches all the stats from the trustlines table
func statTrustlinesInfo(coreSession *db.Session, assetType xdr.AssetType, assetCode string, assetIssuer string) (int32, string, error) {
	coreQ := &core.Q{Session: coreSession}
	return coreQ.BalancesForAsset(int32(assetType), assetCode, assetIssuer)
}

// statAccountInfo fetches all the stats from the accounts table
func statAccountInfo(coreSession *db.Session, accountID string) (int8, string, error) {
	var account core.Account
	// We don't need liabilities data here so let's use the old V9 query
	coreQ := &core.Q{Session: coreSession}
	err := coreQ.AccountByAddress(&account, accountID)
	if err != nil {
		// It is possible that issuer account has been deleted but issued assets
		// are still in circulation. In such case we return default values in 0.15.x
		// but a new field (`deleted`?) should be introduced in 0.16.0.
		// See: https://github.com/diamnet/diamnet-core/issues/1835
		if err == sql.ErrNoRows {
			return 0, "", nil
		}
		return -1, "", errors.Wrap(err, "coreQ.AccountByAddress error")
	}

	var toml string
	if !account.HomeDomain.Valid {
		toml = ""
	} else {
		trimmed := strings.TrimSpace(account.HomeDomain.String)
		if trimmed != "" {
			toml = "https://" + account.HomeDomain.String + "/.well-known/diamnet.toml"
		} else {
			toml = ""
		}
	}

	return int8(account.Flags), toml, nil
}
