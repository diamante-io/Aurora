package compliance

import (
	"encoding/json"
)

// Map transforms SenderInfo to map[string]string for embedding in
// Transaction/Operation struct.
func (senderInfo SenderInfo) Map() (map[string]string, error) {
	bytes, err := json.Marshal(senderInfo)
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
