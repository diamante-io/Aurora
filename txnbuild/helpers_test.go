package txnbuild

import (
	"testing"

	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func newKeypair0() *keypair.Full {
	// Address: GDQNY3PBOJOKYZSRMK2S7LHHGWZIUISD4QORETLMXEWXBI7KFZZMKTL3
	return newKeypair("SBPQUZ6G4FZNWFHKUWC5BEYWF6R52E3SEP7R3GWYSM2XTKGF5LNTWW4R")
}

func newKeypair1() *keypair.Full {
	// Address: GAS4V4O2B7DW5T7IQRPEEVCRXMDZESKISR7DVIGKZQYYV3OSQ5SH5LVP
	return newKeypair("SBMSVD4KKELKGZXHBUQTIROWUAPQASDX7KEJITARP4VMZ6KLUHOGPTYW")
}

func newKeypair2() *keypair.Full {
	// Address: GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H
	return newKeypair("SBZVMB74Z76QZ3ZOY7UTDFYKMEGKW5XFJEB6PFKBF4UYSSWHG4EDH7PY")
}

func newKeypair(seed string) *keypair.Full {
	myKeypair, _ := keypair.Parse(seed)
	return myKeypair.(*keypair.Full)
}

func buildSignEncode(t *testing.T, tx Transaction, kps ...*keypair.Full) string {
	assert.NoError(t, tx.Build())
	assert.NoError(t, tx.Sign(kps...))

	txeBase64, err := tx.Base64()
	assert.NoError(t, err)

	return txeBase64
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkChallengeTx(txeBase64, anchorName string) (bool, error) {
	var txXDR xdr.TransactionEnvelope
	err := xdr.SafeUnmarshalBase64(txeBase64, &txXDR)
	if err != nil {
		return false, err
	}
	op := txXDR.Tx.Operations[0]
	if (xdr.OperationTypeManageData == op.Body.Type) && (op.Body.ManageDataOp.DataName == xdr.String64(anchorName+" auth")) {
		return true, nil
	}
	return false, errors.New("invalid challenge tx")
}
