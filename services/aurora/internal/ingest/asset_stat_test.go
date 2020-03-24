package ingest

import (
	"fmt"
	"sort"
	"testing"

	"github.com/hcnet/go/keypair"
	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestStatTrustlinesInfo(t *testing.T) {
	type AssetState struct {
		assetType       xdr.AssetType
		assetCode       string
		assetIssuer     string
		wantNumAccounts int32
		wantAmount      string
	}

	testCases := []struct {
		scenario   string
		assetState []AssetState
	}{
		{
			"asset_stat_trustlines_1",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}},
		}, {
			"asset_stat_trustlines_2",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}},
		}, {
			"asset_stat_trustlines_3",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD1",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}, {
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD2",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}},
		}, {
			"asset_stat_trustlines_4",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}, {
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
				1,
				"0",
			}},
		}, {
			"asset_stat_trustlines_5",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"0",
			}},
		}, {
			"asset_stat_trustlines_6",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1,
				"1012345000",
			}},
		}, {
			"asset_stat_trustlines_7",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				2,
				"1012345000",
			}},
		}, {
			"allow_trust",
			[]AssetState{{
				xdr.AssetTypeAssetTypeCreditAlphanum4,
				"USD",
				"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
				1, // assets with the auth_required flag should only be counted for authorized accounts
				"0",
			}},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.scenario, func(t *testing.T) {
			tt := test.Start(t).ScenarioWithoutAurora(kase.scenario)
			defer tt.Finish()

			session := &db.Session{DB: tt.CoreDB}

			for i, asset := range kase.assetState {
				numAccounts, amount, err := statTrustlinesInfo(session, asset.assetType, asset.assetCode, asset.assetIssuer)

				tt.Require.NoError(err)
				tt.Assert.Equal(asset.wantNumAccounts, numAccounts, fmt.Sprintf("asset index: %d", i))
				tt.Assert.Equal(asset.wantAmount, amount, fmt.Sprintf("asset index: %d", i))
			}
		})
	}
}

func TestStatAccountInfo(t *testing.T) {
	testCases := []struct {
		account   string
		wantFlags int8
		wantToml  string
	}{
		{
			"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			0,
			"https://example.com/.well-known/hcnet.toml",
		}, {
			"GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2",
			1,
			"https://abc.com/.well-known/hcnet.toml",
		}, {
			"GBXGQJWVLWOYHFLVTKWV5FGHA3LNYY2JQKM7OAJAUEQFU6LPCSEFVXON",
			2,
			"",
		}, {
			"GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4",
			3,
			"",
		},
	}

	for _, kase := range testCases {
		t.Run(kase.account, func(t *testing.T) {
			tt := test.Start(t).ScenarioWithoutAurora("asset_stat_account")
			defer tt.Finish()

			session := &db.Session{DB: tt.CoreDB}

			flags, toml, err := statAccountInfo(session, kase.account)
			tt.Require.NoError(err)
			tt.Assert.Equal(kase.wantFlags, flags)
			tt.Assert.Equal(kase.wantToml, toml)
		})
	}
}

func TestAssetModified(t *testing.T) {
	// GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5
	sourceAccount, sourceUSD := makeAccount("SANFNPZPA4LWBD3RPDSCJU63KCBU3OBFOM5FFBJCGIOCVIABMRTKBAU2", "USD")
	// GCSX4PDUZP3BL522ZVMFXCEJ55NKEOHEMII7PSMJZNAAESJ444GSSJMO
	destAccount, destEUR := makeAccount("SABP5P625YBETJV4BCEWQD674ED4FF4QVNBRL6TQCRODJUWBNJBMND5O", "EUR")
	// GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3
	issuerAccount, issuerUSD := makeAccount("SCCUFFUANIXJPAWBHDXZXY5D4GB32QPM6MOUWDD6PTYBLPE6JVYZFE76", "USD")
	// GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN
	anotherAccount, anotherUSD := makeAccount("SAISD7SISIIW5YNQ7GY5727L6MOFS667K3LVIPYPPUBIPCRQUORFLQMN", "USD")

	testCases := []struct {
		opBody     xdr.OperationBody
		needsCoreQ bool
		wantAssets []string
	}{
		{
			opBody: makeOperationBody(xdr.OperationTypeCreateAccount, xdr.CreateAccountOp{
				Destination:     destAccount,
				StartingBalance: 1000,
			}),
			wantAssets: []string{},
		}, {
			opBody: makeOperationBody(xdr.OperationTypePayment, xdr.PaymentOp{
				Destination: destAccount,
				Asset:       issuerUSD,
				Amount:      100,
			}),
			wantAssets: []string{},
		}, {
			// payments is the only operation where we currently perform the optimization of checking against the issuer
			opBody: makeOperationBody(xdr.OperationTypePayment, xdr.PaymentOp{
				Destination: issuerAccount,
				Asset:       issuerUSD,
				Amount:      100,
			}),
			wantAssets: []string{"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3"}, // issuerUSD
		}, {
			// payments is the only operation where we currently perform the optimization of checking against the issuer
			opBody: makeOperationBody(xdr.OperationTypePayment, xdr.PaymentOp{
				Destination: issuerAccount,
				Asset:       sourceUSD,
				Amount:      100,
			}),
			wantAssets: []string{"credit_alphanum4/USD/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5"}, // sourceUSD
		}, {
			opBody: makeOperationBody(xdr.OperationTypePathPayment, xdr.PathPaymentOp{
				SendAsset:   issuerUSD,
				SendMax:     1000000,
				Destination: destAccount,
				DestAsset:   anotherUSD,
				DestAmount:  100,
				Path:        []xdr.Asset{issuerUSD, destEUR, anotherUSD},
			}),
			wantAssets: []string{
				"credit_alphanum4/EUR/GCSX4PDUZP3BL522ZVMFXCEJ55NKEOHEMII7PSMJZNAAESJ444GSSJMO", // destEUR
				"credit_alphanum4/USD/GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN", // anotherUSD
				"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3", // issuerUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeManageSellOffer, xdr.ManageSellOfferOp{
				Selling: sourceUSD,
				Buying:  anotherUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
				OfferId: 1012,
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN", // anotherUSD
				"credit_alphanum4/USD/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5", // sourceUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeManageSellOffer, xdr.ManageSellOfferOp{
				Selling: issuerUSD,
				Buying:  sourceUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
				OfferId: 1012,
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3", // issuerUSD
				"credit_alphanum4/USD/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5", // sourceUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeManageSellOffer, xdr.ManageSellOfferOp{
				Selling: issuerUSD,
				Buying:  anotherUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
				OfferId: 1012,
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN", // anotherUSD
				"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3", // issuerUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeCreatePassiveSellOffer, xdr.CreatePassiveSellOfferOp{
				Selling: sourceUSD,
				Buying:  anotherUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN", // anotherUSD
				"credit_alphanum4/USD/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5", // sourceUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeCreatePassiveSellOffer, xdr.CreatePassiveSellOfferOp{
				Selling: issuerUSD,
				Buying:  sourceUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3", // issuerUSD
				"credit_alphanum4/USD/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5", // sourceUSD
			},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeCreatePassiveSellOffer, xdr.CreatePassiveSellOfferOp{
				Selling: issuerUSD,
				Buying:  anotherUSD,
				Amount:  1000000,
				Price:   xdr.Price{N: 1, D: 2},
			}),
			wantAssets: []string{
				"credit_alphanum4/USD/GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN", // anotherUSD
				"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3", // issuerUSD
			},
			// }, {
			// TODO NNS 2
			// 	opBody: makeOperationBody(xdr.OperationTypeSetOptions, xdr.SetOptionsOp{
			// 		InflationDest: &destAccount,
			// 	}),
			// 	needsCoreQ: true,
			// 	// the source account trusts issuerUSD in asset_stat_operations.rb
			// 	wantAssets: []string{"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3"}, // issuerUSD
		}, {
			opBody: makeOperationBody(xdr.OperationTypeChangeTrust, xdr.ChangeTrustOp{
				Line:  issuerUSD,
				Limit: 400000,
			}),
			wantAssets: []string{"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3"}, // issuerUSD
		}, {
			opBody: makeOperationBody(xdr.OperationTypeChangeTrust, xdr.ChangeTrustOp{
				Line:  issuerUSD,
				Limit: 0,
			}),
			wantAssets: []string{"credit_alphanum4/USD/GCFZWN3AOVFQM2BZTZX7P47WSI4QMGJC62LILPKODTNDLVKZZNA5BQJ3"}, // issuerUSD
		}, {
			opBody: makeOperationBody(xdr.OperationTypeAllowTrust, xdr.AllowTrustOp{
				Trustor: anotherAccount,
				Asset: xdr.AllowTrustOpAsset{
					Type:       xdr.AssetTypeAssetTypeCreditAlphanum4,
					AssetCode4: makeCodeBytes("CAT"),
				},
				Authorize: true,
			}),
			wantAssets: []string{"credit_alphanum4/CAT/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5"}, // issued by anotherAccount
		}, {
			opBody:     makeOperationBody(xdr.OperationTypeAccountMerge, destAccount),
			needsCoreQ: true,
			// account merge can only happen on accounts that don't trust any assets
			wantAssets: []string{},
		}, {
			opBody:     makeOperationBody(xdr.OperationTypeInflation, nil),
			wantAssets: []string{},
		}, {
			opBody: makeOperationBody(xdr.OperationTypeManageData, xdr.ManageDataOp{
				DataName:  "someKey",
				DataValue: nil,
			}),
			wantAssets: []string{},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.opBody.Type.String(), func(t *testing.T) {
			var session *db.Session
			if kase.needsCoreQ {
				tt := test.Start(t).ScenarioWithoutAurora("asset_stat_operations")
				defer tt.Finish()
				session = &db.Session{DB: tt.CoreDB}
			}

			assetsStats := AssetStats{CoreSession: session}
			assetsStats.IngestOperation(
				&xdr.Operation{
					SourceAccount: &sourceAccount,
					Body:          kase.opBody,
				},
				&sourceAccount)
			assert.Equal(t, kase.wantAssets, extractKeys(assetsStats.toUpdate))
		})
	}
}

func TestSourceAccountForAllowTrust(t *testing.T) {
	// GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5
	sourceAccount, _ := makeAccount("SANFNPZPA4LWBD3RPDSCJU63KCBU3OBFOM5FFBJCGIOCVIABMRTKBAU2", "USD")
	// GAB7GMQPJ5YY2E4UJMLNAZPDEUKPK4AAIPRXIZHKZGUIRC6FP2LAQSDN
	anotherAccount, _ := makeAccount("SAISD7SISIIW5YNQ7GY5727L6MOFS667K3LVIPYPPUBIPCRQUORFLQMN", "USD")

	opBody := makeOperationBody(xdr.OperationTypeAllowTrust, xdr.AllowTrustOp{
		Trustor: anotherAccount,
		Asset: xdr.AllowTrustOpAsset{
			Type:       xdr.AssetTypeAssetTypeCreditAlphanum4,
			AssetCode4: makeCodeBytes("CAT"),
		},
		Authorize: true,
	})
	wantAssets := []string{"credit_alphanum4/CAT/GCYLTPOU7IVYHHA3XKQF4YB4W4ZWHFERMOQ7K47IWANKNBFBNJJNEOG5"} // issued by anotherAccount

	assetsStats := AssetStats{}
	assetsStats.IngestOperation(
		&xdr.Operation{
			// this is the difference between this test and the table-driven case above
			SourceAccount: nil,
			Body:          opBody,
		},
		&sourceAccount)
	assert.Equal(t, wantAssets, extractKeys(assetsStats.toUpdate))
}

func makeAccount(secret string, code string) (xdr.AccountId, xdr.Asset) {
	kp := keypair.MustParse(secret)

	var accountId xdr.AccountId
	err := accountId.SetAddress(kp.Address())
	if err != nil {
		panic(err)
	}

	codeBytes := makeCodeBytes(code)
	asset, err := xdr.NewAsset(xdr.AssetTypeAssetTypeCreditAlphanum4, xdr.AssetAlphaNum4{
		AssetCode: *codeBytes,
		Issuer:    accountId,
	})
	if err != nil {
		panic(err)
	}

	return accountId, asset
}

func makeCodeBytes(code string) *[4]byte {
	codeBytes := [4]byte{}
	copy(codeBytes[:], []byte(code))
	return &codeBytes
}

func makeOperationBody(aType xdr.OperationType, value interface{}) xdr.OperationBody {
	body, err := xdr.NewOperationBody(aType, value)
	if err != nil {
		panic(err)
	}
	return body
}

func extractKeys(m map[string]xdr.Asset) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
