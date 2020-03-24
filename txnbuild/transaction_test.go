package txnbuild

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/hcnet/go/network"
	"github.com/hcnet/go/strkey"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInflation(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(3556091187167235))

	inflation := Inflation{}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&inflation},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAMoj8AAAAEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAJAAAAAAAAAAHqLnLFAAAAQP3NHWXvzKIHB3%2BjjhHITdc%2FtBPntWYj3SoTjpON%2BdxjKqU5ohFamSHeqi5ONXkhE9Uajr5sVZXjQfUcTTzsWAA%3D&type=TransactionEnvelope
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAMoj8AAAAEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAJAAAAAAAAAAHqLnLFAAAAQP3NHWXvzKIHB3+jjhHITdc/tBPntWYj3SoTjpON+dxjKqU5ohFamSHeqi5ONXkhE9Uajr5sVZXjQfUcTTzsWAA="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestCreateAccount(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createAccount},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAiII0AAAAaAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAITg3tq8G0kvnvoIhZPMYJsY+9KVV8xAA6NxhtKxIXZUAAAAAAX14QAAAAAAAAAAAeoucsUAAABAHsyMojA0Q5MiNsR5X5AiNpCn9mlXmqluRsNpTniCR91M4U5TFmrrqVNLkU58/l+Y8hUPwidDTRSzLZKbMUL/Bw=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPayment(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639898))

	payment := Payment{
		Destination: "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H",
		Amount:      "10",
		Asset:       NativeAsset{},
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&payment},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAiII0AAAAbAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAABAAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAAAAAAAAF9eEAAAAAAAAAAAHqLnLFAAAAQNcGQpjNOFCLf9eEmobN+H8SNoDH/jMrfEFPX8kM212ST+TGfirEdXH77GJXvaWplfGKmE3B+UDwLuYLwO+KbQQ="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPaymentFailsIfNoAssetSpecified(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639898))

	payment := Payment{
		Destination: "GB7BDSZU2Y27LYNLALKKALB52WS2IZWYBDGY6EQBLEED3TJOCVMZRH7H",
		Amount:      "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&payment},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	err := tx.Build()
	expectedErrMsg := "failed to build operation *txnbuild.Payment: you must specify an asset for payment"
	require.EqualError(t, err, expectedErrMsg, "An asset is required")
}

func TestBumpSequence(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(9606132444168199))

	bumpSequence := BumpSequence{
		BumpTo: 9606132444168300,
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&bumpSequence},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAiILoAAAAIAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAALACIgugAAAGwAAAAAAAAAAdKHZH4AAABAndjSSWeACpbr0ROAEK6jw5CzHiL/rCDpa6AO05+raHDowSUJBckkwlEuCjbBoO/A06tZNRT1Per3liTQrc8fCg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAccountMerge(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484298))

	accountMerge := AccountMerge{
		Destination: "GAS4V4O2B7DW5T7IQRPEEVCRXMDZESKISR7DVIGKZQYYV3OSQ5SH5LVP",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&accountMerge},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAALAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAIAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAAAAAAHqLnLFAAAAQJ/UcOgE64+GQpwv0uXXa2jrKtFdmDsyZ6ZZ/udxryPS8cNCm2L784ixPYM4XRgkoQCdxC3YK8n5x5+CXLzrrwA="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageData(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(3556091187167235))

	manageData := ManageData{
		Name:  "Fruit preference",
		Value: []byte("Apple"),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&manageData},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	// https://www.hcnet.org/laboratory/#txsigner?xdr=AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAMoj8AAAAEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAKAAAAEEZydWl0IHByZWZlcmVuY2UAAAABAAAABUFwcGxlAAAAAAAAAAAAAAA%3D
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAMoj8AAAAEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAKAAAAEEZydWl0IHByZWZlcmVuY2UAAAABAAAABUFwcGxlAAAAAAAAAAAAAAHqLnLFAAAAQO1ELJBEoqBDyIsS7uSJwe1LOimV/E+09MyF1G/+yrxSggFVPEjD5LXcm/6POze3IsMuIYJU1et5Q2Vt9f73zQo="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageDataRemoveDataEntry(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484309))

	manageData := ManageData{
		Name: "Fruit preference",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&manageData},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAWAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAKAAAAEEZydWl0IHByZWZlcmVuY2UAAAAAAAAAAAAAAAHqLnLFAAAAQB8rkFZgtffUTdCASzwJ3jRcMCzHpVbbuFbye7Ki2dLao6u5d2aSzz3M2ugNJjNFMfSu3io9adCqwVKKjk0UJQA="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsInflationDestination(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484315))

	setOptions := SetOptions{
		InflationDestination: NewInflationDestination(kp1.Address()),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAcAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAQAAAAAlyvHaD8duz+iEXkJUUbsHkklIlH46oMrMMYrt0odkfgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHqLnLFAAAAQB0RLe9DjdHzLM22whFja3ZT97L/818lvWpk5EOTETr9lmDH7/A0/EAzeCkTBzZMCi3C6pV1PrGBr0NJdRrPowg="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsSetFlags(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484318))

	setOptions := SetOptions{
		SetFlags: []AccountFlag{AuthRequired, AuthRevocable},
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAfAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAABAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB6i5yxQAAAECfYTppxtp1A2zSbb6VzkOkyk9D/7xjaXRxR+ZIqgdK3lWkHQRkjyVBj2yaI61J3trdp7CswImptjkjLprt0WIO"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsClearFlags(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484319))

	setOptions := SetOptions{
		ClearFlags: []AccountFlag{AuthRequired, AuthRevocable},
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAgAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAEAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB6i5yxQAAAEANXPAN+RgvqjGF0kJ6MyNTiMnWaELw5vYNwxhv8+mi3KmGWMzojCxcmMAqni0zBMsEjl9z7H8JT9x05OlQ9nsD"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsMasterWeight(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484320))

	setOptions := SetOptions{
		MasterWeight: NewThreshold(10),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAhAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAQAAAAoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB6i5yxQAAAECIxH2W4XZ5fMsG658hdIEys2nlVSAK1FEjT5GADF6sWEThGFc+Wrmlw6GwKn6ZNAmxVULEgircjQx48aYSgFYD"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsThresholds(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484322))

	setOptions := SetOptions{
		LowThreshold:    NewThreshold(1),
		MediumThreshold: NewThreshold(2),
		HighThreshold:   NewThreshold(2),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAjAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAABAAAAAQAAAAIAAAABAAAAAgAAAAAAAAAAAAAAAAAAAAHqLnLFAAAAQFwRcFbzEtxoxZOtWlOQld3nURHZugNj5faEncpv0X/dcrfiQVU7k3fkTYDskiVExFiq78CBsYAr0uuvfH61IQs="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsHomeDomain(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484325))

	setOptions := SetOptions{
		HomeDomain: NewHomeDomain("LovelyLumensLookLuminous.com"),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAmAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAcTG92ZWx5THVtZW5zTG9va0x1bWlub3VzLmNvbQAAAAAAAAAAAAAAAeoucsUAAABAtC4HZzvRfyphRg5jjmz5jzBn86SANXCZS59GejRE8L1uCOxgXSEVoh1b+UetUEi7JN/n1ECBEVJrXgj0c34eBg=="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestSetOptionsHomeDomainTooLong(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484323))

	setOptions := SetOptions{
		HomeDomain: NewHomeDomain("LovelyLumensLookLuminousLately.com"),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	err := tx.Build()
	assert.Error(t, err, "A validation error was expected (home domain > 32 chars)")
}

func TestSetOptionsSigner(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484325))

	setOptions := SetOptions{
		Signer: &Signer{Address: kp1.Address(), Weight: Threshold(4)},
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAAmAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAABAAAAAAAAAAB6i5yxQAAAEBfgmUK+wNj8ROz78Sg0rQ2s7lmtvA4r5epHkqc9yoxLDr/GSkmgWneVqoKNxWF0JB9L+Gql1+f8M8p1McF4MsB"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMultipleOperations(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(9606132444168199))

	inflation := Inflation{}
	bumpSequence := BumpSequence{
		BumpTo: 9606132444168300,
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&inflation, &bumpSequence},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAyAAiILoAAAAIAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAJAAAAAAAAAAsAIiC6AAAAbAAAAAAAAAAB0odkfgAAAEDmf3Ag2Hw5NdlvzJpph4Km+aNKy8kfzS1EAhIVdKJwUnMVWhOpfdXSh/aekEVdoxXh2+ioocrxdtkWAZfS3sMF"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestChangeTrust(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484348))

	changeTrust := ChangeTrust{
		Line:  CreditAsset{"ABCD", kp1.Address()},
		Limit: "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&changeTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAAA9AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAGAAAAAUFCQ0QAAAAAJcrx2g/Hbs/ohF5CVFG7B5JJSJR+OqDKzDGK7dKHZH4AAAAABfXhAAAAAAAAAAAB6i5yxQAAAED7YSd1VdewEdtEURAYuyCy8dWbzALEf1vJn88/gCER4CNdIvojOEafJEhYhzZJhdG7oa+95UjfI9vMJO8qdWMK"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestChangeTrustNativeAssetNotAllowed(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484348))

	changeTrust := ChangeTrust{
		Line:  NativeAsset{},
		Limit: "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&changeTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	err := tx.Build()
	expectedErrMsg := "failed to build operation *txnbuild.ChangeTrust: trustline cannot be extended to a native (XLM) asset"
	require.EqualError(t, err, expectedErrMsg, "No trustlines for native assets")
}

func TestChangeTrustDeleteTrustline(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484354))

	issuedAsset := CreditAsset{"ABCD", kp1.Address()}
	removeTrust := RemoveTrustlineOp(issuedAsset)

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&removeTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAABDAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAGAAAAAUFCQ0QAAAAAJcrx2g/Hbs/ohF5CVFG7B5JJSJR+OqDKzDGK7dKHZH4AAAAAAAAAAAAAAAAAAAAB6i5yxQAAAECgd2wkK35civvf6NKpsSFDyKpdyo/cs7wL+RYfZ2BCP7eGrUUpu2GfQFtf/Hm6aBwT6nJ+dONTSPXnyp7Dq18L"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAllowTrust(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484366))

	issuedAsset := CreditAsset{"ABCD", kp1.Address()}
	allowTrust := AllowTrust{
		Trustor:   kp1.Address(),
		Type:      issuedAsset,
		Authorize: true,
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&allowTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAABPAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAHAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAUFCQ0QAAAABAAAAAAAAAAHqLnLFAAAAQGGBSKitYxpHNMaVVOE2CIylWFJgwqxjhwnIvWauSSkLapntD18G1pMahLbs8Lqcr3+cEs5WjLI4eBhy6WiJhAk="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestAllowTrustNoIssuer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(40385577484366))

	issuedAsset := CreditAsset{Code: "XYZ"}
	allowTrust := AllowTrust{
		Trustor:   kp1.Address(),
		Type:      issuedAsset,
		Authorize: true,
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&allowTrust},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp0)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAAJLsAAABPAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAHAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAVhZWgAAAAABAAAAAAAAAAHqLnLFAAAAQO8mcsi/+RObrKto8tABtN8RwUi6101FqBDTwqMQp4hNuujw+SGEFaBCYLNw/u40DHFRQoBNi6zcBKbBSg+gVwE="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferNewOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761092))

	selling := NativeAsset{}
	buying := CreditAsset{"ABCD", kp0.Address()}
	sellAmount := "100"
	price := "0.01"
	createOffer, err := CreateOfferOp(selling, buying, sellAmount, price)
	check(err)

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAAFAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAADuaygAAAAABAAAAZAAAAAAAAAAAAAAAAAAAAAHSh2R+AAAAQAmXf4BnH8bWhy+Tnxf+7zgsij7pV0b7XC4rqfYWi9ZIVUaidWPbrFhaWjiQbXYB1NKdx0XjidzkcAgMInLqDgs="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferDeleteOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761105))

	offerID := int64(2921622)
	deleteOffer, err := DeleteOfferOp(offerID)
	check(err)

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&deleteOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAASAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAFGQUtFAAAAAEEHgGTElYZi82AkGiJdSja2OBaU2aEcwwp3AY3tFJ2xAAAAAAAAAAAAAAABAAAAAQAAAAAALJSWAAAAAAAAAAHSh2R+AAAAQBSjRfpyEAIMnRQOPf1BBOx8HFC6Lm6bxxdljaegnUts8SmWJGQbZN5a8PQGzOTwGdBKBk9X9d+BIrBVc3kyyQ4="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageSellOfferUpdateOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761097))

	selling := NativeAsset{}
	buying := CreditAsset{"ABCD", kp0.Address()}
	sellAmount := "50"
	price := "0.02"
	offerID := int64(2497628)
	updateOffer, err := UpdateOfferOp(selling, buying, sellAmount, price, offerID)
	check(err)

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&updateOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAAKAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAADAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAAB3NZQAAAAABAAAAMgAAAAAAJhxcAAAAAAAAAAHSh2R+AAAAQAwqWg2C/oe/zH4D3Y7/yg5SlHqFvF6A3j6GQZ9NPh3ROqutovLyAE62+rvXxM7hqSNz1Rtx4frJaOhOabh6DAg="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestCreatePassiveSellOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761100))

	createPassiveOffer := CreatePassiveSellOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp0.Address()},
		Amount:  "10",
		Price:   "1.0",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createPassiveOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAANAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAEAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAAAX14QAAAAABAAAAAQAAAAAAAAAB0odkfgAAAEAgUD7M1UL7x2m2m26ySzcSHxIneOT7/r+s/HLsgWDj6CmpSi1GZrlvtBH+CNuegCwvW09TRZJhp7bLywkaFCoK"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestPathPayment(t *testing.T) {
	kp0 := newKeypair0()
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(187316408680450))

	abcdAsset := CreditAsset{"ABCD", kp0.Address()}
	pathPayment := PathPayment{
		SendAsset:   NativeAsset{},
		SendMax:     "10",
		Destination: kp2.Address(),
		DestAsset:   NativeAsset{},
		DestAmount:  "1",
		Path:        []Asset{abcdAsset},
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&pathPayment},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp2)
	expected := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAAql0AAAADAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAACAAAAAAAAAAAF9eEAAAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAAAAAAAAAmJaAAAAAAQAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAAAAAAAEuFVmYAAAAQF2kLUL/RoFIy1cmt+GXdWn2tDUjJYV3YwF4A82zIBhqYSO6ogOoLPNRt3w+IGCAgfR4Q9lpax+wCXWoQERHSw4="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoText(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3556099777101824))

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&BumpSequence{BumpTo: 1}},
		Memo:          MemoText("Twas brillig"),
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp2)
	// https://www.hcnet.org/laboratory/#txsigner?xdr=AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAMokEAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAABAAAADFR3YXMgYnJpbGxpZwAAAAEAAAAAAAAACwAAAAAAAAABAAAAAAAAAAA%3D&network=test
	expected := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAMokEAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAABAAAADFR3YXMgYnJpbGxpZwAAAAEAAAAAAAAACwAAAAAAAAABAAAAAAAAAAEuFVmYAAAAQILT8/7MGTmWkfjMi6Y23n2cVWs+IMY67xOskTivSZehp7wWaDXLIdCbdijmG64+Nz+fPBT9HYMqSRDcLiZYDQ0="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoID(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&BumpSequence{BumpTo: 1}},
		Memo:          MemoID(314159),
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp2)
	expected := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAMLgoAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAEyy8AAAABAAAAAAAAAAsAAAAAAAAAAQAAAAAAAAABLhVZmAAAAEA5P/V/Veh6pjXj7CnqtWDATh8II+ci1z3/zmNk374XLuVLzx7jRve59AKnPMwIPwDJ8cXwEKz8+fYOIkfEI9AJ"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoHash(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&BumpSequence{BumpTo: 1}},
		Memo:          MemoHash([32]byte{0x01}),
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp2)
	expected := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAMLgoAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAADAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAsAAAAAAAAAAQAAAAAAAAABLhVZmAAAAEAgauaUpqEGF1VeXYtkYg0I19QC3GJVrCPOqDHPIdXvGkQ9N+3Vt6yfKIN0sE/X5NuD6FhArQ3adwvZeaNDilwN"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestMemoReturn(t *testing.T) {
	kp2 := newKeypair2()
	sourceAccount := NewSimpleAccount(kp2.Address(), int64(3428320205078528))

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&BumpSequence{BumpTo: 1}},
		Memo:          MemoReturn([32]byte{0x01}),
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp2)
	expected := "AAAAAH4RyzTWNfXhqwLUoCw91aWkZtgIzY8SAVkIPc0uFVmYAAAAZAAMLgoAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAEAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAsAAAAAAAAAAQAAAAAAAAABLhVZmAAAAEAuLFTunY08pbWKompoepHdazLmr7uePUSOzA4P33+SVRKWiu+h2tngOsP8hga+wpLJXT9l/0uMQ3iziRVUrh0K"
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferNewOffer(t *testing.T) {
	kp0 := newKeypair0()
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761092))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp0.Address()},
		Amount:  "100",
		Price:   "0.01",
		OfferID: 0,
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&buyOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R%2BAAAAZAAAJWoAAAAFAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAADuaygAAAAABAAAAZAAAAAAAAAAAAAAAAAAAAAHSh2R%2BAAAAQHwuorW7BvBwJAz%2BETSteeDZ9UKhox1y1BqJLvaIkWSr5rNbOpimjWQxrUNQoy%2B%2BwmtY8tiMSv3Jbz8Dd4QTaQU%3D&type=TransactionEnvelope&network=test
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAAFAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAADuaygAAAAABAAAAZAAAAAAAAAAAAAAAAAAAAAHSh2R+AAAAQHwuorW7BvBwJAz+ETSteeDZ9UKhox1y1BqJLvaIkWSr5rNbOpimjWQxrUNQoy++wmtY8tiMSv3Jbz8Dd4QTaQU="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferDeleteOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761105))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp1.Address()},
		Amount:  "0",
		Price:   "0.01",
		OfferID: int64(2921622),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&buyOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R%2BAAAAZAAAJWoAAAASAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R%2BAAAAAAAAAAAAAAABAAAAZAAAAAAALJSWAAAAAAAAAAHSh2R%2BAAAAQItno%2BcpmUYFvxLcYVaDonTV3dmvzz%2B2SLzKRrYoXOqK8wCZjcP%2FkgzPMmXhTtF2tgQ9qb0rAIYpH9%2FrjtZPBgY%3D&type=TransactionEnvelope&network=test
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAASAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAAAAAAAAAAABAAAAZAAAAAAALJSWAAAAAAAAAAHSh2R+AAAAQItno+cpmUYFvxLcYVaDonTV3dmvzz+2SLzKRrYoXOqK8wCZjcP/kgzPMmXhTtF2tgQ9qb0rAIYpH9/rjtZPBgY="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestManageBuyOfferUpdateOffer(t *testing.T) {
	kp1 := newKeypair1()
	sourceAccount := NewSimpleAccount(kp1.Address(), int64(41137196761097))

	buyOffer := ManageBuyOffer{
		Selling: NativeAsset{},
		Buying:  CreditAsset{"ABCD", kp1.Address()},
		Amount:  "50",
		Price:   "0.02",
		OfferID: int64(2921622),
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&buyOffer},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	received := buildSignEncode(t, tx, kp1)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R%2BAAAAZAAAJWoAAAAKAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R%2BAAAAAB3NZQAAAAABAAAAMgAAAAAALJSWAAAAAAAAAAHSh2R%2BAAAAQK%2FsasTxgNqvkz3dGaDOyUgfa9UAAmUBmgiyaQU1dMlNNvTVH1D7PQKXkTooWmb6qK7Ee8vaTCFU6gGmShhA9wE%3D&type=TransactionEnvelope&network=test
	expected := "AAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAZAAAJWoAAAAKAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAMAAAAAAAAAAFBQkNEAAAAACXK8doPx27P6IReQlRRuweSSUiUfjqgyswxiu3Sh2R+AAAAAB3NZQAAAAABAAAAMgAAAAAALJSWAAAAAAAAAAHSh2R+AAAAQK/sasTxgNqvkz3dGaDOyUgfa9UAAmUBmgiyaQU1dMlNNvTVH1D7PQKXkTooWmb6qK7Ee8vaTCFU6gGmShhA9wE="
	assert.Equal(t, expected, received, "Base 64 XDR should match")
}

func TestBuildChallengeTx(t *testing.T) {
	kp0 := newKeypair0()

	// infinite timebound
	txeBase64, err := BuildChallengeTx(kp0.Seed(), kp0.Address(), "SDF", network.TestNetworkPassphrase, 0)
	assert.NoError(t, err)
	var txXDR xdr.TransactionEnvelope
	err = xdr.SafeUnmarshalBase64(txeBase64, &txXDR)
	assert.NoError(t, err)
	assert.Equal(t, xdr.SequenceNumber(0), txXDR.Tx.SeqNum, "sequence number should be 0")
	assert.Equal(t, xdr.Uint32(100), txXDR.Tx.Fee, "Fee should be 100")
	assert.Equal(t, 1, len(txXDR.Tx.Operations), "number operations should be 1")
	assert.Equal(t, xdr.TimePoint(0), xdr.TimePoint(txXDR.Tx.TimeBounds.MinTime), "Min time should be 0")
	assert.Equal(t, xdr.TimePoint(0), xdr.TimePoint(txXDR.Tx.TimeBounds.MaxTime), "Max time should be 0")
	op := txXDR.Tx.Operations[0]
	assert.Equal(t, xdr.OperationTypeManageData, op.Body.Type, "operation type should be manage data")
	assert.Equal(t, xdr.String64("SDF auth"), op.Body.ManageDataOp.DataName, "DataName should be 'SDF auth'")
	assert.Equal(t, 64, len(*op.Body.ManageDataOp.DataValue), "DataValue should be 64 bytes")

	// 5 minutes timebound
	txeBase64, err = BuildChallengeTx(kp0.Seed(), kp0.Address(), "SDF1", network.TestNetworkPassphrase, time.Duration(5*time.Minute))
	assert.NoError(t, err)

	var txXDR1 xdr.TransactionEnvelope
	err = xdr.SafeUnmarshalBase64(txeBase64, &txXDR1)
	assert.NoError(t, err)
	assert.Equal(t, xdr.SequenceNumber(0), txXDR1.Tx.SeqNum, "sequence number should be 0")
	assert.Equal(t, xdr.Uint32(100), txXDR1.Tx.Fee, "Fee should be 100")
	assert.Equal(t, 1, len(txXDR1.Tx.Operations), "number operations should be 1")

	timeDiff := txXDR1.Tx.TimeBounds.MaxTime - txXDR1.Tx.TimeBounds.MinTime
	assert.Equal(t, int64(300), int64(timeDiff), "time difference should be 300 seconds")
	op1 := txXDR1.Tx.Operations[0]
	assert.Equal(t, xdr.OperationTypeManageData, op1.Body.Type, "operation type should be manage data")
	assert.Equal(t, xdr.String64("SDF1 auth"), op1.Body.ManageDataOp.DataName, "DataName should be 'SDF1 auth'")
	assert.Equal(t, 64, len(*op1.Body.ManageDataOp.DataValue), "DataValue should be 64 bytes")
}

func TestHashHex(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createAccount},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	err := tx.Build()
	assert.NoError(t, err)

	err = tx.Sign(kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAAZAAiII0AAAAaAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAITg3tq8G0kvnvoIhZPMYJsY+9KVV8xAA6NxhtKxIXZUAAAAAAX14QAAAAAAAAAAAeoucsUAAABAHsyMojA0Q5MiNsR5X5AiNpCn9mlXmqluRsNpTniCR91M4U5TFmrrqVNLkU58/l+Y8hUPwidDTRSzLZKbMUL/Bw=="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

	hashHex, err := tx.HashHex()
	assert.NoError(t, err)
	expected = "1b3905ba8c3c0ecc68ae812f2d77f27c697195e8daf568740fc0f5662f65f759"
	assert.Equal(t, expected, hashHex, "hex encoded hash should match")

	txEnv := tx.TxEnvelope()
	assert.NotNil(t, txEnv, "transaction xdr envelope should not be nil")
	assert.IsType(t, txEnv, &xdr.TransactionEnvelope{}, "tx.TxEnvelope should return type of *xdr.TransactionEnvelope")
}

func TestTransactionFee(t *testing.T) {
	kp0 := newKeypair0()
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(9605939170639897))

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createAccount},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
	}

	txFee := tx.TransactionFee()
	assert.Equal(t, 0, txFee, "Transaction fee should match")

	err := tx.Build()
	assert.NoError(t, err)
	txFee = tx.TransactionFee()
	assert.Equal(t, 100, txFee, "Transaction fee should match")

	tx = Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createAccount},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
		BaseFee:       500,
	}
	err = tx.Build()
	assert.NoError(t, err)
	txFee = tx.TransactionFee()
	assert.Equal(t, 500, txFee, "Transaction fee should match")

	err = tx.Sign(kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)
	expected := "AAAAAODcbeFyXKxmUWK1L6znNbKKIkPkHRJNbLktcKPqLnLFAAAB9AAiII0AAAAbAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAITg3tq8G0kvnvoIhZPMYJsY+9KVV8xAA6NxhtKxIXZUAAAAAAX14QAAAAAAAAAAAeoucsUAAABAnc69aKw6dg1LlHxkIetKZu8Ou8hgbj4mICV0tiOJeuiq8DvivSlAngnD+FlVIaotmg8i3dEzBg+LcLnG9UttBQ=="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")
}

func TestPreAuthTransaction(t *testing.T) {
	// Address: GDK3YEHGI3ORGVO7ZEV2XF4SV5JU3BOKHMHPP4QFJ74ZRIIRROZ7ITOJ
	kp0 := newKeypair("SDY4PF6F6OWWERZT6OL2LVNREHUGHKALUI5W4U2JK4GAKPAC2RM43OAU")
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(4353383146192898)) // sequence number is in the future

	createAccount := CreateAccount{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
	}

	// build transaction to be submitted in the future.
	txFuture := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&createAccount},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
		BaseFee:       100,
	}

	err := txFuture.Build()
	assert.NoError(t, err)

	// save the hash of the future transaction.
	txFutureHash, err := txFuture.Hash()
	assert.NoError(t, err)

	// sign transaction without keypairs, the hash of the future transaction on the account
	//  will be used for authorisation.
	err = txFuture.Sign()
	assert.NoError(t, err)

	txeFutureB64, err := txFuture.Base64()
	assert.NoError(t, err)
	expected := "AAAAANW8EOZG3RNV38krq5eSr1NNhco7DvfyBU/5mKERi7P0AAAAZAAPd2EAAAADAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAITg3tq8G0kvnvoIhZPMYJsY+9KVV8xAA6NxhtKxIXZUAAAAAAX14QAAAAAAAAAAAA=="
	assert.Equal(t, expected, txeFutureB64, "Base 64 XDR should match")

	//encode the txFutureHash as a hcnet HashTx signer key.
	preAuth, err := strkey.Encode(strkey.VersionByteHashTx, txFutureHash[:])
	assert.NoError(t, err)

	// set sequence number to the current number.
	sourceAccount.Sequence = int64(4353383146192897)

	// add hash of future transaction as signer to account
	setOptions := SetOptions{
		Signer: &Signer{Address: preAuth, Weight: Threshold(2)},
	}

	// build a transaction to add the hash of the future transaction as a signer on the account.
	txNow := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
		BaseFee:       500,
	}
	err = txNow.Build()
	assert.NoError(t, err)
	txFee := txNow.TransactionFee()
	assert.Equal(t, 500, txFee, "Transaction fee should match")

	err = txNow.Sign(kp0)
	assert.NoError(t, err)

	txeNowB64, err := txNow.Base64()
	assert.NoError(t, err)
	expected = "AAAAANW8EOZG3RNV38krq5eSr1NNhco7DvfyBU/5mKERi7P0AAAB9AAPd2EAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAY9c66YpCPn8yMopKaNBd7gbiD2cr+aTLOaZE4whmeO1AAAAAgAAAAAAAAABEYuz9AAAAEC62tXQKDTcrB8VvOQIaI3ECV0uypBkpGNuyodnLLY27ii4QMdB4g4otYIvKY6nbWQqYYapNh6Q9dVsYfK6OHQM"
	assert.Equal(t, expected, txeNowB64, "Base 64 XDR should match")
	// Note: txeFutureB64 can be submitted to the network after txeNowB64 has been applied to the account
}

func TestHashXTransaction(t *testing.T) {
	// 256 bit preimage
	preimage := "this is a preimage for hashx transactions on the hcnet network"

	preimageHash := sha256.Sum256([]byte(preimage))

	//encode preimageHash as a hcnet HashX signer key
	hashx, err := strkey.Encode(strkey.VersionByteHashX, preimageHash[:])
	assert.NoError(t, err)

	// add hashx as signer to the account
	setOptions := SetOptions{
		Signer: &Signer{Address: hashx, Weight: Threshold(1)},
	}

	// Address: GDK3YEHGI3ORGVO7ZEV2XF4SV5JU3BOKHMHPP4QFJ74ZRIIRROZ7ITOJ
	kp0 := newKeypair("SDY4PF6F6OWWERZT6OL2LVNREHUGHKALUI5W4U2JK4GAKPAC2RM43OAU")
	sourceAccount := NewSimpleAccount(kp0.Address(), int64(4353383146192899))

	tx := Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&setOptions},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
		BaseFee:       500,
	}
	err = tx.Build()
	assert.NoError(t, err)
	txFee := tx.TransactionFee()
	assert.Equal(t, 500, txFee, "Transaction fee should match")

	err = tx.Sign(kp0)
	assert.NoError(t, err)

	txeB64, err := tx.Base64()
	assert.NoError(t, err)

	expected := "AAAAANW8EOZG3RNV38krq5eSr1NNhco7DvfyBU/5mKERi7P0AAAB9AAPd2EAAAAEAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAvslgb5oIfuISFP8FYvTqsciG1iSUerJB3Au6T2WLqMFAAAAAQAAAAAAAAABEYuz9AAAAECHBwfCbcOwFyoILLW7OZejvdbsVPEwB6z6ocAG4cRGu69vXKCrBFYD2mMdJRCeglgJgfgaFj2qshBgL8yQ14UH"
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

	// build a transaction to test hashx signer
	payment := Payment{
		Destination: "GCCOBXW2XQNUSL467IEILE6MMCNRR66SSVL4YQADUNYYNUVREF3FIV2Z",
		Amount:      "10",
		Asset:       NativeAsset{},
	}

	sourceAccount.Sequence = int64(4353383146192902)

	tx = Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []Operation{&payment},
		Timebounds:    NewInfiniteTimeout(),
		Network:       network.TestNetworkPassphrase,
		BaseFee:       100,
	}

	err = tx.Build()
	assert.NoError(t, err)

	// sign transaction with the preimage
	err = tx.SignHashX([]byte(preimage))
	assert.NoError(t, err)

	txeB64, err = tx.Base64()
	assert.NoError(t, err)
	expected = "AAAAANW8EOZG3RNV38krq5eSr1NNhco7DvfyBU/5mKERi7P0AAAAZAAPd2EAAAAHAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAABAAAAAITg3tq8G0kvnvoIhZPMYJsY+9KVV8xAA6NxhtKxIXZUAAAAAAAAAAAF9eEAAAAAAAAAAAGWLqMFAAAAQHRoaXMgaXMgYSBwcmVpbWFnZSBmb3IgaGFzaHggdHJhbnNhY3Rpb25zIG9uIHRoZSBzdGVsbGFyIG5ldHdvcms="
	assert.Equal(t, expected, txeB64, "Base 64 XDR should match")

}
