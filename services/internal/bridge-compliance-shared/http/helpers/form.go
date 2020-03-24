package helpers

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/hcnet/go/support/errors"
)

// FromRequest will populate destination fields using http.Request post values.
func FromRequest(r *http.Request, destination interface{}) error {
	rvalue := reflect.ValueOf(destination).Elem()
	typ := rvalue.Type()
	for i := 0; i < rvalue.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("form")
		switch tag {
		case "":
		case "-":
			continue
		default:
			value := r.PostFormValue(tag)
			if value == "" {
				continue
			}

			switch rvalue.Field(i).Kind() {
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				rvalue.Field(i).SetBool(b)
			case reflect.String:
				rvalue.Field(i).SetString(value)
			default:
				return errors.New("Invalid value: " + value + " type for type: " + tag)
			}
		}
	}

	s, special := destination.(SpecialValuesConvertable)
	if special {
		err := s.FromRequestSpecial(r, destination)
		if err != nil {
			return errors.Wrap(err, "Error from FromRequestSpecial")
		}
	}

	return nil
}

// ToValues will create url.Values from source.
func ToValues(source interface{}) url.Values {
	values := make(url.Values)
	rvalue := reflect.ValueOf(source).Elem()
	typ := rvalue.Type()
	for i := 0; i < rvalue.NumField(); i++ {
		field := rvalue.Field(i)
		tag := typ.Field(i).Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}
		switch field.Interface().(type) {
		case bool:
			value := rvalue.Field(i).Bool()
			values.Set(tag, strconv.FormatBool(value))
		case string:
			value := rvalue.Field(i).String()
			if value == "" {
				continue
			}
			values.Set(tag, value)
		}
	}

	s, special := source.(SpecialValuesConvertable)
	if special {
		s.ToValuesSpecial(values)
	}

	return values
}
