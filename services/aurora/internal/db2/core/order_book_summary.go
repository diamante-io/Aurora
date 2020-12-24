package core

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/go-errors/errors"
	"github.com/diamnet/go/xdr"
)

type orderbookQueryBuilder struct {
	SellingType   xdr.AssetType
	SellingCode   string
	SellingIssuer string
	BuyingType    xdr.AssetType
	BuyingCode    string
	BuyingIssuer  string
	args          []interface{}
}

var orderbookQueryTemplateSchema8 *template.Template
var orderbookQuerySchema9 string

// Asks filters the summary into a slice of PriceLevelRecords where the type is 'ask'
func (o *OrderBookSummary) Asks() []OrderBookSummaryPriceLevel {
	return o.filter("ask", false)
}

// Bids filters the summary into a slice of PriceLevelRecords where the type is 'bid'
func (o *OrderBookSummary) Bids() []OrderBookSummaryPriceLevel {
	return o.filter("bid", true)
}

func (o *OrderBookSummary) filter(typ string, prepend bool) []OrderBookSummaryPriceLevel {
	result := []OrderBookSummaryPriceLevel{}

	for _, r := range *o {
		if r.Type != typ {
			continue
		}

		if prepend {
			head := []OrderBookSummaryPriceLevel{r}
			result = append(head, result...)
		} else {
			result = append(result, r)
		}
	}

	return result
}

// GetOrderBookSummary loads a summary of an order book identified by a
// selling/buying pair. It is designed to drive an order book summary client
// interface (bid/ask spread, prices and volume, etc).
func (q *Q) GetOrderBookSummary(dest interface{}, selling xdr.Asset, buying xdr.Asset, limit uint64) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	if schemaVersion < 9 {
		return q.getOrderBookSummarySchema8(dest, selling, buying, limit)
	} else {
		return q.getOrderBookSummarySchema9(dest, selling, buying, limit)
	}
}

func (q *Q) getOrderBookSummarySchema8(dest interface{}, selling xdr.Asset, buying xdr.Asset, limit uint64) error {
	var sql bytes.Buffer
	var oq orderbookQueryBuilder
	err := selling.Extract(&oq.SellingType, &oq.SellingCode, &oq.SellingIssuer)
	if err != nil {
		return err
	}
	err = buying.Extract(&oq.BuyingType, &oq.BuyingCode, &oq.BuyingIssuer)
	if err != nil {
		return err
	}

	oq.pushArg(limit)

	err = orderbookQueryTemplateSchema8.Execute(&sql, &oq)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	err = q.SelectRaw(dest, sql.String(), oq.args...)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

// Filter helps manage positional parameters and "IS NULL" checks for an order
// book query. An empty string will be converted into a null comparison.
func (q *orderbookQueryBuilder) Filter(col string, v interface{}) string {
	str, ok := v.(string)

	if ok && str == "" {
		return fmt.Sprintf("%s IS NULL", col)
	}

	n := q.pushArg(v)
	return fmt.Sprintf("%s = $%d", col, n)
}

// pushArg appends the provided value to this queries argument list and returns
// the placeholder position to use in a sql snippet
func (q *orderbookQueryBuilder) pushArg(v interface{}) int {
	q.args = append(q.args, v)
	return len(q.args)
}

func (q *Q) getOrderBookSummarySchema9(dest interface{}, selling xdr.Asset, buying xdr.Asset, limit uint64) error {
	var sellingXDRString, buyingXDRString string

	sellingXDRString, err := xdr.MarshalBase64(selling)
	if err != nil {
		return err
	}

	buyingXDRString, err = xdr.MarshalBase64(buying)
	if err != nil {
		return err
	}

	err = q.SelectRaw(dest, orderbookQuerySchema9, sellingXDRString, buyingXDRString, limit)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

func init() {
	orderbookQueryTemplateSchema8 = template.Must(template.New("sql").Parse(`
SELECT
	*,
	(pricen :: double precision / priced :: double precision) as pricef

FROM
((
	-- This query returns the "asks" portion of the summary, and it is very straightforward
	SELECT
		'ask' as type,
		co.pricen,
		co.priced,
		SUM(co.amount) as amount

	FROM  offers co

	WHERE 1=1
	AND   {{ .Filter "co.sellingassettype" .SellingType }}
	AND   {{ .Filter "co.sellingassetcode" .SellingCode}}
	AND   {{ .Filter "co.sellingissuer"    .SellingIssuer}}
	AND   {{ .Filter "co.buyingassettype"  .BuyingType }}
	AND   {{ .Filter "co.buyingassetcode"  .BuyingCode}}
	AND   {{ .Filter "co.buyingissuer"     .BuyingIssuer}}

	GROUP BY
		co.pricen,
		co.priced,
		co.price

	ORDER BY co.price ASC

	LIMIT $1

) UNION (
	-- This query returns the "bids" portion, inverting the where clauses
	-- and the pricen/priced.  This inversion is necessary to produce the "bid"
	-- view of a given offer (which are stored in the db as an offer to sell)
	SELECT
		'bid'  as type,
		co.priced as pricen,
		co.pricen as priced,
		SUM(co.amount) as amount

	FROM offers co

	WHERE 1=1
	AND   {{ .Filter "co.sellingassettype" .BuyingType }}
	AND   {{ .Filter "co.sellingassetcode" .BuyingCode}}
	AND   {{ .Filter "co.sellingissuer"    .BuyingIssuer}}
	AND   {{ .Filter "co.buyingassettype"  .SellingType }}
	AND   {{ .Filter "co.buyingassetcode"  .SellingCode}}
	AND   {{ .Filter "co.buyingissuer"     .SellingIssuer}}

	GROUP BY
		co.pricen,
		co.priced,
		co.price

	ORDER BY co.price ASC
	
	LIMIT $1
)) summary

ORDER BY type, pricef
`))

	orderbookQuerySchema9 = `
SELECT
	*,
	(pricen :: double precision / priced :: double precision) as pricef

FROM
((
	-- This query returns the "asks" portion of the summary, and it is very straightforward
	SELECT
		'ask' as type,
		co.pricen,
		co.priced,
		SUM(co.amount) as amount

	FROM  offers co

	WHERE 1=1
	AND   sellingasset = $1
	AND   buyingasset = $2

	GROUP BY
		co.pricen,
		co.priced,
		co.price

	ORDER BY co.price ASC

	LIMIT $3

) UNION (
	-- This query returns the "bids" portion, inverting the where clauses
	-- and the pricen/priced.  This inversion is necessary to produce the "bid"
	-- view of a given offer (which are stored in the db as an offer to sell)
	SELECT
		'bid'  as type,
		co.priced as pricen,
		co.pricen as priced,
		SUM(co.amount) as amount

	FROM offers co

	WHERE 1=1
	AND   sellingasset = $2
	AND   buyingasset = $1

	GROUP BY
		co.pricen,
		co.priced,
		co.price

	ORDER BY co.price ASC
	
	LIMIT $3
)) summary

ORDER BY type, pricef
`
}
