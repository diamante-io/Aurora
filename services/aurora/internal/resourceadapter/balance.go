package resourceadapter

import (
	"github.com/diamnet/go/amount"
	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/assets"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

func PopulateBalance(dest *protocol.Balance, row core.Trustline) (err error) {
	dest.Type, err = assets.String(row.Assettype)
	if err != nil {
		return errors.Wrap(err, "getting the string representation from the provided xdr asset type")
	}

	dest.Balance = amount.String(row.Balance)
	dest.BuyingLiabilities = amount.String(row.BuyingLiabilities)
	dest.SellingLiabilities = amount.String(row.SellingLiabilities)
	dest.Limit = amount.String(row.Tlimit)
	dest.Issuer = row.Issuer
	dest.Code = row.Assetcode
	dest.LastModifiedLedger = row.LastModified
	isAuthorized := row.IsAuthorized()
	dest.IsAuthorized = &isAuthorized
	return
}

func PopulateNativeBalance(dest *protocol.Balance, stroops, buyingLiabilities, sellingLiabilities xdr.Int64) (err error) {
	dest.Type, err = assets.String(xdr.AssetTypeAssetTypeNative)
	if err != nil {
		return errors.Wrap(err, "getting the string representation from the provided xdr asset type")
	}

	dest.Balance = amount.String(stroops)
	dest.BuyingLiabilities = amount.String(buyingLiabilities)
	dest.SellingLiabilities = amount.String(sellingLiabilities)
	dest.LastModifiedLedger = 0
	dest.Limit = ""
	dest.Issuer = ""
	dest.Code = ""
	dest.IsAuthorized = nil
	return
}
