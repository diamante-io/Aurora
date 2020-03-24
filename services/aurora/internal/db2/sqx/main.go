// Package sqx contains utilities and extensions for the squirrel package which
// is used by aurora to generate sql statements.
package sqx

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// StringArray returns a sq.Expr suitable for inclusion in an insert that
// represents the Postgres-compatible array insert.  Strings are quoted.
func StringArray(input []string) interface{} {
	quoted := make([]string, len(input))

	// NOTE (scott): this is naive and janked.  We should be using
	// https://godoc.org/github.com/lib/pq#Array instead.  We should update this
	// func code when our version of pq is updated.
	for i, str := range input {
		quoted[i] = fmt.Sprintf(`"%s"`, str)
	}

	return sq.Expr(
		"?::character varying[]",
		fmt.Sprintf("{%s}", strings.Join(quoted, ",")),
	)
}
