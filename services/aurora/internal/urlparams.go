package aurora

import (
	"net/http"
	"strconv"

	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/hchi"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/toid"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/render/problem"
)

// getCursor gets the param cursor from either the request URL or the request
// header and validates it.
func getCursor(r *http.Request) (string, error) {
	cursor, err := hchi.GetStringFromURL(r, actions.ParamCursor)
	if err != nil {
		return "", errors.Wrap(err, "loading cursor from URL")
	}

	if cursor == "now" {
		cursor = toid.AfterLedger(ledger.CurrentState().HistoryLatest).String()
	}

	if lastEventID := r.Header.Get("Last-Event-ID"); lastEventID != "" {
		cursor = lastEventID
	}

	if cursor == "" {
		return "", nil
	}

	curInt64, err := strconv.ParseInt(cursor, 10, 64)
	if err != nil {
		return "", problem.MakeInvalidFieldProblem(actions.ParamCursor, errors.New("invalid int64 value"))
	}
	if curInt64 < 0 {
		return "", problem.MakeInvalidFieldProblem(actions.ParamCursor, errors.Errorf("cursor %d is a negative number: ", curInt64))
	}

	return cursor, nil
}

// getOrder gets the param order from the request URL and validates it.
func getOrder(r *http.Request) (string, error) {
	order, err := hchi.GetStringFromURL(r, actions.ParamOrder)
	if err != nil {
		return "", errors.Wrap(err, "getting param order from URL")
	}

	// Set order
	if order == "" {
		order = db2.OrderAscending
	}
	if order != db2.OrderAscending && order != db2.OrderDescending {
		return "", db2.ErrInvalidOrder
	}

	return order, nil
}

// getLimit gets the param limit from the request URL and validates it.
func getLimit(r *http.Request, defaultSize, maxSize uint64) (uint64, error) {
	limit, err := hchi.GetStringFromURL(r, actions.ParamLimit)
	if err != nil {
		return 0, errors.Wrap(err, "loading param limit from URL")
	}
	if limit == "" {
		return defaultSize, nil
	}

	limitInt64, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return 0, problem.MakeInvalidFieldProblem(actions.ParamLimit, errors.New("invalid int64 value"))
	}
	if limitInt64 <= 0 {
		return 0, problem.MakeInvalidFieldProblem(actions.ParamLimit, errors.Errorf("limit %d is a non-positive number: ", limitInt64))
	}
	if limitInt64 > int64(maxSize) {
		return 0, problem.MakeInvalidFieldProblem(actions.ParamLimit, errors.Errorf("limit %d is greater than limit max of %d", limitInt64, maxSize))
	}

	return uint64(limitInt64), nil
}

// getAccountsPageQuery gets the page query for /accounts
func getAccountsPageQuery(r *http.Request) (db2.PageQuery, error) {
	cursor, err := hchi.GetStringFromURL(r, actions.ParamCursor)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param cursor")
	}

	order, err := getOrder(r)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param order")
	}

	limit, err := getLimit(r, db2.DefaultPageSize, db2.MaxPageSize)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param limit")
	}

	return db2.PageQuery{
		Cursor: cursor,
		Order:  order,
		Limit:  limit,
	}, nil
}

// getPageQuery gets the page query and does the pair validation if
// disablePairValidation is false.
func getPageQuery(r *http.Request, disablePairValidation bool) (db2.PageQuery, error) {
	cursor, err := getCursor(r)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param cursor")
	}

	order, err := getOrder(r)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param order")
	}

	limit, err := getLimit(r, db2.DefaultPageSize, db2.MaxPageSize)
	if err != nil {
		return db2.PageQuery{}, errors.Wrap(err, "getting param limit")
	}

	pq := db2.PageQuery{
		Cursor: cursor,
		Order:  order,
		Limit:  limit,
	}
	if !disablePairValidation {
		_, _, err = pq.CursorInt64Pair(db2.DefaultPairSep)
		if err != nil {
			return db2.PageQuery{}, problem.MakeInvalidFieldProblem(actions.ParamCursor, db2.ErrInvalidCursor)
		}
	}

	return pq, nil
}

// getInt32ParamFromURL gets the int32 param with the provided key. It errors
// if the param value cannot be parsed as int32.
func getInt32ParamFromURL(r *http.Request, key string) (int32, error) {
	val, err := hchi.GetStringFromURL(r, key)
	if err != nil {
		return 0, errors.Wrapf(err, "loading %s from URL", key)
	}
	if val == "" {
		return int32(0), nil
	}

	asI64, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return int32(0), problem.MakeInvalidFieldProblem(key, errors.New("invalid int32 value"))
	}

	return int32(asI64), nil
}

// getInt64ParamFromURL gets the int64 param with the provided key. It errors
// if the param value cannot be parsed as int64.
func getInt64ParamFromURL(r *http.Request, key string) (int64, error) {
	val, err := hchi.GetStringFromURL(r, key)
	if err != nil {
		return 0, errors.Wrapf(err, "loading %s from URL", key)
	}
	if val == "" {
		return 0, nil
	}

	asI64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, problem.MakeInvalidFieldProblem(key, errors.New("invalid int64 value"))
	}

	return asI64, nil
}

// getBoolParamFromURL gets the bool param with the provided key. It errors if
// the value is not "true" or "false" or "".
func getBoolParamFromURL(r *http.Request, key string) (bool, error) {
	val, err := hchi.GetStringFromURL(r, key)
	if err != nil {
		return false, errors.Wrapf(err, "loading %s from URL", key)
	}

	if val == "true" {
		return true, nil
	}
	if val == "false" || val == "" {
		return false, nil
	}

	return false, problem.MakeInvalidFieldProblem(key, errors.New("invalid bool value"))
}
