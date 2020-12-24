package actions

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-chi/chi"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/diamnet/go/services/aurora/internal/toid"
	"github.com/diamnet/go/support/render/problem"
	"github.com/diamnet/go/xdr"
)

func TestGetAccountID(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	aid := action.GetAccountID("4_asset_issuer")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(
		"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		aid.Address(),
	)
}

func TestGetAsset(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	ts := action.GetAsset("native_")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeNative, ts.Type)
	}

	ts = action.GetAsset("4_")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeCreditAlphanum4, ts.Type)
	}

	ts = action.GetAsset("12_")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeCreditAlphanum12, ts.Type)
	}

	// bad path
	action.GetAsset("cursor")
	tt.Assert.Error(action.Err)

	// regression #298:  GetAsset panics when asset_code is longer than allowes
	tt.Assert.NotPanics(func() {
		action.Err = nil
		action.GetAsset("long_4_")
	})

	tt.Assert.NotPanics(func() {
		action.Err = nil
		action.GetAsset("long_12_")
	})
}

func TestGetAssetType(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	ts := action.GetAssetType("native_asset_type")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeNative, ts)
	}

	ts = action.GetAssetType("4_asset_type")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeCreditAlphanum4, ts)
	}

	ts = action.GetAssetType("12_asset_type")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(xdr.AssetTypeAssetTypeCreditAlphanum12, ts)
	}
}
func TestGetCursor(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()

	// now uses the ledger state
	action := makeAction("/?cursor=now", nil)
	cursor := action.GetCursor("cursor")
	if tt.Assert.NoError(action.Err) {
		expected := toid.AfterLedger(ledger.CurrentState().HistoryLatest).String()
		tt.Assert.Equal(expected, cursor)
	}

	//Last-Event-ID overrides cursor
	action = makeTestAction()
	action.R.Header.Set("Last-Event-ID", "from_header")
	cursor = action.GetCursor("cursor")
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal("from_header", cursor)
	}
}

func TestGetInt32(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	result := action.GetInt32("blank")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(int32(0), result)

	result = action.GetInt32("zero")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(int32(0), result)

	result = action.GetInt32("two")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(int32(2), result)

	result = action.GetInt32("32max")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int32(math.MaxInt32), result)

	result = action.GetInt32("32min")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int32(math.MinInt32), result)

	result = action.GetInt32("limit")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int32(2), result)

	// overflows
	action.Err = nil
	_ = action.GetInt32("64max")
	if tt.Assert.IsType(&problem.P{}, action.Err) {
		p := action.Err.(*problem.P)
		tt.Assert.Equal("bad_request", p.Type)
		tt.Assert.Equal("64max", p.Extras["invalid_field"])
	}

	action.Err = nil
	_ = action.GetInt32("64min")
	if tt.Assert.IsType(&problem.P{}, action.Err) {
		p := action.Err.(*problem.P)
		tt.Assert.Equal("bad_request", p.Type)
		tt.Assert.Equal("64min", p.Extras["invalid_field"])
	}
}

func TestGetInt64(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	result := action.GetInt64("blank")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int64(0), result)

	result = action.GetInt64("zero")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int64(0), result)

	result = action.GetInt64("two")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(int64(2), result)

	result = action.GetInt64("64max")
	tt.Assert.NoError(action.Err)
	tt.Assert.EqualValues(int64(math.MaxInt64), result)

	result = action.GetInt64("64min")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(int64(math.MinInt64), result)
}

func TestAmount(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	result := action.GetAmount("minus_one")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(-10000000), result)

	result = action.GetAmount("zero")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(0), result)

	result = action.GetAmount("two")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(20000000), result)

	result = action.GetAmount("twenty")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(200000000), result)
}

func TestPositiveAmount(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	result := action.GetPositiveAmount("minus_one")
	tt.Assert.Error(action.Err)
	tt.Assert.Equal(xdr.Int64(0), result)
	action.Err = nil

	result = action.GetPositiveAmount("zero")
	tt.Assert.Error(action.Err)
	tt.Assert.Equal(xdr.Int64(0), result)
	action.Err = nil

	result = action.GetPositiveAmount("two")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(20000000), result)
	action.Err = nil

	result = action.GetPositiveAmount("twenty")
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal(xdr.Int64(200000000), result)
	action.Err = nil
}

func TestGetLimit(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()

	// happy path
	action := makeTestAction()
	limit := action.GetLimit("limit", 5, 200)
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(uint64(2), limit)
	}

	action = makeAction("/?limit=200", nil)
	limit = action.GetLimit("limit", 5, 200)
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(uint64(200), limit)
	}

	// defaults
	action = makeAction("/", nil)
	limit = action.GetLimit("limit", 5, 200)
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(uint64(5), limit)
	}

	action = makeAction("/?limit=", nil)
	limit = action.GetLimit("limit", 5, 200)
	if tt.Assert.NoError(action.Err) {
		tt.Assert.Equal(uint64(5), limit)
	}

	// invalids
	action = makeAction("/?limit=0", nil)
	_ = action.GetLimit("limit", 5, 200)
	tt.Assert.Error(action.Err)

	action = makeAction("/?limit=-1", nil)
	_ = action.GetLimit("limit", 5, 200)
	tt.Assert.Error(action.Err)

	action = makeAction("/?limit=201", nil)
	_ = action.GetLimit("limit", 5, 200)
	tt.Assert.Error(action.Err)
}

func TestGetPageQuery(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	// happy path
	pq := action.GetPageQuery()
	tt.Assert.NoError(action.Err)
	tt.Assert.Equal("123456", pq.Cursor)
	tt.Assert.Equal(uint64(2), pq.Limit)
	tt.Assert.Equal("asc", pq.Order)

	// regression: GetPagQuery does not overwrite err
	action = makeAction("/?limit=foo", nil)
	_ = action.GetLimit("limit", 1, 200)
	tt.Assert.Error(action.Err)
	_ = action.GetPageQuery()
	tt.Assert.Error(action.Err)

	// regression: https://github.com/diamnet/go/services/aurora/internal/issues/372
	// (limit of 0 turns into 10)
	makeAction("/?limit=0", nil)
	_ = action.GetPageQuery()
	tt.Assert.Error(action.Err)
}

func TestGetString(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	tt.Assert.Equal("123456", action.GetString("cursor"))
	action.R.Form = url.Values{
		"cursor": {"goodbye"},
	}
	tt.Assert.Equal("goodbye", action.GetString("cursor"))
}

func TestPath(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	tt.Assert.Equal("/foo-bar/blah", action.Path())
}

func TestGetURLParam(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	action := makeTestAction()

	val, ok := action.GetURLParam("two")
	tt.Assert.Equal("2", val)
	tt.Assert.Equal(true, ok)

	// valid empty string
	val, ok = action.GetURLParam("blank")
	tt.Assert.Equal("", val)
	tt.Assert.Equal(true, ok)

	// url param not found
	val, ok = action.GetURLParam("foobarcursor")
	tt.Assert.Equal("", val)
	tt.Assert.Equal(false, ok)
}

func makeTestAction() *Base {
	return makeAction("/foo-bar/blah?limit=2&cursor=123456", testURLParams())
}

func makeAction(path string, body map[string]string) *Base {
	rctx := chi.NewRouteContext()
	for k, v := range body {
		rctx.URLParams.Add(k, v)
	}

	r, _ := http.NewRequest("GET", path, nil)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	action := &Base{
		R: r,
	}
	return action
}

func testURLParams() map[string]string {
	return map[string]string{
		"blank":                "",
		"minus_one":            "-1",
		"zero":                 "0",
		"two":                  "2",
		"twenty":               "20",
		"32min":                fmt.Sprint(math.MinInt32),
		"32max":                fmt.Sprint(math.MaxInt32),
		"64min":                fmt.Sprint(math.MinInt64),
		"64max":                fmt.Sprint(math.MaxInt64),
		"native_asset_type":    "native",
		"4_asset_type":         "credit_alphanum4",
		"4_asset_code":         "USD",
		"4_asset_issuer":       "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		"12_asset_type":        "credit_alphanum12",
		"12_asset_code":        "USD",
		"12_asset_issuer":      "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		"long_4_asset_type":    "credit_alphanum4",
		"long_4_asset_code":    "SPOOON",
		"long_4_asset_issuer":  "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		"long_12_asset_type":   "credit_alphanum12",
		"long_12_asset_code":   "OHMYGODITSSOLONG",
		"long_12_asset_issuer": "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
	}
}
