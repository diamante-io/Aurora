package test

import (
	"encoding/json"
)

// StringToJSONMap transforms
func StringToJSONMap(value string, ignoredFields ...string) (m map[string]interface{}) {
	err := json.Unmarshal([]byte(value), &m)
	if err != nil {
		panic(err)
	}
	for _, ignoredField := range ignoredFields {
		delete(m, ignoredField)
	}
	return
}
