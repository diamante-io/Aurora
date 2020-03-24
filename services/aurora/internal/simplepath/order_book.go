package simplepath

import (
	"errors"
	"fmt"
	"math"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/price"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/xdr"
)

// ErrNotEnough represents an error that occurs when pricing a trade on an
// orderbook.  This error occurs when the orderbook cannot fulfill the
// requested amount.
var ErrNotEnough = errors.New("not enough depth")

// orderbook represents a one-way orderbook that is selling you a specific asset (ob.Selling)
type orderBook struct {
	Selling xdr.Asset // the offers are selling this asset
	Buying  xdr.Asset // the offers are buying this asset
	Q       *core.Q
}

// CostToConsumeLiquidity returns the buyingAmount (ob.Buying) needed to consume the sellingAmount (ob.Selling)
func (ob *orderBook) CostToConsumeLiquidity(sellingAmount xdr.Int64) (xdr.Int64, error) {
	// load orderbook from core's db
	sql, e := ob.query()
	if e != nil {
		return 0, e
	}
	rows, e := ob.Q.Query(sql)
	if e != nil {
		return 0, e
	}
	defer rows.Close()

	// remaining is the units of ob.Selling that we want to consume
	remaining := int64(sellingAmount)
	var buyingAmount int64
	for rows.Next() {
		// load data from the row
		var offerAmount, pricen, priced, offerid int64
		e = rows.Scan(&offerAmount, &pricen, &priced, &offerid)
		if e != nil {
			return 0, e
		}

		buyingUnitsExtracted, sellingUnitsExtracted, e := price.ConvertToBuyingUnits(offerAmount, remaining, pricen, priced)
		if e != nil {
			return 0, e
		}
		// overflow check
		if willAddOverflow(buyingAmount, buyingUnitsExtracted) {
			return xdr.Int64(0), fmt.Errorf("adding these two values will cause an integer overflow: %d, %d", buyingAmount, buyingUnitsExtracted)
		}
		buyingAmount += buyingUnitsExtracted
		remaining -= sellingUnitsExtracted

		// check if we got all the units we wanted
		if remaining <= 0 {
			return xdr.Int64(buyingAmount), nil
		}
	}
	return 0, ErrNotEnough
}

func willAddOverflow(a int64, b int64) bool {
	return a > math.MaxInt64-b
}

func (ob *orderBook) query() (sq.SelectBuilder, error) {
	schemaVersion, err := ob.Q.SchemaVersion()
	if err != nil {
		return sq.SelectBuilder{}, err
	}

	if schemaVersion < 9 {
		return ob.querySchema8()
	} else {
		return ob.querySchema9()
	}
}

func (ob *orderBook) querySchema8() (sq.SelectBuilder, error) {
	var (
		// selling/buying types
		st, bt xdr.AssetType
		// selling/buying codes
		sc, bc string
		// selling/buying issuers
		si, bi string
	)
	e := ob.Selling.Extract(&st, &sc, &si)
	if e != nil {
		return sq.SelectBuilder{}, e
	}
	e = ob.Buying.Extract(&bt, &bc, &bi)
	if e != nil {
		return sq.SelectBuilder{}, e
	}

	sql := sq.
		Select("amount", "pricen", "priced", "offerid").
		From("offers").
		Where(sq.Eq{
			"sellingassettype":               st,
			"COALESCE(sellingassetcode, '')": sc,
			"COALESCE(sellingissuer, '')":    si}).
		Where(sq.Eq{
			"buyingassettype":               bt,
			"COALESCE(buyingassetcode, '')": bc,
			"COALESCE(buyingissuer, '')":    bi}).
		OrderBy("price ASC")
	return sql, nil
}

func (ob *orderBook) querySchema9() (sq.SelectBuilder, error) {
	var sellingXDRString, buyingXDRString string

	sellingXDRString, err := xdr.MarshalBase64(ob.Selling)
	if err != nil {
		return sq.SelectBuilder{}, err
	}

	buyingXDRString, err = xdr.MarshalBase64(ob.Buying)
	if err != nil {
		return sq.SelectBuilder{}, err
	}

	sql := sq.
		Select("amount", "pricen", "priced", "offerid").
		From("offers").
		Where(sq.Eq{"sellingasset": sellingXDRString}).
		Where(sq.Eq{"buyingasset": buyingXDRString}).
		OrderBy("price ASC")
	return sql, nil
}
