package aurora

import (
	"net/http"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	hProblem "github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/services/aurora/internal/txsub"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/support/render/problem"
)

// Interface verification
var _ actions.JSONer = (*TransactionCreateAction)(nil)

// TransactionCreateAction submits a transaction to the diamnet-core network
// on behalf of the requesting client.
type TransactionCreateAction struct {
	Action
	TX       string
	Result   txsub.Result
	Resource aurora.TransactionSuccess
}

// JSON format action handler
func (action *TransactionCreateAction) JSON() error {
	action.Do(
		action.loadTX,
		action.loadResult,
		action.loadResource,
		func() { hal.Render(action.W, action.Resource) },
	)
	return action.Err
}

func (action *TransactionCreateAction) loadTX() {
	action.ValidateBodyType()
	action.TX = action.GetString("tx")
}

func (action *TransactionCreateAction) loadResult() {
	submission := action.App.submitter.Submit(action.R.Context(), action.TX)

	select {
	case result := <-submission:
		action.Result = result
	case <-action.R.Context().Done():
		action.Err = &hProblem.Timeout
	}
}

func (action *TransactionCreateAction) loadResource() {
	if action.Result.Err == nil {
		resourceadapter.PopulateTransactionSuccess(action.R.Context(), &action.Resource, action.Result)
		return
	}

	if action.Result.Err == txsub.ErrTimeout {
		action.Err = &hProblem.Timeout
		return
	}

	if action.Result.Err == txsub.ErrCanceled {
		action.Err = &hProblem.Timeout
		return
	}

	switch err := action.Result.Err.(type) {
	case *txsub.FailedTransactionError:
		rcr := aurora.TransactionResultCodes{}
		resourceadapter.PopulateTransactionResultCodes(action.R.Context(), &rcr, err)

		action.Err = &problem.P{
			Type:   "transaction_failed",
			Title:  "Transaction Failed",
			Status: http.StatusBadRequest,
			Detail: "The transaction failed when submitted to the diamnet network. " +
				"The `extras.result_codes` field on this response contains further " +
				"details.  Descriptions of each code can be found at: " +
				"https://www.diamnet.org/developers/learn/concepts/list-of-operations.html",
			Extras: map[string]interface{}{
				"envelope_xdr": action.Result.EnvelopeXDR,
				"result_xdr":   err.ResultXDR,
				"result_codes": rcr,
			},
		}
	case *txsub.MalformedTransactionError:
		action.Err = &problem.P{
			Type:   "transaction_malformed",
			Title:  "Transaction Malformed",
			Status: http.StatusBadRequest,
			Detail: "Aurora could not decode the transaction envelope in this " +
				"request. A transaction should be an XDR TransactionEnvelope struct " +
				"encoded using base64.  The envelope read from this request is " +
				"echoed in the `extras.envelope_xdr` field of this response for your " +
				"convenience.",
			Extras: map[string]interface{}{
				"envelope_xdr": err.EnvelopeXDR,
			},
		}
	default:
		action.Err = err
	}
}
