package compliance

import (
	"encoding/json"
	"strconv"
)

func (r *Route) UnmarshalJSON(data []byte) error {
	var number int64
	err := json.Unmarshal(data, &number)
	if err == nil {
		*r = Route(strconv.FormatInt(number, 10))
		return nil
	}

	var str string
	err = json.Unmarshal(data, &str)
	if err == nil {
		*r = Route(str)
		return nil
	}

	return err
}
