package test

import (
	"fmt"

	"github.com/go-errors/errors"
)

func ShouldBeErr(a interface{}, options ...interface{}) string {
	actual := a.(error)
	expected := options[0].(error)
	ok := errors.Is(actual, expected)

	if !ok {
		return fmt.Sprintf("Errors don't match:\n%v+\n%v+", actual, expected)
	}

	return ""
}
