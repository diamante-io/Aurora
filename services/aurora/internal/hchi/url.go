package hchi

import (
	"errors"
	"net/http"
	"net/url"
	"unicode/utf8"

	"github.com/go-chi/chi"
	"github.com/hcnet/go/support/render/problem"
)

func GetStringFromURL(r *http.Request, key string) (string, error) {
	val, ok := getChiURLParam(r, key)
	if ok {
		unescapedVal, err := url.PathUnescape(val)
		if err != nil {
			return "", problem.MakeInvalidFieldProblem(key, err)
		}

		return unescapedVal, checkUTF8(key, unescapedVal)
	}

	val = r.FormValue(key)
	if val != "" {
		return val, checkUTF8(key, val)
	}

	val = r.URL.Query().Get(key)
	return val, checkUTF8(key, val)
}

func checkUTF8(key, val string) error {
	if !utf8.ValidString(val) {
		// TODO: add errInvalidValue
		return problem.MakeInvalidFieldProblem(key, errors.New("invalid value"))
	}
	return nil
}

func getChiURLParam(r *http.Request, key string) (string, bool) {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		for k := len(rctx.URLParams.Keys) - 1; k >= 0; k-- {
			if rctx.URLParams.Keys[k] == key {
				return rctx.URLParams.Values[k], true
			}
		}
	}

	return "", false
}
