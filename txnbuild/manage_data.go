package txnbuild

import (
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// ManageData represents the DiamNet manage data operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type ManageData struct {
	Name          string
	Value         []byte
	SourceAccount Account
}

// BuildXDR for ManageData returns a fully configured XDR Operation.
func (md *ManageData) BuildXDR() (xdr.Operation, error) {
	xdrOp := xdr.ManageDataOp{DataName: xdr.String64(md.Name)}

	// No data value clears the named data entry on the account
	if md.Value == nil {
		xdrOp.DataValue = nil
	} else {
		xdrDV := xdr.DataValue(md.Value)
		xdrOp.DataValue = &xdrDV
	}

	opType := xdr.OperationTypeManageData
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}
	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, md.SourceAccount)
	return op, nil
}
