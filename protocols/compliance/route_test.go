package compliance

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalNumber(t *testing.T) {
	marshalledTx := `{"route": 15, "note": "note", "extra": "extra"}`
	var tx Transaction
	err := json.Unmarshal([]byte(marshalledTx), &tx)
	require.Nil(t, err)
	assert.Equal(t, Route("15"), tx.Route)
}

func TestUnmarshalString(t *testing.T) {
	marshalledTx := `{"route": "15", "note": "note", "extra": "extra"}`
	var tx Transaction
	err := json.Unmarshal([]byte(marshalledTx), &tx)
	require.Nil(t, err)
	assert.Equal(t, Route("15"), tx.Route)
}

func TestUnmarshalInvalid(t *testing.T) {
	marshalledTx := `{"route": test, "note": "note", "extra": "extra"}`
	var tx Transaction
	err := json.Unmarshal([]byte(marshalledTx), &tx)
	assert.NotNil(t, err)
}

func TestMarshal(t *testing.T) {
	tx := Transaction{
		Route: "15",
	}
	bytes, err := json.Marshal(tx)
	require.Nil(t, err)
	assert.Equal(t, `{"sender_info":null,"route":"15","note":"","extra":""}`, string(bytes))
}
