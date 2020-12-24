package bridge

import (
	"encoding/json"
	"strconv"

	shared "github.com/diamnet/go/services/internal/bridge-compliance-shared"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
	"github.com/diamnet/go/txnbuild"
)

// OperationType is the type of operation
type OperationType string

const (
	// OperationTypeCreateAccount represents create_account operation
	OperationTypeCreateAccount OperationType = "create_account"
	// OperationTypePayment represents payment operation
	OperationTypePayment OperationType = "payment"
	// OperationTypePathPayment represents path_payment operation
	OperationTypePathPayment OperationType = "path_payment"
	// OperationTypeManageOffer represents manage_offer operation
	OperationTypeManageOffer OperationType = "manage_offer"
	// OperationTypeCreatePassiveOffer represents create_passive_offer operation
	OperationTypeCreatePassiveOffer OperationType = "create_passive_offer"
	// OperationTypeSetOptions represents set_options operation
	OperationTypeSetOptions OperationType = "set_options"
	// OperationTypeChangeTrust represents change_trust operation
	OperationTypeChangeTrust OperationType = "change_trust"
	// OperationTypeAllowTrust represents allow_trust operation
	OperationTypeAllowTrust OperationType = "allow_trust"
	// OperationTypeAccountMerge represents account_merge operation
	OperationTypeAccountMerge OperationType = "account_merge"
	// OperationTypeInflation represents inflation operation
	OperationTypeInflation OperationType = "inflation"
	// OperationTypeManageData represents manage_data operation
	OperationTypeManageData OperationType = "manage_data"
)

// BuilderRequest represents request made to /builder endpoint of bridge server
type BuilderRequest struct {
	Source         string
	SequenceNumber string `json:"sequence_number"`
	Operations     []Operation
	Signers        []string
}

// Process parses operations and creates OperationBody object for each operation
func (r BuilderRequest) Process() error {
	var err error
	for i, operation := range r.Operations {
		var operationBody OperationBody

		switch operation.Type {
		case OperationTypeCreateAccount:
			var createAccount CreateAccountOperationBody
			err = json.Unmarshal(operation.RawBody, &createAccount)
			operationBody = createAccount
		case OperationTypePayment:
			var payment PaymentOperationBody
			err = json.Unmarshal(operation.RawBody, &payment)
			operationBody = payment
		case OperationTypePathPayment:
			var pathPayment PathPaymentOperationBody
			err = json.Unmarshal(operation.RawBody, &pathPayment)
			operationBody = pathPayment
		case OperationTypeManageOffer:
			var manageOffer ManageOfferOperationBody
			err = json.Unmarshal(operation.RawBody, &manageOffer)
			operationBody = manageOffer
		case OperationTypeCreatePassiveOffer:
			var manageOffer ManageOfferOperationBody
			err = json.Unmarshal(operation.RawBody, &manageOffer)
			manageOffer.PassiveOffer = true
			operationBody = manageOffer
		case OperationTypeSetOptions:
			var setOptions SetOptionsOperationBody
			err = json.Unmarshal(operation.RawBody, &setOptions)
			operationBody = setOptions
		case OperationTypeChangeTrust:
			var changeTrust ChangeTrustOperationBody
			err = json.Unmarshal(operation.RawBody, &changeTrust)
			operationBody = changeTrust
		case OperationTypeAllowTrust:
			var allowTrust AllowTrustOperationBody
			err = json.Unmarshal(operation.RawBody, &allowTrust)
			operationBody = allowTrust
		case OperationTypeAccountMerge:
			var accountMerge AccountMergeOperationBody
			err = json.Unmarshal(operation.RawBody, &accountMerge)
			operationBody = accountMerge
		case OperationTypeInflation:
			var inflation InflationOperationBody
			err = json.Unmarshal(operation.RawBody, &inflation)
			operationBody = inflation
		case OperationTypeManageData:
			var manageData ManageDataOperationBody
			err = json.Unmarshal(operation.RawBody, &manageData)
			operationBody = manageData
		default:
			return helpers.NewInvalidParameterError("operations["+strconv.Itoa(i)+"][type]", "Invalid operation type.")
		}

		if err != nil {
			return helpers.NewInvalidParameterError("operations["+strconv.Itoa(i)+"][body]", "Operation is invalid: "+err.Error())
		}

		r.Operations[i].Body = operationBody
	}

	return nil
}

// Validate validates if the request is correct.
func (r BuilderRequest) Validate() error {
	if !shared.IsValidAccountID(r.Source) {
		return helpers.NewInvalidParameterError("source", "Source parameter must start with `G`.")
	}

	for i, signer := range r.Signers {
		if !shared.IsValidSecret(signer) {
			return helpers.NewInvalidParameterError("signers["+strconv.Itoa(i)+"]", "Signer must start with `S`.")
		}
	}

	for _, operation := range r.Operations {
		err := operation.Body.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Operation struct contains operation type and body
type Operation struct {
	Type    OperationType
	RawBody json.RawMessage `json:"body"` // Delay parsing until we know operation type
	Body    OperationBody   `json:"-"`    // Created during processing stage
}

// OperationBody interface is a common interface for builder operations
type OperationBody interface {
	// ToTransactionMutator() b.TransactionMutator
	Build() txnbuild.Operation
	Validate() error
}

// BuilderResponse represents response returned by /builder endpoint of bridge server
type BuilderResponse struct {
	helpers.SuccessResponse
	TransactionEnvelope string `json:"transaction_envelope"`
}

// Marshal marshals BuilderResponse
func (response *BuilderResponse) Marshal() ([]byte, error) {
	return json.MarshalIndent(response, "", "  ")
}
