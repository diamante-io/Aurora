package participants

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

func TestForOperation(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()
	q := &core.Q{Session: tt.CoreSession()}

	load := func(lg int32, tx int, op int) []xdr.AccountId {
		var txs []core.Transaction

		err := q.TransactionsByLedger(&txs, lg)
		tt.Require.NoError(err, "failed to load transaction data")
		xtx := txs[tx].Envelope.Tx
		xop := xtx.Operations[op]
		ret, err := ForOperation(&xtx, &xop)
		tt.Require.NoError(err, "ForOperation() errored")
		return ret
	}

	// test create account
	p := load(3, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"))
	tt.Assert.Contains(p, aid("GAXI33UCLQTCKM2NMRBS7XYBR535LLEVAHL5YBN4FTCB4HZHT7ZA5CVK"))

	// test payment
	p = load(8, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GA46VRKBCLI2X6DXLX7AIEVRFLH3UA7XBE3NGNP6O74HQ5LXHMGTV2JB"))
	tt.Assert.Contains(p, aid("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"))

	// test path payment
	p = load(19, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GDRW375MAYR46ODGF2WGANQC2RRZL7O246DYHHCGWTV2RE7IHE2QUQLD"))
	tt.Assert.Contains(p, aid("GACAR2AEYEKITE2LKI5RMXF5MIVZ6Q7XILROGDT22O7JX4DSWFS7FDDP"))

	// test manage offer
	p = load(18, 1, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GAXMF43TGZHW3QN3REOUA2U5PW5BTARXGGYJ3JIFHW3YT6QRKRL3CPPU"))

	// test passive offer
	p = load(26, 0, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GB6GN3LJUW6JYR7EDOJ47VBH7D45M4JWHXGK6LHJRAEI5JBSN2DBQY7Q"))

	// test set options
	p = load(28, 0, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GCIFFRQKHMH6JD7CK5OI4XVCYCMNRNF6PYA7JTCR3FPHPJZQTYYFB5ES"))

	// test change trust
	p = load(22, 0, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GBOK7BOUSOWPHBANBYM6MIRYZJIDIPUYJPXHTHADF75UEVIVYWHHONQC"))

	// test allow trust
	p = load(42, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GD4SMOE3VPSF7ZR3CTEQ3P5UNTBMEJDA2GLXTHR7MMARANKKJDZ7RPGF"))
	tt.Assert.Contains(p, aid("GCVW5LCRZFP7PENXTAGOVIQXADDNUXXZJCNKF4VQB2IK7W2LPJWF73UG"))

	// test account merge
	p = load(44, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GCHPXGVDKPF5KT4CNAT7X77OXYZ7YVE4JHKFDUHCGCVWCL4K4PQ67KKZ"))
	tt.Assert.Contains(p, aid("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"))

	// test inflation
	p = load(47, 0, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"))

	// test manage data
	p = load(49, 0, 0)
	tt.Assert.Len(p, 1)
	tt.Assert.Contains(p, aid("GAYSCMKQY6EYLXOPTT6JPPOXDMVNBWITPTSZIVWW4LWARVBOTH5RTLAD"))
}

func TestForTransaction(t *testing.T) {
	tt := test.Start(t).ScenarioWithoutAurora("kahuna")
	defer tt.Finish()
	q := &core.Q{Session: tt.CoreSession()}

	load := func(lg int32, tx int, op int) []xdr.AccountId {
		var txs []core.Transaction
		var fees []core.TransactionFee
		err := q.TransactionsByLedger(&txs, lg)
		tt.Require.NoError(err, "failed to load transaction data")
		err = q.TransactionFeesByLedger(&fees, lg)
		tt.Require.NoError(err, "failed to load transaction fee data")

		xtx := txs[tx].Envelope.Tx
		meta := txs[tx].ResultMeta
		fee := fees[tx].Changes

		ret, err := ForTransaction(&xtx, &meta, &fee)
		tt.Require.NoError(err, "ForOperation() errored")
		return ret
	}

	p := load(3, 0, 0)
	tt.Require.Len(p, 2)
	tt.Assert.Contains(p, aid("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"))
	tt.Assert.Contains(p, aid("GAXI33UCLQTCKM2NMRBS7XYBR535LLEVAHL5YBN4FTCB4HZHT7ZA5CVK"))
}

// helper function to convert an address into an accountid
func aid(addy string) (ret xdr.AccountId) {
	err := ret.SetAddress(addy)
	if err != nil {
		panic(err)
	}
	return
}
