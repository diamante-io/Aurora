package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	senderInfo := SenderInfo{
		FirstName:   "a",
		MiddleName:  "b",
		LastName:    "c",
		Address:     "d",
		City:        "e",
		Province:    "f",
		Country:     "g",
		DateOfBirth: "h",
	}

	m, err := senderInfo.Map()
	require.NoError(t, err)
	assert.Equal(t, "a", m["first_name"])
	assert.Equal(t, "b", m["middle_name"])
	assert.Equal(t, "c", m["last_name"])
	assert.Equal(t, "d", m["address"])
	assert.Equal(t, "e", m["city"])
	assert.Equal(t, "f", m["province"])
	assert.Equal(t, "g", m["country"])
	assert.Equal(t, "h", m["date_of_birth"])
	// Should not populate empty values
	_, ok := m["company_name"]
	assert.False(t, ok)
}
