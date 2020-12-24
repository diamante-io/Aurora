package txsub

import (
	"errors"
	"fmt"

	"github.com/diamnet/go/services/aurora/internal/codes"
	"github.com/diamnet/go/xdr"
)

var (
	ErrNoResults = errors.New("No result found")
	ErrCanceled  = errors.New("canceled")
	ErrTimeout   = errors.New("timeout")

	// ErrBadSequence is a canned error response for transactions whose sequence
	// number is wrong.
	ErrBadSequence = &FailedTransactionError{"AAAAAAAAAAD////7AAAAAA=="}
	// ErrNoAccount is returned when the source account for the transaction
	// cannot be found in the database
	ErrNoAccount = &FailedTransactionError{"AAAAAAAAAAD////4AAAAAA=="}
)

// FailedTransactionError represent an error that occurred because
// diamnet-core rejected the transaction.  ResultXDR is a base64
// encoded TransactionResult struct
type FailedTransactionError struct {
	ResultXDR string
}

func (err *FailedTransactionError) Error() string {
	return fmt.Sprintf("tx failed: %s", err.ResultXDR)
}

func (fte *FailedTransactionError) Result() (result xdr.TransactionResult, err error) {
	err = xdr.SafeUnmarshalBase64(fte.ResultXDR, &result)
	return
}

func (fte *FailedTransactionError) TransactionResultCode() (result string, err error) {
	r, err := fte.Result()
	if err != nil {
		return
	}

	result, err = codes.String(r.Result.Code)
	return
}

func (fte *FailedTransactionError) OperationResultCodes() (result []string, err error) {
	r, err := fte.Result()
	if err != nil {
		return
	}

	oprs, ok := r.Result.GetResults()

	if !ok {
		return
	}

	result = make([]string, len(oprs))

	for i, opr := range oprs {
		result[i], err = codes.ForOperationResult(opr)
		if err != nil {
			return
		}
	}

	return
}

// MalformedTransactionError represent an error that occurred because
// a TransactionEnvelope could not be decoded from the provided data.
type MalformedTransactionError struct {
	EnvelopeXDR string
}

func (err *MalformedTransactionError) Error() string {
	return "tx malformed"
}
