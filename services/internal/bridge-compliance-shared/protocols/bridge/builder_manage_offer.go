package bridge

import (
	"strconv"

	"github.com/hcnet/go/txnbuild"

	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols"
)

// ManageOfferOperationBody represents manage_offer operation
type ManageOfferOperationBody struct {
	PassiveOffer bool `json:"-"`
	Source       *string
	Selling      protocols.Asset
	Buying       protocols.Asset
	Amount       string
	Price        string
	OfferID      *string `json:"offer_id"`
}

// Build returns a txnbuild.Operation
func (op ManageOfferOperationBody) Build() txnbuild.Operation {
	if op.PassiveOffer {
		txnOp := txnbuild.CreatePassiveSellOffer{
			Selling: op.Selling.ToBaseAsset(),
			Buying:  op.Buying.ToBaseAsset(),
			Amount:  op.Amount,
			Price:   op.Price,
		}

		if op.Source != nil {
			txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
		}

		return &txnOp
	}

	txnOp := txnbuild.ManageSellOffer{
		Selling: txnbuild.CreditAsset{Code: op.Selling.Code, Issuer: op.Selling.Issuer},
		Buying:  txnbuild.CreditAsset{Code: op.Buying.Code, Issuer: op.Buying.Issuer},
		Amount:  op.Amount,
		Price:   op.Price,
	}

	if op.OfferID != nil {
		// Validated in Validate()
		u, _ := strconv.ParseInt(*op.OfferID, 10, 64)
		txnOp.OfferID = u
	}

	if op.Source != nil {
		txnOp.SourceAccount = &txnbuild.SimpleAccount{AccountID: *op.Source}
	}

	return &txnOp
}

// Validate validates if operation body is valid.
func (op ManageOfferOperationBody) Validate() error {
	if op.OfferID != nil {
		_, err := strconv.ParseInt(*op.OfferID, 10, 64)
		if err != nil {
			return helpers.NewInvalidParameterError("offer_id", "Not a number.")
		}
	}

	if op.Source != nil && !shared.IsValidAccountID(*op.Source) {
		return helpers.NewInvalidParameterError("source", "Source must be a public key (starting with `G`).")
	}

	return nil
}
