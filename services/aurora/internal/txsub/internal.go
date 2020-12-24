package txsub

import (
	"context"

	"github.com/diamnet/go/build"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/xdr"
)

type envelopeInfo struct {
	Hash          string
	Sequence      uint64
	SourceAddress string
}

func extractEnvelopeInfo(ctx context.Context, env string, passphrase string) (result envelopeInfo, err error) {
	var tx xdr.TransactionEnvelope

	err = xdr.SafeUnmarshalBase64(env, &tx)

	if err != nil {
		err = &MalformedTransactionError{env}
		return
	}

	txb := build.TransactionBuilder{TX: &tx.Tx}
	txb.Mutate(build.Network{passphrase})

	result.Hash, err = txb.HashHex()
	if err != nil {
		return
	}

	result.Sequence = uint64(tx.Tx.SeqNum)

	aid := tx.Tx.SourceAccount.MustEd25519()
	result.SourceAddress, err = strkey.Encode(strkey.VersionByteAccountID, aid[:])

	return
}
