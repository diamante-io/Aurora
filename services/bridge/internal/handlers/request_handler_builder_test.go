package handlers

import (
	"net/http"
	"strings"
	"testing"

	hc "github.com/hcnet/go/clients/auroraclient"
	hProtocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/bridge/internal/config"

	"github.com/hcnet/go/services/bridge/internal/mocks"
	"github.com/hcnet/go/services/bridge/internal/test"
	"github.com/hcnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
)

func TestRequestHandlerBuilder(t *testing.T) {
	c := &config.Config{NetworkPassphrase: "Test SDF Network ; September 2015"}

	mockAurora := new(hc.MockClient)
	mockHTTPClient := new(mocks.MockHTTPClient)
	mockTS := new(mocks.MockTransactionSubmitter)
	mockFederationResolver := new(mocks.MockFederationResolver)
	mockHcNettomlResolver := new(mocks.MockHcNettomlResolver)

	requestHandler := RequestHandler{
		Config:               c,
		Client:               mockHTTPClient,
		Aurora:              mockAurora,
		TransactionSubmitter: mockTS,
		FederationResolver:   mockFederationResolver,
		HcNetTomlResolver:  mockHcNettomlResolver,
	}

	testServer := httptest.NewServer(t, http.HandlerFunc(requestHandler.Builder))
	defer testServer.Close()

	// When no sequence number IS NOT supplied
	data := test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "",
		"operations": [
		{
				"type": "create_account",
				"body": {
					"destination": "GCOEGO43PFSLE4K7WRZQNRO3PIOTRLKRASP32W7DSPBF65XFT4V6PSV3",
					"starting_balance": "50"
				}
		}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// Loading sequence number
	accountRequest := hc.AccountRequest{AccountID: "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5"}

	mockAurora.On(
		"AccountDetail",
		accountRequest,
	).Return(
		hProtocol.Account{
			Sequence: "123",
		},
		nil,
	).Once()

	// it should return the correct XDR
	statusCode, response := mocks.JSONGetResponse(testServer, data)
	responseString := strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB8AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAB3NZQAAAAAAAAAAAZ%2BNtPwAAABAXJ4I9DwPZKt7yO0j8IsIsc6VLUJz%2FyZ3LK%2F%2F%2Bxpxf6Mbbl%2B7cFq7wG776h7VLmDTBrSEQOydKuR0Yup4gFqYAQ%3D%3D&type=TransactionEnvelope&network=test
	expected := test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB8AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAB3NZQAAAAAAAAAAAZ+NtPwAAABAXJ4I9DwPZKt7yO0j8IsIsc6VLUJz/yZ3LK//+xpxf6Mbbl+7cFq7wG776h7VLmDTBrSEQOydKuR0Yup4gFqYAQ=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// When sequence number IS supplied
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
		{
				"type": "create_account",
				"body": {
					"destination": "GCOEGO43PFSLE4K7WRZQNRO3PIOTRLKRASP32W7DSPBF65XFT4V6PSV3",
					"starting_balance": "50"
				}
		}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR, sequence number should not be incremented
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)
	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAB3NZQAAAAAAAAAAAZ%2BNtPwAAABAN836LQKoUHzAKLTkizVKb9PsFdf73eNOSRKKGd%2BzAB9GrnlsKDLbegtt9eqvYjQ4AHzbeqJBIGX%2FHQFXCadAAw%3D%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAB3NZQAAAAAAAAAAAZ+NtPwAAABAN836LQKoUHzAKLTkizVKb9PsFdf73eNOSRKKGd+zAB9GrnlsKDLbegtt9eqvYjQ4AHzbeqJBIGX/HQFXCadAAw=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// build payment
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "payment",
					"body": {
						"destination": "GCOEGO43PFSLE4K7WRZQNRO3PIOTRLKRASP32W7DSPBF65XFT4V6PSV3",
						"amount": "100",
						"asset": {
							"code": "USD",
							"issuer": "GACETOPHMOLSZLG5IQ3D6KQDKCAAYUYTTQHIEY6IGZE4VOBDD2YY6YAO"
						}
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAABAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAVVTRAAAAAAABEm552OXLKzdRDY%2FKgNQgAxTE5wOgmPINknKuCMesY8AAAAAO5rKAAAAAAAAAAABn420%2FAAAAEBckKXSMZ9M2sZqbv53XTGw0Mv91MntHbQpn%2FV0TtoVGHBWMLIJ8ePG7E0%2B7Dc06g%2BAUR%2FoTWaoI6WWCe5SKMAL&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAABAAAAAJxDO5t5ZLJxX7RzBsXbeh04rVEEn71b45PCX3blnyvnAAAAAVVTRAAAAAAABEm552OXLKzdRDY/KgNQgAxTE5wOgmPINknKuCMesY8AAAAAO5rKAAAAAAAAAAABn420/AAAAEBckKXSMZ9M2sZqbv53XTGw0Mv91MntHbQpn/V0TtoVGHBWMLIJ8ePG7E0+7Dc06g+AUR/oTWaoI6WWCe5SKMAL"
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Path Payment
	data = test.StringToJSONMap(`{
  "source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
  "sequence_number": "123",
  "operations": [
    {
        "type": "path_payment",
        "body": {
        	"source": "GBYLUAJBHGZMVAYCALM4ZRTAOY74NTSSULG42VKVY2EOWS5X2HFBB2VL",
        	"destination": "GCOEGO43PFSLE4K7WRZQNRO3PIOTRLKRASP32W7DSPBF65XFT4V6PSV3",
        	"destination_amount": "500",
        	"destination_asset": {
        		"code": "EUR",
        		"issuer": "GDOJMKTDLGGLROSSM5BV5MXIAQ3JZHASQFUV55WBJ45AFOUXSVVFGPTJ"
        	},
        	"send_max": "100",
        	"send_asset": {
        		"code": "USD",
        		"issuer": "GACETOPHMOLSZLG5IQ3D6KQDKCAAYUYTTQHIEY6IGZE4VOBDD2YY6YAO"
        	},
        	"path": [
        		{
	        		"code": "ABCDEFG",
	        		"issuer": "GD4RIHH2HWB4MPJN72G2VGLRPUXDODFNQG6DVU47HMSSSF3RIQ4UXALD"
	        	},
	        	{}
        	]
        }
    }
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAcLoBITmyyoMCAtnMxmB2P8bOUqLNzVVVxojrS7fRyhAAAAACAAAAAVVTRAAAAAAABEm552OXLKzdRDY%2FKgNQgAxTE5wOgmPINknKuCMesY8AAAAAO5rKAAAAAACcQzubeWSycV%2B0cwbF23odOK1RBJ%2B9W%2BOTwl925Z8r5wAAAAFFVVIAAAAAANyWKmNZjLi6UmdDXrLoBDacnBKBaV72wU86ArqXlWpTAAAAASoF8gAAAAACAAAAAkFCQ0RFRkcAAAAAAAAAAAD5FBz6PYPGPS3%2BjaqZcX0uNwytgbw60587JSkXcUQ5SwAAAAAAAAAAAAAAAZ%2BNtPwAAABA1z8yIaQPklvS08JcoyY6puiTzpG9KyCiJPRlLYUYp04xEVsktvBjVZTwy%2Bbt2JvCo03iO0e3xBS7IVVIMwW0Bg%3D%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAcLoBITmyyoMCAtnMxmB2P8bOUqLNzVVVxojrS7fRyhAAAAACAAAAAVVTRAAAAAAABEm552OXLKzdRDY/KgNQgAxTE5wOgmPINknKuCMesY8AAAAAO5rKAAAAAACcQzubeWSycV+0cwbF23odOK1RBJ+9W+OTwl925Z8r5wAAAAFFVVIAAAAAANyWKmNZjLi6UmdDXrLoBDacnBKBaV72wU86ArqXlWpTAAAAASoF8gAAAAACAAAAAkFCQ0RFRkcAAAAAAAAAAAD5FBz6PYPGPS3+jaqZcX0uNwytgbw60587JSkXcUQ5SwAAAAAAAAAAAAAAAZ+NtPwAAABA1z8yIaQPklvS08JcoyY6puiTzpG9KyCiJPRlLYUYp04xEVsktvBjVZTwy+bt2JvCo03iO0e3xBS7IVVIMwW0Bg=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Set Options
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "set_options",
					"body": {
						"inflation_dest": "GBMPZVOMJ67WQBTBCVURDKTGL4557272EGQMAJCXPSMLOE63XPLL6SVA",
						"set_flags": [1, 2],
						"clear_flags": [4],
						"master_weight": 100,
						"low_threshold": 1,
						"medium_threshold": 2,
						"high_threshold": 3,
						"home_domain": "hcnet.org",
						"signer": {
							"public_key": "GA6VMJJQM2QBPPIXK2UVTAOS4XSSSAKSCOGFQE55IMRBQR65GIVDTTQV",
							"weight": 5
						}
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAQAAAABY%2FNXMT79oBmEVaRGqZl873%2Bv6IaDAJFd8mLcT27vWvwAAAAEAAAAEAAAAAQAAAAMAAAABAAAAZAAAAAEAAAABAAAAAQAAAAIAAAABAAAAAwAAAAEAAAALc3RlbGxhci5vcmcAAAAAAQAAAAA9ViUwZqAXvRdWqVmB0uXlKQFSE4xYE71DIhhH3TIqOQAAAAUAAAAAAAAAAZ%2BNtPwAAABAHYQbLjs%2FAAekVMuuU6cxLxQYG7m396Im%2BSNiSsjSyUdgjjSKt7xzupvgudjBHM1t2akOmx1OCzmlsWCSEWzoCQ%3D%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAFAAAAAQAAAABY/NXMT79oBmEVaRGqZl873+v6IaDAJFd8mLcT27vWvwAAAAEAAAAEAAAAAQAAAAMAAAABAAAAZAAAAAEAAAABAAAAAQAAAAIAAAABAAAAAwAAAAEAAAALc3RlbGxhci5vcmcAAAAAAQAAAAA9ViUwZqAXvRdWqVmB0uXlKQFSE4xYE71DIhhH3TIqOQAAAAUAAAAAAAAAAZ+NtPwAAABAHYQbLjs/AAekVMuuU6cxLxQYG7m396Im+SNiSsjSyUdgjjSKt7xzupvgudjBHM1t2akOmx1OCzmlsWCSEWzoCQ=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Change Trust
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "change_trust",
					"body": {
						"asset": {
							"code": "USD",
							"issuer": "GCHGRVNTXAV3OXNMCSA63BUCD6AZZX6PN2542QB6GIVTXGHQ65XS35DS"
						}
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAGAAAAAVVTRAAAAAAAjmjVs7grt12sFIHthoIfgZzfz267zUA%2BMis7mPD3by1%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwAAAAAAAAABn420%2FAAAAEAr4jYveod1stUGiW9Uy99mDz5gjCJp%2FNPTu3P0uVRLEGOAcM9GyEMCvm2VnK4HAtSZiCxrCmcZGTqSd38zUBgC&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAGAAAAAVVTRAAAAAAAjmjVs7grt12sFIHthoIfgZzfz267zUA+Mis7mPD3by1//////////wAAAAAAAAABn420/AAAAEAr4jYveod1stUGiW9Uy99mDz5gjCJp/NPTu3P0uVRLEGOAcM9GyEMCvm2VnK4HAtSZiCxrCmcZGTqSd38zUBgC"
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Allow Trust
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "allow_trust",
					"body": {
						"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
						"asset_code": "USDUSD",
						"trustor": "GBLH67TQHRNRLERQEIQJDNBV2DSWPHAPP43MBIF7DVKA7X55APUNS4LL",
						"authorize": true
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
	}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAbJJLdnKV98VpVmpkPqU0JSMKLMfPsws%2FCQ3szGHhc3kAAAAHAAAAAFZ%2FfnA8WxWSMCIgkbQ10OVnnA9%2FNsCgvx1UD9%2B9A%2BjZAAAAAlVTRFVTRAAAAAAAAAAAAAEAAAAAAAAAAZ%2BNtPwAAABAWROxcSLBvH04%2BZoTD%2BYv47Xv%2Bympi9pC1pYW%2Bh9JKUI9yc49FQLWM3svhyOxF%2BmyCQvt%2Fkb7yrnlPVLlWSvACA%3D%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAbJJLdnKV98VpVmpkPqU0JSMKLMfPsws/CQ3szGHhc3kAAAAHAAAAAFZ/fnA8WxWSMCIgkbQ10OVnnA9/NsCgvx1UD9+9A+jZAAAAAlVTRFVTRAAAAAAAAAAAAAEAAAAAAAAAAZ+NtPwAAABAWROxcSLBvH04+ZoTD+Yv47Xv+ympi9pC1pYW+h9JKUI9yc49FQLWM3svhyOxF+myCQvt/kb7yrnlPVLlWSvACA=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Account Merge
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "account_merge",
					"body": {
						"destination": "GBLH67TQHRNRLERQEIQJDNBV2DSWPHAPP43MBIF7DVKA7X55APUNS4LL"
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	// https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAIAAAAAFZ%2FfnA8WxWSMCIgkbQ10OVnnA9%2FNsCgvx1UD9%2B9A%2BjZAAAAAAAAAAGfjbT8AAAAQKC1XKE3ThTPMsc%2Fda80CqkmesOhPa2lrsLLszR2VzbUF%2BsSJSPpq39CdsQj0rRPY61hakf5hv319NpsCmqYNAI%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAIAAAAAFZ/fnA8WxWSMCIgkbQ10OVnnA9/NsCgvx1UD9+9A+jZAAAAAAAAAAGfjbT8AAAAQKC1XKE3ThTPMsc/da80CqkmesOhPa2lrsLLszR2VzbUF+sSJSPpq39CdsQj0rRPY61hakf5hv319NpsCmqYNAI="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	// Build Inflation
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "inflation",
					"body": {}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	//https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAJAAAAAAAAAAGfjbT8AAAAQDukGibnqm%2B%2FGQyn0xZH4a%2FyZlUNAolyrhXzsmVeCmxs3ss24RwEST2lSqrwBLnTYbpLrHPlFT2Ye14apwJ8gA4%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAJAAAAAAAAAAGfjbT8AAAAQDukGibnqm+/GQyn0xZH4a/yZlUNAolyrhXzsmVeCmxs3ss24RwEST2lSqrwBLnTYbpLrHPlFT2Ye14apwJ8gA4="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

	//Build Manage Data
	data = test.StringToJSONMap(`{
		"source": "GBWJES3WOKK7PRLJKZVGIPVFGQSSGCRMY7H3GCZ7BEG6ZTDB4FZXTPJ5",
		"sequence_number": "123",
		"operations": [
			{
					"type": "manage_data",
					"body": {
						"name": "test_data",
						"data": "AQIDBAUG"
					}
			}
		],
		"signers": ["SABY7FRMMJWPBTKQQ2ZN43AUJQ3Z2ZAK36VYSG2SPE2ABNQXA66H5E5G"]
		}`,
	)

	// it should return the correct XDR
	statusCode, response = mocks.JSONGetResponse(testServer, data)
	responseString = strings.TrimSpace(string(response))
	assert.Equal(t, 200, statusCode)

	//https://www.hcnet.org/laboratory/#xdr-viewer?input=AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAKAAAACXRlc3RfZGF0YQAAAAAAAAEAAAAGAQIDBAUGAAAAAAAAAAAAAZ%2BNtPwAAABAHv%2BHVBeU%2B2Cz6xpleNt3sKt%2BK%2FRCEbp349iBKnXbBz3fu4vy0UiZPMDpBJF3weCryacTXP0JxH47eQUmwCm1AA%3D%3D&type=TransactionEnvelope&network=test
	expected = test.StringToJSONMap(`{
		"transaction_envelope": "AAAAAGySS3ZylffFaVZqZD6lNCUjCizHz7MLPwkN7Mxh4XN5AAAAZAAAAAAAAAB7AAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAKAAAACXRlc3RfZGF0YQAAAAAAAAEAAAAGAQIDBAUGAAAAAAAAAAAAAZ+NtPwAAABAHv+HVBeU+2Cz6xpleNt3sKt+K/RCEbp349iBKnXbBz3fu4vy0UiZPMDpBJF3weCryacTXP0JxH47eQUmwCm1AA=="
		}`)
	assert.Equal(t, expected, test.StringToJSONMap(responseString))

}
