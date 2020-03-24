package resourceadapter

import (
	"context"

	protocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
)

// Populate fills out the fields of the signer, using one of an account's
// secondary signers.
func PopulateSigner(ctx context.Context, dest *protocol.Signer, row core.Signer) {
	dest.Weight = row.Weight
	dest.Key = row.Publickey
	dest.Type = protocol.MustKeyTypeFromAddress(dest.Key)
}

// PopulateMaster fills out the fields of the signer, using a hcnet account to
// provide the data.
func PopulateMasterSigner(dest *protocol.Signer, row core.Account) {
	dest.Weight = int32(row.Thresholds[0])
	dest.Key = row.Accountid
	dest.Type = protocol.MustKeyTypeFromAddress(dest.Key)
}
