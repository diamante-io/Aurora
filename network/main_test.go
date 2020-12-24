package network

import (
	"testing"

	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashTransaction(t *testing.T) {
	var txe xdr.TransactionEnvelope

	err := xdr.SafeUnmarshalBase64("AAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3AAAACgAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAArqN6LeOagjxMaUP96Bzfs9e0corNZXzBWJkFoK7kvkwAAAAAO5rKAAAAAAAAAAABVvwF9wAAAEAKZ7IPj/46PuWU6ZOtyMosctNAkXRNX9WCAI5RnfRk+AyxDLoDZP/9l3NvsxQtWj9juQOuoBlFLnWu8intgxQA", &txe)

	require.NoError(t, err)

	expected := [32]byte{
		0xc4, 0x92, 0xd8, 0x7c, 0x46, 0x42, 0x81, 0x5d,
		0xfb, 0x3c, 0x7d, 0xcc, 0xe0, 0x1a, 0xf4, 0xef,
		0xfd, 0x16, 0x2b, 0x03, 0x10, 0x64, 0x09, 0x8a,
		0x0d, 0x78, 0x6b, 0x6e, 0x0a, 0x00, 0xfd, 0x74,
	}

	actual, err := HashTransaction(&txe.Tx, TestNetworkPassphrase)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	// sadpath: empty passphrase
	_, err = HashTransaction(&txe.Tx, "")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "empty network passphrase")
	}
}
