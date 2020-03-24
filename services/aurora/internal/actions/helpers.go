package actions

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/go-chi/chi"
	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/services/aurora/internal/assets"
	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/httpx"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	hProblem "github.com/hcnet/go/services/aurora/internal/render/problem"
	"github.com/hcnet/go/services/aurora/internal/toid"
	"github.com/hcnet/go/strkey"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/render/problem"
	"github.com/hcnet/go/support/time"
	"github.com/hcnet/go/xdr"
)

// TODO: move these constants to urlparam.go as we should parse the params with http handlers
// in the upper level package.
const (
	// ParamCursor is a query string param name
	ParamCursor = "cursor"
	// ParamOrder is a query string param name
	ParamOrder = "order"
	// ParamLimit is a query string param name
	ParamLimit = "limit"
)

type Opt int

const (
	// DisableCursorValidation disables cursor validation in GetPageQuery
	DisableCursorValidation Opt = iota
	// RequiredParam is used in Get* methods and defines a required parameter
	// (errors if value is empty).
	RequiredParam
)

// GetCursor retrieves a string from either the URLParams, form or query string.
// This method uses the priority (URLParams, Form, Query).
func (base *Base) GetCursor(name string) string {
	if base.Err != nil {
		return ""
	}

	cursor := base.GetString(name)
	if cursor == "now" {
		tid := toid.AfterLedger(ledger.CurrentState().HistoryLatest)
		cursor = tid.String()
	}

	if lastEventID := base.R.Header.Get("Last-Event-ID"); lastEventID != "" {
		cursor = lastEventID
	}

	// In case cursor is negative value, return InvalidField error
	cursorInt, err := strconv.Atoi(cursor)
	if err == nil && cursorInt < 0 {
		msg := fmt.Sprintf("the cursor %d is a negative number: ", cursorInt)
		base.SetInvalidField("cursor", errors.New(msg))
	}

	return cursor
}

// checkUTF8 checks if value is a valid UTF-8 string, otherwise sets
// error to `action.Err`.
func (base *Base) checkUTF8(name, value string) {
	if !utf8.ValidString(value) {
		base.SetInvalidField(name, errors.New("invalid value"))
	}
}

// GetStringFromURLParam retrieves a string from the URLParams.
func (base *Base) GetStringFromURLParam(name string) string {
	if base.Err != nil {
		return ""
	}

	fromURL, ok := base.GetURLParam(name)
	if ok {
		ret, err := url.PathUnescape(fromURL)
		if err != nil {
			base.SetInvalidField(name, err)
			return ""
		}

		base.checkUTF8(name, ret)
		return ret
	}

	return ""
}

// GetString retrieves a string from either the URLParams, form or query string.
// This method uses the priority (URLParams, Form, Query).
func (base *Base) GetString(name string) string {
	if base.Err != nil {
		return ""
	}

	fromURL, ok := base.GetURLParam(name)
	if ok {
		ret, err := url.PathUnescape(fromURL)
		if err != nil {
			base.SetInvalidField(name, err)
			return ""
		}

		base.checkUTF8(name, ret)
		return ret
	}

	fromForm := base.R.FormValue(name)
	if fromForm != "" {
		base.checkUTF8(name, fromForm)
		return fromForm
	}

	value := base.R.URL.Query().Get(name)
	base.checkUTF8(name, value)
	return value
}

// GetInt64 retrieves an int64 from the action parameter of the given name.
// Populates err if the value is not a valid int64
func (base *Base) GetInt64(name string) int64 {
	if base.Err != nil {
		return 0
	}

	asStr := base.GetString(name)
	if asStr == "" {
		return 0
	}

	asI64, err := strconv.ParseInt(asStr, 10, 64)
	if err != nil {
		base.SetInvalidField(name, errors.New("unparseable value"))
		return 0
	}

	return asI64
}

// GetInt32 retrieves an int32 from the action parameter of the given name.
// Populates err if the value is not a valid int32
func (base *Base) GetInt32(name string) int32 {
	if base.Err != nil {
		return 0
	}

	asStr := base.GetString(name)
	if asStr == "" {
		return 0
	}

	asI64, err := strconv.ParseInt(asStr, 10, 32)
	if err != nil {
		base.SetInvalidField(name, errors.New("unparseable value"))
		return 0
	}

	return int32(asI64)
}

// GetBool retrieves a bool from the query parameter for the given name.
// Populates err if the value is not a valid bool.
// Defaults to `false` in case of an empty string. WARNING, do not change
// this behaviour without checking other modules, ex. this is critical
// that failed transactions are not included (`false`) by default.
func (base *Base) GetBool(name string) bool {
	if base.Err != nil {
		return false
	}

	asStr := base.R.URL.Query().Get(name)
	if asStr == "" {
		return false
	}

	if asStr == "true" {
		return true
	} else if asStr == "false" || asStr == "" {
		return false
	} else {
		base.SetInvalidField(name, errors.New("unparseable value"))
		return false
	}
}

// GetLimit retrieves a uint64 limit from the action parameter of the given
// name. Populates err if the value is not a valid limit.  Uses the provided
// default value if the limit parameter is a blank string.
func (base *Base) GetLimit(name string, def uint64, max uint64) uint64 {
	if base.Err != nil {
		return 0
	}

	limit := base.GetString(name)
	if limit == "" {
		return def
	}

	asI64, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		base.SetInvalidField(name, errors.New("unparseable value"))
		return 0
	}

	if asI64 <= 0 {
		err = errors.New("invalid limit: non-positive value provided")
	} else if asI64 > int64(max) {
		err = errors.Errorf("invalid limit: value provided that is over limit max of %d", max)
	}

	if err != nil {
		base.SetInvalidField(name, err)
		return 0
	}

	return uint64(asI64)
}

// GetPageQuery is a helper that returns a new db.PageQuery struct initialized
// using the results from a call to GetPagingParams()
func (base *Base) GetPageQuery(opts ...Opt) db2.PageQuery {
	if base.Err != nil {
		return db2.PageQuery{}
	}

	disableCursorValidation := false
	for _, opt := range opts {
		if opt == DisableCursorValidation {
			disableCursorValidation = true
		}
	}

	cursor := base.GetCursor(ParamCursor)
	order := base.GetString(ParamOrder)
	limit := base.GetLimit(ParamLimit, db2.DefaultPageSize, db2.MaxPageSize)
	if base.Err != nil {
		return db2.PageQuery{}
	}

	r, err := db2.NewPageQuery(cursor, !disableCursorValidation, order, limit)
	if err != nil {
		if invalidFieldError, ok := err.(*db2.InvalidFieldError); ok {
			base.SetInvalidField(invalidFieldError.Name, err)
		} else {
			base.Err = problem.BadRequest
		}
	}

	return r
}

// GetAddress retrieves a hcnet address.  It confirms the value loaded is a
// valid hcnet address, setting an invalid field error if it is not.
func (base *Base) GetAddress(name string, opts ...Opt) (result string) {
	if base.Err != nil {
		return
	}

	requiredParam := false
	for _, opt := range opts {
		if opt == RequiredParam {
			requiredParam = true
		}
	}

	// We should check base.Err after this call. This is why it's better to remove base.Err.
	result = base.GetString(name)
	if result == "" && !requiredParam {
		return result
	}

	_, err := strkey.Decode(strkey.VersionByteAccountID, result)
	if err != nil {
		base.SetInvalidField(name, errors.New("invalid address"))
	}

	return result
}

// GetAccountID retireves an xdr.AccountID by attempting to decode a hcnet
// address at the provided name.
func (base *Base) GetAccountID(name string) (result xdr.AccountId) {
	if base.Err != nil {
		return
	}

	raw, err := strkey.Decode(strkey.VersionByteAccountID, base.GetString(name))
	if err != nil {
		base.SetInvalidField(name, errors.New("invalid address"))
		return
	}

	var key xdr.Uint256
	copy(key[:], raw)

	result, err = xdr.NewAccountId(xdr.PublicKeyTypePublicKeyTypeEd25519, key)
	if err != nil {
		base.SetInvalidField(name, errors.New("invalid address"))
		return
	}

	return
}

// GetAmount returns a native amount (i.e. 64-bit integer) by parsing
// the string at the provided name in accordance with the hcnet client
// conventions
func (base *Base) GetAmount(name string) (result xdr.Int64) {
	if base.Err != nil {
		return
	}

	var err error
	result, err = amount.Parse(base.GetString(name))
	if err != nil {
		base.SetInvalidField(name, errors.New("invalid amount"))
		return
	}

	return
}

// GetPositiveAmount returns a native amount (i.e. 64-bit integer) by parsing
// the string at the provided name in accordance with the hcnet client
// conventions. Renders error for negative amounts and zero.
func (base *Base) GetPositiveAmount(name string) (result xdr.Int64) {
	if base.Err != nil {
		return
	}

	result = base.GetAmount(name)
	if result <= 0 {
		base.SetInvalidField(name, errors.New("Value must be positive"))
		return xdr.Int64(0)
	}

	return
}

// GetAssetType is a helper that returns a xdr.AssetType by reading a string
func (base *Base) GetAssetType(name string) xdr.AssetType {
	if base.Err != nil {
		return xdr.AssetTypeAssetTypeNative
	}

	val := base.GetString(name)
	if base.Err != nil {
		return xdr.AssetTypeAssetTypeNative
	}

	r, err := assets.Parse(val)
	if err != nil {
		base.SetInvalidField(name, err)
	}

	return r
}

// GetAsset decodes an asset from the request fields prefixed by `prefix`.  To
// succeed, three prefixed fields must be present: asset_type, asset_code, and
// asset_issuer.
func (base *Base) GetAsset(prefix string) (result xdr.Asset) {
	if base.Err != nil {
		return
	}

	var value interface{}
	t := base.GetAssetType(prefix + "asset_type")

	switch t {
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		a := xdr.AssetAlphaNum4{}
		a.Issuer = base.GetAccountID(prefix + "asset_issuer")

		c := base.GetString(prefix + "asset_code")
		if len(c) > len(a.AssetCode) {
			base.SetInvalidField(prefix+"asset_code", errors.New("code too long"))
			return
		}

		copy(a.AssetCode[:len(c)], []byte(c))
		value = a
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		a := xdr.AssetAlphaNum12{}
		a.Issuer = base.GetAccountID(prefix + "asset_issuer")

		c := base.GetString(prefix + "asset_code")
		if len(c) > len(a.AssetCode) {
			base.SetInvalidField(prefix+"asset_code", errors.New("code too long"))
			return
		}

		copy(a.AssetCode[:len(c)], []byte(c))
		value = a
	}

	result, err := xdr.NewAsset(t, value)
	if err != nil {
		panic(err)
	}

	return
}

// MaybeGetAsset decodes an asset from the request fields as GetAsset does, but
// only if type field is populated. returns an additional boolean reflecting whether
// or not the decoding was performed
func (base *Base) MaybeGetAsset(prefix string) (xdr.Asset, bool) {
	if base.Err != nil {
		return xdr.Asset{}, false
	}

	if base.GetString(prefix+"asset_type") == "" {
		return xdr.Asset{}, false
	}

	return base.GetAsset(prefix), true
}

// GetTimeMillis retrieves a TimeMillis from the action parameter of the given name.
// Populates err if the value is not a valid TimeMillis
func (base *Base) GetTimeMillis(name string) (timeMillis time.Millis) {
	if base.Err != nil {
		return
	}

	asStr := base.GetString(name)
	if asStr == "" {
		return
	}

	timeMillis, err := time.MillisFromString(asStr)
	if err != nil {
		base.SetInvalidField(name, err)
		return
	}

	return
}

// GetURLParam returns the corresponding URL parameter value from the request
// routing context and an additional boolean reflecting whether or not the
// param was found. This is ported from Chi since the Chi version returns ""
// for params not found. This is undesirable since "" also is a valid url param.
// Ref: https://github.com/go-chi/chi/blob/d132b31857e5922a2cc7963f4fcfd8f46b3f2e97/context.go#L69
func (base *Base) GetURLParam(key string) (string, bool) {
	rctx := chi.RouteContext(base.R.Context())
	for k := len(rctx.URLParams.Keys) - 1; k >= 0; k-- {
		if rctx.URLParams.Keys[k] == key {
			return rctx.URLParams.Values[k], true
		}
	}

	return "", false
}

// SetInvalidField establishes an error response triggered by an invalid
// input field from the user.
func (base *Base) SetInvalidField(name string, reason error) {
	base.Err = problem.MakeInvalidFieldProblem(name, reason)
}

// Path returns the current action's path, as determined by the http.Request of
// this action
func (base *Base) Path() string {
	return base.R.URL.Path
}

// ValidateBodyType sets an error on the action if the requests Content-Type
//  is not `application/x-www-form-urlencoded`
func (base *Base) ValidateBodyType() {
	c := base.R.Header.Get("Content-Type")
	if c == "" {
		return
	}

	mt, _, err := mime.ParseMediaType(c)
	if err != nil {
		base.Err = err
		return
	}

	if mt != "application/x-www-form-urlencoded" && mt != "multipart/form-data" {
		base.Err = &hProblem.UnsupportedMediaType
	}
}

// fullURL returns a URL containing the information regarding the original
// request stored in the context.
func fullURL(ctx context.Context) *url.URL {
	url := httpx.BaseURL(ctx)
	r := httpx.RequestFromContext(ctx)
	if r != nil {
		url.Path = r.URL.Path
		url.RawQuery = r.URL.RawQuery
	}
	return url
}
