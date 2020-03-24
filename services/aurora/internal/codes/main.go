//Package codes is a helper package to help convert to transaction and operation result codes
//to strings used in aurora.
package codes

import (
	"github.com/go-errors/errors"
	"github.com/hcnet/go/xdr"
)

// ErrUnknownCode is returned when an unexepcted value is provided to `String`
var ErrUnknownCode = errors.New("Unknown result code")

const (
	// OpSuccess is the string code used to specify the operation was successful
	OpSuccess = "op_success"
	// OpMalformed is the string code used to specify the operation was malformed
	// in some way.
	OpMalformed = "op_malformed"
	// OpUnderfunded is the string code used to specify the operation failed
	// due to a lack of funds.
	OpUnderfunded = "op_underfunded"

	// OpLowReserve is the string code used to specify the operation failed
	// because the account in question does not have enough balance to satisfy
	// what their new minimum balance would be.
	OpLowReserve = "op_low_reserve"

	// OpLineFull occurs when a payment would cause a destination account to
	// exceed their declared trust limit for the asset being sent.
	OpLineFull = "op_line_full"

	// OpNoIssuer occurs when a operation does not correctly specify an issuing
	// asset
	OpNoIssuer = "op_no_issuer"
)

//String returns the appropriate string representation of the provided result code
func String(code interface{}) (string, error) {
	switch code := code.(type) {
	case xdr.TransactionResultCode:
		switch code {
		case xdr.TransactionResultCodeTxSuccess:
			return "tx_success", nil
		case xdr.TransactionResultCodeTxFailed:
			return "tx_failed", nil
		case xdr.TransactionResultCodeTxTooEarly:
			return "tx_too_early", nil
		case xdr.TransactionResultCodeTxTooLate:
			return "tx_too_late", nil
		case xdr.TransactionResultCodeTxMissingOperation:
			return "tx_missing_operation", nil
		case xdr.TransactionResultCodeTxBadSeq:
			return "tx_bad_seq", nil
		case xdr.TransactionResultCodeTxBadAuth:
			return "tx_bad_auth", nil
		case xdr.TransactionResultCodeTxInsufficientBalance:
			return "tx_insufficient_balance", nil
		case xdr.TransactionResultCodeTxNoAccount:
			return "tx_no_source_account", nil
		case xdr.TransactionResultCodeTxInsufficientFee:
			return "tx_insufficient_fee", nil
		case xdr.TransactionResultCodeTxBadAuthExtra:
			return "tx_bad_auth_extra", nil
		case xdr.TransactionResultCodeTxInternalError:
			return "tx_internal_error", nil
		}
	case xdr.OperationResultCode:
		switch code {
		case xdr.OperationResultCodeOpInner:
			return "op_inner", nil
		case xdr.OperationResultCodeOpBadAuth:
			return "op_bad_auth", nil
		case xdr.OperationResultCodeOpNoAccount:
			return "op_no_source_account", nil
		case xdr.OperationResultCodeOpNotSupported:
			return "op_not_supported", nil
		case xdr.OperationResultCodeOpTooManySubentries:
			return "op_too_many_subentries", nil
		case xdr.OperationResultCodeOpExceededWorkLimit:
			return "op_exceeded_work_limit", nil
		}
	case xdr.CreateAccountResultCode:
		switch code {
		case xdr.CreateAccountResultCodeCreateAccountSuccess:
			return OpSuccess, nil
		case xdr.CreateAccountResultCodeCreateAccountMalformed:
			return OpMalformed, nil
		case xdr.CreateAccountResultCodeCreateAccountUnderfunded:
			return OpUnderfunded, nil
		case xdr.CreateAccountResultCodeCreateAccountLowReserve:
			return OpLowReserve, nil
		case xdr.CreateAccountResultCodeCreateAccountAlreadyExist:
			return "op_already_exists", nil
		}
	case xdr.PaymentResultCode:
		switch code {
		case xdr.PaymentResultCodePaymentSuccess:
			return OpSuccess, nil
		case xdr.PaymentResultCodePaymentMalformed:
			return OpMalformed, nil
		case xdr.PaymentResultCodePaymentUnderfunded:
			return OpUnderfunded, nil
		case xdr.PaymentResultCodePaymentSrcNoTrust:
			return "op_src_no_trust", nil
		case xdr.PaymentResultCodePaymentSrcNotAuthorized:
			return "op_src_not_authorized", nil
		case xdr.PaymentResultCodePaymentNoDestination:
			return "op_no_destination", nil
		case xdr.PaymentResultCodePaymentNoTrust:
			return "op_no_trust", nil
		case xdr.PaymentResultCodePaymentNotAuthorized:
			return "op_not_authorized", nil
		case xdr.PaymentResultCodePaymentLineFull:
			return OpLineFull, nil
		case xdr.PaymentResultCodePaymentNoIssuer:
			return OpNoIssuer, nil
		}
	case xdr.PathPaymentResultCode:
		switch code {
		case xdr.PathPaymentResultCodePathPaymentSuccess:
			return OpSuccess, nil
		case xdr.PathPaymentResultCodePathPaymentMalformed:
			return OpMalformed, nil
		case xdr.PathPaymentResultCodePathPaymentUnderfunded:
			return OpUnderfunded, nil
		case xdr.PathPaymentResultCodePathPaymentSrcNoTrust:
			return "op_src_no_trust", nil
		case xdr.PathPaymentResultCodePathPaymentSrcNotAuthorized:
			return "op_src_not_authorized", nil
		case xdr.PathPaymentResultCodePathPaymentNoDestination:
			return "op_no_destination", nil
		case xdr.PathPaymentResultCodePathPaymentNoTrust:
			return "op_no_trust", nil
		case xdr.PathPaymentResultCodePathPaymentNotAuthorized:
			return "op_not_authorized", nil
		case xdr.PathPaymentResultCodePathPaymentLineFull:
			return OpLineFull, nil
		case xdr.PathPaymentResultCodePathPaymentNoIssuer:
			return OpNoIssuer, nil
		case xdr.PathPaymentResultCodePathPaymentTooFewOffers:
			return "op_too_few_offers", nil
		case xdr.PathPaymentResultCodePathPaymentOfferCrossSelf:
			return "op_cross_self", nil
		case xdr.PathPaymentResultCodePathPaymentOverSendmax:
			return "op_over_source_max", nil
		}
	case xdr.ManageBuyOfferResultCode:
		switch code {
		case xdr.ManageBuyOfferResultCodeManageBuyOfferSuccess:
			return OpSuccess, nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferMalformed:
			return OpMalformed, nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferSellNoTrust:
			return "op_sell_no_trust", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferBuyNoTrust:
			return "op_buy_no_trust", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferSellNotAuthorized:
			return "sell_not_authorized", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferBuyNotAuthorized:
			return "buy_not_authorized", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferLineFull:
			return OpLineFull, nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferUnderfunded:
			return OpUnderfunded, nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferCrossSelf:
			return "op_cross_self", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferSellNoIssuer:
			return "op_sell_no_issuer", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferBuyNoIssuer:
			return "buy_no_issuer", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferNotFound:
			return "op_offer_not_found", nil
		case xdr.ManageBuyOfferResultCodeManageBuyOfferLowReserve:
			return OpLowReserve, nil
		}
	case xdr.ManageSellOfferResultCode:
		switch code {
		case xdr.ManageSellOfferResultCodeManageSellOfferSuccess:
			return OpSuccess, nil
		case xdr.ManageSellOfferResultCodeManageSellOfferMalformed:
			return OpMalformed, nil
		case xdr.ManageSellOfferResultCodeManageSellOfferSellNoTrust:
			return "op_sell_no_trust", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferBuyNoTrust:
			return "op_buy_no_trust", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferSellNotAuthorized:
			return "sell_not_authorized", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferBuyNotAuthorized:
			return "buy_not_authorized", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferLineFull:
			return OpLineFull, nil
		case xdr.ManageSellOfferResultCodeManageSellOfferUnderfunded:
			return OpUnderfunded, nil
		case xdr.ManageSellOfferResultCodeManageSellOfferCrossSelf:
			return "op_cross_self", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferSellNoIssuer:
			return "op_sell_no_issuer", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferBuyNoIssuer:
			return "buy_no_issuer", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferNotFound:
			return "op_offer_not_found", nil
		case xdr.ManageSellOfferResultCodeManageSellOfferLowReserve:
			return OpLowReserve, nil
		}
	case xdr.SetOptionsResultCode:
		switch code {
		case xdr.SetOptionsResultCodeSetOptionsSuccess:
			return OpSuccess, nil
		case xdr.SetOptionsResultCodeSetOptionsLowReserve:
			return OpLowReserve, nil
		case xdr.SetOptionsResultCodeSetOptionsTooManySigners:
			return "op_too_many_signers", nil
		case xdr.SetOptionsResultCodeSetOptionsBadFlags:
			return "op_bad_flags", nil
		case xdr.SetOptionsResultCodeSetOptionsInvalidInflation:
			return "op_invalid_inflation", nil
		case xdr.SetOptionsResultCodeSetOptionsCantChange:
			return "op_cant_change", nil
		case xdr.SetOptionsResultCodeSetOptionsUnknownFlag:
			return "op_unknown_flag", nil
		case xdr.SetOptionsResultCodeSetOptionsThresholdOutOfRange:
			return "op_threshold_out_of_range", nil
		case xdr.SetOptionsResultCodeSetOptionsBadSigner:
			return "op_bad_signer", nil
		case xdr.SetOptionsResultCodeSetOptionsInvalidHomeDomain:
			return "op_invalid_home_domain", nil
		}
	case xdr.ChangeTrustResultCode:
		switch code {
		case xdr.ChangeTrustResultCodeChangeTrustSuccess:
			return OpSuccess, nil
		case xdr.ChangeTrustResultCodeChangeTrustMalformed:
			return OpMalformed, nil
		case xdr.ChangeTrustResultCodeChangeTrustNoIssuer:
			return OpNoIssuer, nil
		case xdr.ChangeTrustResultCodeChangeTrustInvalidLimit:
			return "op_invalid_limit", nil
		case xdr.ChangeTrustResultCodeChangeTrustLowReserve:
			return OpLowReserve, nil
		case xdr.ChangeTrustResultCodeChangeTrustSelfNotAllowed:
			return "op_self_not_allowed", nil
		}
	case xdr.AllowTrustResultCode:
		switch code {
		case xdr.AllowTrustResultCodeAllowTrustSuccess:
			return OpSuccess, nil
		case xdr.AllowTrustResultCodeAllowTrustMalformed:
			return OpMalformed, nil
		case xdr.AllowTrustResultCodeAllowTrustNoTrustLine:
			return "op_no_trustline", nil
		case xdr.AllowTrustResultCodeAllowTrustTrustNotRequired:
			return "op_not_required", nil
		case xdr.AllowTrustResultCodeAllowTrustCantRevoke:
			return "op_cant_revoke", nil
		}
	case xdr.AccountMergeResultCode:
		switch code {
		case xdr.AccountMergeResultCodeAccountMergeSuccess:
			return OpSuccess, nil
		case xdr.AccountMergeResultCodeAccountMergeMalformed:
			return OpMalformed, nil
		case xdr.AccountMergeResultCodeAccountMergeNoAccount:
			return "op_no_account", nil
		case xdr.AccountMergeResultCodeAccountMergeImmutableSet:
			return "op_immutable_set", nil
		case xdr.AccountMergeResultCodeAccountMergeHasSubEntries:
			return "op_has_sub_entries", nil
		case xdr.AccountMergeResultCodeAccountMergeSeqnumTooFar:
			return "op_seq_num_too_far", nil
		case xdr.AccountMergeResultCodeAccountMergeDestFull:
			return "op_dest_full", nil
		}
	case xdr.InflationResultCode:
		switch code {
		case xdr.InflationResultCodeInflationSuccess:
			return OpSuccess, nil
		case xdr.InflationResultCodeInflationNotTime:
			return "op_not_time", nil
		}
	case xdr.ManageDataResultCode:
		switch code {
		case xdr.ManageDataResultCodeManageDataSuccess:
			return OpSuccess, nil
		case xdr.ManageDataResultCodeManageDataNotSupportedYet:
			return "op_not_supported_yet", nil
		case xdr.ManageDataResultCodeManageDataNameNotFound:
			return "op_data_name_not_found", nil
		case xdr.ManageDataResultCodeManageDataLowReserve:
			return "op_low_reserve", nil
		case xdr.ManageDataResultCodeManageDataInvalidName:
			return "op_data_invalid_name", nil
		}
	case xdr.BumpSequenceResultCode:
		switch code {
		case xdr.BumpSequenceResultCodeBumpSequenceSuccess:
			return OpSuccess, nil
		case xdr.BumpSequenceResultCodeBumpSequenceBadSeq:
			return "op_bad_seq", nil
		}
	}

	return "", errors.New(ErrUnknownCode)
}

// ForOperationResult returns the strong represtation used by aurora for the
// error code `opr`
func ForOperationResult(opr xdr.OperationResult) (string, error) {
	if opr.Code != xdr.OperationResultCodeOpInner {
		return String(opr.Code)
	}

	ir := opr.MustTr()
	var ic interface{}

	switch ir.Type {
	case xdr.OperationTypeCreateAccount:
		ic = ir.MustCreateAccountResult().Code
	case xdr.OperationTypePayment:
		ic = ir.MustPaymentResult().Code
	case xdr.OperationTypePathPayment:
		ic = ir.MustPathPaymentResult().Code
	case xdr.OperationTypeManageBuyOffer:
		ic = ir.MustManageBuyOfferResult().Code
	case xdr.OperationTypeManageSellOffer:
		ic = ir.MustManageSellOfferResult().Code
	case xdr.OperationTypeCreatePassiveSellOffer:
		ic = ir.MustCreatePassiveSellOfferResult().Code
	case xdr.OperationTypeSetOptions:
		ic = ir.MustSetOptionsResult().Code
	case xdr.OperationTypeChangeTrust:
		ic = ir.MustChangeTrustResult().Code
	case xdr.OperationTypeAllowTrust:
		ic = ir.MustAllowTrustResult().Code
	case xdr.OperationTypeAccountMerge:
		ic = ir.MustAccountMergeResult().Code
	case xdr.OperationTypeInflation:
		ic = ir.MustInflationResult().Code
	case xdr.OperationTypeManageData:
		ic = ir.MustManageDataResult().Code
	case xdr.OperationTypeBumpSequence:
		ic = ir.MustBumpSeqResult().Code
	}

	return String(ic)
}
