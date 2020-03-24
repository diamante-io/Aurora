package txnbuild

import (
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// Inflation represents the HcNet inflation operation. See
// https://www.hcnet.org/developers/guides/concepts/list-of-operations.html
type Inflation struct {
	SourceAccount Account
}

// BuildXDR for Inflation returns a fully configured XDR Operation.
func (inf *Inflation) BuildXDR() (xdr.Operation, error) {
	opType := xdr.OperationTypeInflation
	body, err := xdr.NewOperationBody(opType, nil)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, inf.SourceAccount)
	return op, nil
}
