package ingest

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/hcnet/go/clients/hcnetcore"

	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/meta"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ingest/participants"
	"github.com/hcnet/go/support/errors"
	ilog "github.com/hcnet/go/support/log"
	sTime "github.com/hcnet/go/support/time"
	"github.com/hcnet/go/xdr"
)

// Run starts an attempt to ingest the range of ledgers specified in this
// session.
func (is *Session) Run() {
	if is.Cursor == nil {
		is.Err = errors.New("no cursor set on session")
		return
	}

	is.Err = is.Ingestion.Start()
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Ingestion.Start error")
		return
	}

	defer is.Ingestion.Rollback()

	var sectionStart, i int32

	for is.Cursor.NextLedger() {
		if sectionStart == 0 {
			sectionStart = is.Cursor.LedgerSequence()
		}
		// Ensure no errors in Cursor
		is.Err = errors.Wrap(is.Cursor.Err, "Cursor.NextLedger error")

		is.validateLedger()
		is.clearLedger()
		is.ingestLedger()
		is.flush()

		i++
		if i%100 == 0 {
			// Print status update every 100 ledgers. Can be helpful when reingesting or
			// backfilling large number of ledgers.
			log.WithFields(ilog.F{
				"start": sectionStart,
				"end":   is.Cursor.LedgerSequence(),
			}).Info("Status: ingested ledgers")
			sectionStart = 0
		}

		if is.Err != nil {
			break
		}
	}

	if is.Config.EnableAssetStats && is.Err == nil {
		is.Err = is.AssetStats.UpdateAssetStats()
	}

	if is.Err != nil {
		is.Ingestion.Rollback()
		return
	}

	if is.Cursor.Err != nil {
		is.Err = is.Cursor.Err
		return
	}

	is.Err = is.Ingestion.Close()
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Ingestion.Close error")
		return
	}

	is.Err = errors.Wrap(is.reportCursorState(), "reportCursorState error")
}

func (is *Session) clearLedger() {
	if is.Err != nil {
		return
	}

	if !is.ClearExisting {
		return
	}

	start := time.Now()
	is.Err = is.Ingestion.Clear(is.Cursor.LedgerRange())
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Ingestion.Clear error")
	}

	if is.Metrics != nil {
		is.Metrics.ClearLedgerTimer.Update(time.Since(start))
	}
}

func (is *Session) effectFlagDetails(flagDetails map[string]bool, flagPtr *xdr.Uint32, setValue bool) {
	if flagPtr != nil {
		flags := xdr.AccountFlags(*flagPtr)

		if flags&xdr.AccountFlagsAuthRequiredFlag != 0 {
			flagDetails["auth_required_flag"] = setValue
		}
		if flags&xdr.AccountFlagsAuthRevocableFlag != 0 {
			flagDetails["auth_revocable_flag"] = setValue
		}
		if flags&xdr.AccountFlagsAuthImmutableFlag != 0 {
			flagDetails["auth_immutable_flag"] = setValue
		}
	}
}

func (is *Session) flush() {
	if is.Err != nil {
		return
	}
	is.Err = is.Ingestion.Flush()
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Ingestion.Flush error")
	}
}

func (is *Session) ingestEffects() {
	if is.Err != nil {
		return
	}

	effects := &EffectIngestion{
		Dest:        is.Ingestion,
		OperationID: is.Cursor.OperationID(),
		parent:      is.Ingestion,
	}
	source := is.Cursor.OperationSourceAccount()
	opbody := is.Cursor.Operation().Body

	switch is.Cursor.OperationType() {
	case xdr.OperationTypeCreateAccount:
		op := opbody.MustCreateAccountOp()

		effects.Add(op.Destination, history.EffectAccountCreated,
			map[string]interface{}{
				"starting_balance": amount.String(op.StartingBalance),
			},
		)

		effects.Add(source, history.EffectAccountDebited,
			map[string]interface{}{
				"asset_type": "native",
				"amount":     amount.String(op.StartingBalance),
			},
		)

		effects.Add(op.Destination, history.EffectSignerCreated,
			map[string]interface{}{
				"public_key": op.Destination.Address(),
				"weight":     keypair.DefaultSignerWeight,
			},
		)

	case xdr.OperationTypePayment:
		op := opbody.MustPaymentOp()

		dets := map[string]interface{}{"amount": amount.String(op.Amount)}
		is.assetDetails(dets, op.Asset, "")
		effects.Add(op.Destination, history.EffectAccountCredited, dets)
		effects.Add(source, history.EffectAccountDebited, dets)

	case xdr.OperationTypePathPayment:
		op := opbody.MustPathPaymentOp()
		resultSuccess := is.Cursor.OperationResult().MustPathPaymentResult().MustSuccess()

		dets := map[string]interface{}{"amount": amount.String(op.DestAmount)}
		is.assetDetails(dets, op.DestAsset, "")
		effects.Add(op.Destination, history.EffectAccountCredited, dets)

		result := is.Cursor.OperationResult().MustPathPaymentResult()
		dets = map[string]interface{}{"amount": amount.String(result.SendAmount())}
		is.assetDetails(dets, op.SendAsset, "")
		effects.Add(source, history.EffectAccountDebited, dets)

		is.ingestTradeEffects(effects, source, resultSuccess.Offers)
	case xdr.OperationTypeManageBuyOffer:
		result := is.Cursor.OperationResult().MustManageBuyOfferResult().MustSuccess()
		is.ingestTradeEffects(effects, source, result.OffersClaimed)
	case xdr.OperationTypeManageSellOffer:
		result := is.Cursor.OperationResult().MustManageSellOfferResult().MustSuccess()
		is.ingestTradeEffects(effects, source, result.OffersClaimed)
	case xdr.OperationTypeCreatePassiveSellOffer:
		var claims []xdr.ClaimOfferAtom
		result := is.Cursor.OperationResult()

		// KNOWN ISSUE:  hcnet-core creates results for CreatePassiveOffer operations
		// with the wrong result arm set.
		if result.Type == xdr.OperationTypeManageSellOffer {
			claims = result.MustManageSellOfferResult().MustSuccess().OffersClaimed
		} else {
			claims = result.MustCreatePassiveSellOfferResult().MustSuccess().OffersClaimed
		}

		is.ingestTradeEffects(effects, source, claims)
	case xdr.OperationTypeSetOptions:
		op := opbody.MustSetOptionsOp()

		if op.HomeDomain != nil {
			effects.Add(source, history.EffectAccountHomeDomainUpdated,
				map[string]interface{}{
					"home_domain": string(*op.HomeDomain),
				},
			)
		}

		thresholdDetails := map[string]interface{}{}

		if op.LowThreshold != nil {
			thresholdDetails["low_threshold"] = *op.LowThreshold
		}

		if op.MedThreshold != nil {
			thresholdDetails["med_threshold"] = *op.MedThreshold
		}

		if op.HighThreshold != nil {
			thresholdDetails["high_threshold"] = *op.HighThreshold
		}

		if len(thresholdDetails) > 0 {
			effects.Add(source, history.EffectAccountThresholdsUpdated, thresholdDetails)
		}

		flagDetails := map[string]bool{}
		is.effectFlagDetails(flagDetails, op.SetFlags, true)
		is.effectFlagDetails(flagDetails, op.ClearFlags, false)

		if len(flagDetails) > 0 {
			effects.Add(source, history.EffectAccountFlagsUpdated, flagDetails)
		}

		if op.InflationDest != nil {
			effects.Add(source, history.EffectAccountInflationDestinationUpdated,
				map[string]interface{}{
					"inflation_destination": op.InflationDest.Address(),
				},
			)
		}

		is.ingestSignerEffects(effects, op)

	case xdr.OperationTypeChangeTrust:
		op := opbody.MustChangeTrustOp()
		dets := map[string]interface{}{"limit": amount.String(op.Limit)}
		key := xdr.LedgerKey{}
		effect := history.EffectType(0)

		is.assetDetails(dets, op.Line, "")

		key.SetTrustline(source, op.Line)

		before, after, err := is.Cursor.BeforeAndAfter(key)

		// NOTE:  when an account trusts itself, the transaction is successful but
		// no ledger entries are actually modified, leading to an "empty meta"
		// situation.  We simply continue on to the next operation in that scenario.
		if err == meta.ErrMetaNotFound {
			return
		}

		if err != nil {
			is.Err = errors.Wrap(err, "is.Cursor.BeforeAndAfter error")
			return
		}

		switch {
		case before == nil && after != nil:
			effect = history.EffectTrustlineCreated
		case before != nil && after == nil:
			effect = history.EffectTrustlineRemoved
		case before != nil && after != nil:
			effect = history.EffectTrustlineUpdated
		default:
			panic("Invalid before-and-after state")
		}

		effects.Add(source, effect, dets)
	case xdr.OperationTypeAllowTrust:
		op := opbody.MustAllowTrustOp()
		asset := op.Asset.ToAsset(source)
		dets := map[string]interface{}{
			"trustor": op.Trustor.Address(),
		}
		is.assetDetails(dets, asset, "")

		if op.Authorize {
			effects.Add(source, history.EffectTrustlineAuthorized, dets)
		} else {
			effects.Add(source, history.EffectTrustlineDeauthorized, dets)
		}

	case xdr.OperationTypeAccountMerge:
		dest := opbody.MustDestination()
		result := is.Cursor.OperationResult().MustAccountMergeResult()
		dets := map[string]interface{}{
			"amount":     amount.String(result.MustSourceAccountBalance()),
			"asset_type": "native",
		}
		effects.Add(source, history.EffectAccountDebited, dets)
		effects.Add(dest, history.EffectAccountCredited, dets)
		effects.Add(source, history.EffectAccountRemoved, map[string]interface{}{})
	case xdr.OperationTypeInflation:
		payouts := is.Cursor.OperationResult().MustInflationResult().MustPayouts()
		for _, payout := range payouts {
			effects.Add(payout.Destination, history.EffectAccountCredited,
				map[string]interface{}{
					"amount":     amount.String(payout.Amount),
					"asset_type": "native",
				},
			)
		}
	case xdr.OperationTypeManageData:
		op := opbody.MustManageDataOp()
		dets := map[string]interface{}{"name": op.DataName}
		key := xdr.LedgerKey{}
		effect := history.EffectType(0)

		key.SetData(source, string(op.DataName))

		before, after, err := is.Cursor.BeforeAndAfter(key)
		if err != nil {
			is.Err = errors.Wrap(err, "is.Cursor.BeforeAndAfter error")
			return
		}

		if after != nil {
			raw := after.Data.MustData().DataValue
			dets["value"] = base64.StdEncoding.EncodeToString(raw)
		}

		switch {
		case before == nil && after != nil:
			effect = history.EffectDataCreated
		case before != nil && after == nil:
			effect = history.EffectDataRemoved
		case before != nil && after != nil:
			effect = history.EffectDataUpdated
		default:
			panic("Invalid before-and-after state")
		}

		effects.Add(source, effect, dets)

	case xdr.OperationTypeBumpSequence:
		opChanges := is.Cursor.OperationChanges()
		if len(opChanges) > 0 {
			op := opbody.MustBumpSequenceOp()
			dets := map[string]interface{}{"new_seq": op.BumpTo}
			effects.Add(source, history.EffectSequenceBumped, dets)
		}

	default:
		is.Err = fmt.Errorf("Unknown operation type: %s", is.Cursor.OperationType())
		return
	}

	is.Err = effects.Finish()
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "effects.Finish error")
	}
}

// ingestLedger ingests the current ledger
func (is *Session) ingestLedger() {
	if is.Err != nil {
		return
	}

	start := time.Now()
	is.Ingestion.Ledger(
		is.Cursor.LedgerID(),
		is.Cursor.Ledger(),
		is.Cursor.SuccessfulTransactionCount(),
		is.Cursor.FailedTransactionCount(),
		is.Cursor.SuccessfulLedgerOperationCount(),
	)

	for is.Cursor.NextTx() {
		is.ingestTransaction()
	}

	is.Ingested++
	if is.Metrics != nil {
		is.Metrics.IngestLedgerTimer.Update(time.Since(start))
	}
}

func (is *Session) ingestOperation() {
	if is.Err != nil {
		return
	}

	is.Err = is.Ingestion.Operation(
		is.Cursor.OperationID(),
		is.Cursor.TransactionID(),
		is.Cursor.OperationOrder(),
		is.Cursor.OperationSourceAccount(),
		is.Cursor.OperationType(),
		is.operationDetails(),
	)
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Ingestion.Operation error")
		return
	}

	is.ingestOperationParticipants()

	if is.Cursor.Transaction().IsSuccessful() {
		is.ingestEffects()
		is.ingestTrades()

		if is.Config.EnableAssetStats && is.Err == nil {
			is.Err = is.AssetStats.IngestOperation(
				is.Cursor.Operation(),
				&is.Cursor.Transaction().Envelope.Tx.SourceAccount,
			)
		}
	}

	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "Cursor.AssetsModified.IngestOperation error")
		return
	}
}

func (is *Session) ingestOperationParticipants() {
	if is.Err != nil {
		return
	}

	// Find the participants
	var p []xdr.AccountId
	p, is.Err = participants.ForOperation(
		&is.Cursor.Transaction().Envelope.Tx,
		is.Cursor.Operation(),
	)
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "participants.ForOperation error")
		return
	}

	is.Ingestion.OperationParticipants(is.Cursor.OperationID(), p)
}

func (is *Session) ingestSignerEffects(effects *EffectIngestion, op xdr.SetOptionsOp) {
	source := is.Cursor.OperationSourceAccount()

	be, ae, err := is.Cursor.BeforeAndAfter(source.LedgerKey())
	if err != nil {
		is.Err = errors.Wrap(err, "Cursor.BeforeAndAfter error")
		return
	}

	// HACK (scott) 2017-11-27:  Prevent crashes when BeforeAndAfter fails to
	// correctly work.
	if be == nil || ae == nil {
		// TODO (scott) 2018-03-02: log some info to help us track down the crash, doofus
		return
	}

	beforeAccount := be.Data.MustAccount()
	afterAccount := ae.Data.MustAccount()

	before := beforeAccount.SignerSummary()
	after := afterAccount.SignerSummary()

	// if before and after are the same, the signers have not changed
	if reflect.DeepEqual(before, after) {
		return
	}

	for addy := range before {
		weight, ok := after[addy]
		if !ok {
			effects.Add(source, history.EffectSignerRemoved, map[string]interface{}{
				"public_key": addy,
			})
			continue
		}
		effects.Add(source, history.EffectSignerUpdated, map[string]interface{}{
			"public_key": addy,
			"weight":     weight,
		})
	}
	// Add the "created" effects
	for addy, weight := range after {
		// if `addy` is in before, the previous for loop should have recorded
		// the update, so skip this key
		if _, ok := before[addy]; ok {
			continue
		}

		effects.Add(source, history.EffectSignerCreated, map[string]interface{}{
			"public_key": addy,
			"weight":     weight,
		})
	}

}

func (is *Session) ingestTrades() {
	if is.Err != nil {
		return
	}

	cursor := is.Cursor
	buyer := cursor.OperationSourceAccount()
	buyOfferExists := false
	buyOffer := xdr.OfferEntry{}
	var trades []xdr.ClaimOfferAtom

	switch cursor.OperationType() {
	case xdr.OperationTypePathPayment:
		trades = cursor.OperationResult().
			MustPathPaymentResult().
			MustSuccess().
			Offers

	case xdr.OperationTypeManageBuyOffer:
		manageOfferResult := cursor.OperationResult().MustManageBuyOfferResult().MustSuccess()
		trades = manageOfferResult.OffersClaimed
		buyOffer, buyOfferExists = manageOfferResult.Offer.GetOffer()

	case xdr.OperationTypeManageSellOffer:
		manageOfferResult := cursor.OperationResult().MustManageSellOfferResult().MustSuccess()
		trades = manageOfferResult.OffersClaimed
		buyOffer, buyOfferExists = manageOfferResult.Offer.GetOffer()

	case xdr.OperationTypeCreatePassiveSellOffer:
		result := cursor.OperationResult()

		// KNOWN ISSUE:  hcnet-core creates results for CreatePassiveOffer operations
		// with the wrong result arm set.
		if result.Type == xdr.OperationTypeManageSellOffer {
			manageOfferResult := result.MustManageSellOfferResult().MustSuccess()
			trades = manageOfferResult.OffersClaimed
			buyOffer, buyOfferExists = manageOfferResult.Offer.GetOffer()
		} else {
			passiveOfferResult := result.MustCreatePassiveSellOfferResult().MustSuccess()
			trades = passiveOfferResult.OffersClaimed
			buyOffer, buyOfferExists = passiveOfferResult.Offer.GetOffer()
		}
	}
	q := history.Q{Session: is.Ingestion.DB}
	for i, trade := range trades {
		// hcnet-core will opportunisticly garbage collect invalid offers (in the
		// event that a trader spends down their balance).  These garbage collected
		// offers get emitted in the result with the amount values set to zero.
		//
		// These zeroed ClaimOfferAtom values do not represent trades, and so we
		// skip them.
		if trade.AmountBought == 0 && trade.AmountSold == 0 {
			continue
		}

		//extract original offer price
		key := xdr.LedgerKey{}
		key.SetOffer(trade.SellerId, uint64(trade.OfferId))
		before, _, err := is.Cursor.BeforeAndAfter(key)
		if err != nil {
			is.Err = errors.Wrap(err, "Cursor.BeforeAndAfter error")
			return
		}
		sellOfferPrice := before.Data.Offer.Price

		is.Err = q.InsertTrade(
			is.Cursor.OperationID(),
			int32(i),
			buyer,
			buyOfferExists,
			buyOffer,
			trade,
			sellOfferPrice,
			sTime.MillisFromSeconds(is.Cursor.Ledger().CloseTime),
		)
		if is.Err != nil {
			is.Err = errors.Wrap(is.Err, "q.InsertTrade error")
			return
		}
	}
}

func (is *Session) ingestTradeEffects(effects *EffectIngestion, buyer xdr.AccountId, claims []xdr.ClaimOfferAtom) {
	if is.Err != nil {
		return
	}

	for _, claim := range claims {
		if claim.AmountSold == 0 && claim.AmountBought == 0 {
			continue
		}

		seller := claim.SellerId
		bd, sd := is.tradeDetails(buyer, seller, claim)
		effects.Add(buyer, history.EffectTrade, bd)
		effects.Add(seller, history.EffectTrade, sd)
	}
}

func (is *Session) tradeDetails(buyer, seller xdr.AccountId, claim xdr.ClaimOfferAtom) (bd map[string]interface{}, sd map[string]interface{}) {
	bd = map[string]interface{}{
		"offer_id":      claim.OfferId,
		"seller":        seller.Address(),
		"bought_amount": amount.String(claim.AmountSold),
		"sold_amount":   amount.String(claim.AmountBought),
	}
	is.assetDetails(bd, claim.AssetSold, "bought_")
	is.assetDetails(bd, claim.AssetBought, "sold_")

	sd = map[string]interface{}{
		"offer_id":      claim.OfferId,
		"seller":        buyer.Address(),
		"bought_amount": amount.String(claim.AmountBought),
		"sold_amount":   amount.String(claim.AmountSold),
	}
	is.assetDetails(sd, claim.AssetBought, "bought_")
	is.assetDetails(sd, claim.AssetSold, "sold_")

	return
}

func (is *Session) ingestTransaction() {
	if is.Err != nil {
		return
	}

	if !is.Config.IngestFailedTransactions && !is.Cursor.Transaction().IsSuccessful() {
		return
	}

	is.Err = is.Ingestion.Transaction(
		is.Cursor.Transaction().IsSuccessful(),
		is.Cursor.TransactionID(),
		is.Cursor.Transaction(),
		is.Cursor.TransactionFee(),
	)

	if is.Err != nil {
		return
	}

	for is.Cursor.NextOp() {
		is.ingestOperation()
	}

	is.ingestTransactionParticipants()
}

func (is *Session) ingestTransactionParticipants() {
	if is.Err != nil {
		return
	}

	// Find the participants
	var p []xdr.AccountId
	p, is.Err = participants.ForTransaction(
		&is.Cursor.Transaction().Envelope.Tx,
		&is.Cursor.Transaction().ResultMeta,
		&is.Cursor.TransactionFee().Changes,
	)
	if is.Err != nil {
		is.Err = errors.Wrap(is.Err, "participants.ForTransaction error")
		return
	}

	is.Ingestion.TransactionParticipants(is.Cursor.TransactionID(), p)
}

// assetDetails sets the details for `a` on `result` using keys with `prefix`
func (is *Session) assetDetails(result map[string]interface{}, a xdr.Asset, prefix string) error {
	var (
		t    string
		code string
		i    string
	)
	err := a.Extract(&t, &code, &i)
	if err != nil {
		err = errors.Wrap(err, "xdr.Asset.Extract error")
		return err
	}
	result[prefix+"asset_type"] = t

	if a.Type == xdr.AssetTypeAssetTypeNative {
		return nil
	}

	result[prefix+"asset_code"] = code
	result[prefix+"asset_issuer"] = i
	return nil
}

// operationDetails returns the details regarding the current operation, suitable
// for ingestion into a history_operation row
func (is *Session) operationDetails() map[string]interface{} {
	details := map[string]interface{}{}
	c := is.Cursor
	source := c.OperationSourceAccount()

	switch c.OperationType() {
	case xdr.OperationTypeCreateAccount:
		op := c.Operation().Body.MustCreateAccountOp()
		details["funder"] = source.Address()
		details["account"] = op.Destination.Address()
		details["starting_balance"] = amount.String(op.StartingBalance)
	case xdr.OperationTypePayment:
		op := c.Operation().Body.MustPaymentOp()
		details["from"] = source.Address()
		details["to"] = op.Destination.Address()
		details["amount"] = amount.String(op.Amount)
		is.assetDetails(details, op.Asset, "")
	case xdr.OperationTypePathPayment:
		op := c.Operation().Body.MustPathPaymentOp()
		details["from"] = source.Address()
		details["to"] = op.Destination.Address()

		details["amount"] = amount.String(op.DestAmount)
		details["source_amount"] = amount.String(0)
		details["source_max"] = amount.String(op.SendMax)
		is.assetDetails(details, op.DestAsset, "")
		is.assetDetails(details, op.SendAsset, "source_")

		if c.Transaction().IsSuccessful() {
			result := c.OperationResult().MustPathPaymentResult()
			details["source_amount"] = amount.String(result.SendAmount())
		}

		var path = make([]map[string]interface{}, len(op.Path))
		for i := range op.Path {
			path[i] = make(map[string]interface{})
			is.assetDetails(path[i], op.Path[i], "")
		}
		details["path"] = path
	case xdr.OperationTypeManageBuyOffer:
		op := c.Operation().Body.MustManageBuyOfferOp()
		details["offer_id"] = op.OfferId
		details["amount"] = amount.String(op.BuyAmount)
		details["price"] = op.Price.String()
		details["price_r"] = map[string]interface{}{
			"n": op.Price.N,
			"d": op.Price.D,
		}
		is.assetDetails(details, op.Buying, "buying_")
		is.assetDetails(details, op.Selling, "selling_")
	case xdr.OperationTypeManageSellOffer:
		op := c.Operation().Body.MustManageSellOfferOp()
		details["offer_id"] = op.OfferId
		details["amount"] = amount.String(op.Amount)
		details["price"] = op.Price.String()
		details["price_r"] = map[string]interface{}{
			"n": op.Price.N,
			"d": op.Price.D,
		}
		is.assetDetails(details, op.Buying, "buying_")
		is.assetDetails(details, op.Selling, "selling_")
	case xdr.OperationTypeCreatePassiveSellOffer:
		op := c.Operation().Body.MustCreatePassiveSellOfferOp()
		details["amount"] = amount.String(op.Amount)
		details["price"] = op.Price.String()
		details["price_r"] = map[string]interface{}{
			"n": op.Price.N,
			"d": op.Price.D,
		}
		is.assetDetails(details, op.Buying, "buying_")
		is.assetDetails(details, op.Selling, "selling_")
	case xdr.OperationTypeSetOptions:
		op := c.Operation().Body.MustSetOptionsOp()

		if op.InflationDest != nil {
			details["inflation_dest"] = op.InflationDest.Address()
		}

		if op.SetFlags != nil && *op.SetFlags > 0 {
			is.operationFlagDetails(details, int32(*op.SetFlags), "set")
		}

		if op.ClearFlags != nil && *op.ClearFlags > 0 {
			is.operationFlagDetails(details, int32(*op.ClearFlags), "clear")
		}

		if op.MasterWeight != nil {
			details["master_key_weight"] = *op.MasterWeight
		}

		if op.LowThreshold != nil {
			details["low_threshold"] = *op.LowThreshold
		}

		if op.MedThreshold != nil {
			details["med_threshold"] = *op.MedThreshold
		}

		if op.HighThreshold != nil {
			details["high_threshold"] = *op.HighThreshold
		}

		if op.HomeDomain != nil {
			details["home_domain"] = *op.HomeDomain
		}

		if op.Signer != nil {
			details["signer_key"] = op.Signer.Key.Address()
			details["signer_weight"] = op.Signer.Weight
		}
	case xdr.OperationTypeChangeTrust:
		op := c.Operation().Body.MustChangeTrustOp()
		is.assetDetails(details, op.Line, "")
		details["trustor"] = source.Address()
		details["trustee"] = details["asset_issuer"]
		details["limit"] = amount.String(op.Limit)
	case xdr.OperationTypeAllowTrust:
		op := c.Operation().Body.MustAllowTrustOp()
		is.assetDetails(details, op.Asset.ToAsset(source), "")
		details["trustee"] = source.Address()
		details["trustor"] = op.Trustor.Address()
		details["authorize"] = op.Authorize
	case xdr.OperationTypeAccountMerge:
		aid := c.Operation().Body.MustDestination()
		details["account"] = source.Address()
		details["into"] = aid.Address()
	case xdr.OperationTypeInflation:
		// no inflation details, presently
	case xdr.OperationTypeManageData:
		op := c.Operation().Body.MustManageDataOp()
		details["name"] = string(op.DataName)
		if op.DataValue != nil {
			details["value"] = base64.StdEncoding.EncodeToString(*op.DataValue)
		} else {
			details["value"] = nil
		}
	case xdr.OperationTypeBumpSequence:
		op := c.Operation().Body.MustBumpSequenceOp()
		details["bump_to"] = fmt.Sprintf("%d", op.BumpTo)
	default:
		panic(fmt.Errorf("Unknown operation type: %s", c.OperationType()))
	}

	return details
}

// operationFlagDetails sets the account flag details for `f` on `result`.
func (is *Session) operationFlagDetails(result map[string]interface{}, f int32, prefix string) {
	var (
		n []int32
		s []string
	)

	if (f & int32(xdr.AccountFlagsAuthRequiredFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthRequiredFlag))
		s = append(s, "auth_required")
	}

	if (f & int32(xdr.AccountFlagsAuthRevocableFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthRevocableFlag))
		s = append(s, "auth_revocable")
	}

	if (f & int32(xdr.AccountFlagsAuthImmutableFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthImmutableFlag))
		s = append(s, "auth_immutable")
	}

	result[prefix+"_flags"] = n
	result[prefix+"_flags_s"] = s
}

// reportCursorState makes an http request to the configured hcnet-core server
// to report that it has finished processing the data being ingested.  This
// allows hcnet-core to free that storage when next it runs its own
// maintenance.
func (is *Session) reportCursorState() error {
	if is.HcNetCoreURL == "" {
		return nil
	}

	if is.SkipCursorUpdate {
		return nil
	}

	core := &hcnetcore.Client{URL: is.HcNetCoreURL}

	err := core.SetCursor(context.Background(), is.Cursor.Name, is.Cursor.LastLedger)

	if err != nil {
		return errors.Wrap(err, "SetCursor failed")
	}

	return nil
}

// validate ledger
func (is *Session) validateLedger() {
	if is.Err != nil {
		return
	}

	// TODO: if the cursor is running forward, load the previous legder and
	// validate.  if reverse, load the next ledger and validate.

	// if we can find no ledger where one should be, emit a warning because we
	// cannot validate.  The normal scenario for this to occur is an empty history
	// databse.

	// if hashes mistmatch, return an error

}
