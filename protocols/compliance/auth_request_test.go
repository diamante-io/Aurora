package compliance

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/diamnet/go/build"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulate(t *testing.T) {
	request := &http.Request{
		PostForm: url.Values{
			"data": []string{"data"},
			"sig":  []string{"sig"},
		},
	}

	authRequest := &AuthRequest{}
	authRequest.Populate(request)

	assert.Equal(t, "data", authRequest.DataJSON)
	assert.Equal(t, "sig", authRequest.Signature)
}

func TestToURLValues(t *testing.T) {
	request := &http.Request{
		PostForm: url.Values{
			"data": []string{`{"hello": "world"}`},
			"sig":  []string{"si/g="},
		},
	}

	authRequest := &AuthRequest{}
	authRequest.Populate(request)

	assert.Equal(t, `data=%7B%22hello%22%3A+%22world%22%7D&sig=si%2Fg%3D`, authRequest.ToURLValues().Encode())
}

func TestValidateSuccess(t *testing.T) {
	attachment := Attachment{
		Transaction: Transaction{
			SenderInfo: map[string]string{
				"first_name": "Bartek",
			},
			Route: "jed*diamnet.org",
		},
	}
	attachment.GenerateNonce()

	attachHash, err := attachment.Hash()
	require.NoError(t, err)
	attachMarshalled, err := attachment.Marshal()
	require.NoError(t, err)

	txBuilder, err := build.Transaction(
		build.SourceAccount{"GAW77Z6GPWXSODJOMF5L5BMX6VMYGEJRKUNBC2CZ725JTQZORK74HQQD"},
		build.Sequence{0},
		build.TestNetwork,
		build.MemoHash{attachHash},
		build.Payment(
			build.Destination{"GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE"},
			build.CreditAmount{"USD", "GAMVF7G4GJC4A7JMFJWLUAEIBFQD5RT3DCB5DC5TJDEKQBBACQ4JZVEE", "20"},
		),
	)
	require.NoError(t, err)

	txB64, err := xdr.MarshalBase64(txBuilder.TX)
	require.NoError(t, err)

	authData := &AuthData{
		Sender:         "bartek*diamnet.org",
		NeedInfo:       false,
		Tx:             txB64,
		AttachmentJSON: string(attachMarshalled),
	}

	dataJSON, err := authData.Marshal()
	require.NoError(t, err)

	authRequest := &AuthRequest{
		DataJSON:  string(dataJSON),
		Signature: "test",
	}

	assert.NoError(t, authRequest.Validate())
}

func TestValidateError(t *testing.T) {
	authRequest := &AuthRequest{
		DataJSON:  "",
		Signature: "test",
	}

	assert.EqualError(t, authRequest.Validate(), "DataJSON: non zero value required")

	authData := &AuthData{
		Sender:         "bartekdiamnet.org",
		NeedInfo:       false,
		Tx:             "&^%",
		AttachmentJSON: "abc",
	}

	assert.EqualError(t, authData.Validate(), "sender: bartekdiamnet.org does not validate as diamnet_address;tx: &^% does not validate as base64;attachment: abc does not validate as json")
}

func TestData(t *testing.T) {
	authRequest := &AuthRequest{
		DataJSON: `{"sender": "sender", "need_info": true, "tx": "tx", "attachment": "attachment"}`,
	}

	authData, err := authRequest.Data()
	require.NoError(t, err)
	assert.Equal(t, "sender", authData.Sender)
	assert.Equal(t, true, authData.NeedInfo)
	assert.Equal(t, "tx", authData.Tx)
	assert.Equal(t, "attachment", authData.AttachmentJSON)
}
