package resourceadapter

import (
	"context"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/xdr"
)

func PopulateAsset(ctx context.Context, dest *protocol.Asset, asset xdr.Asset) error {
	return asset.Extract(&dest.Type, &dest.Code, &dest.Issuer)
}
