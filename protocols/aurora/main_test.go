package aurora

import (
	"encoding/json"
	"testing"

	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

// Account Tests
// An example account to be used in all the Account tests
var exampleAccount = Account{
	Data: map[string]string{
		"test":    "aGVsbG8=",
		"invalid": "a_*&^*",
	},
	Sequence: "3002985298788353",
}

func TestAccount_IncrementSequenceNumber(t *testing.T) {
	seqNum, err := exampleAccount.IncrementSequenceNumber()

	assert.Nil(t, err)
	assert.Equal(t, "3002985298788354", exampleAccount.Sequence, "sequence number string was incremented")
	assert.Equal(t, xdr.SequenceNumber(3002985298788354), seqNum, "incremented sequence number is correct value/type")
}

func TestAccount_GetData(t *testing.T) {
	decoded, err := exampleAccount.GetData("test")
	assert.Nil(t, err)
	assert.Equal(t, string(decoded), "hello", "returns decoded value when key exists")

	decoded, err = exampleAccount.GetData("test2")
	assert.Nil(t, err)
	assert.Equal(t, len(decoded), 0, "returns empty slice if key doesn't exist")

	_, err = exampleAccount.GetData("invalid")
	assert.NotNil(t, err, "returns error slice if value is invalid")
}

func TestAccount_MustGetData(t *testing.T) {
	decoded := exampleAccount.MustGetData("test")
	assert.Equal(t, string(decoded), "hello", "returns decoded value when the key exists")

	decoded = exampleAccount.MustGetData("test2")
	assert.Equal(t, len(decoded), 0, "returns empty slice if key doesn't exist")

	assert.Panics(t, func() { exampleAccount.MustGetData("invalid") }, "panics on invalid input")
}

// Transaction Tests
func TestTransactionJSONMarshal(t *testing.T) {
	transaction := Transaction{
		ID:       "12345",
		FeePaid:  10,
		MemoType: "text",
		Memo:     "",
	}
	marshaledTransaction, marshalErr := json.Marshal(transaction)
	assert.Nil(t, marshalErr)
	var result Transaction
	json.Unmarshal(marshaledTransaction, &result)
	assert.Equal(t, result, transaction, "data matches original input")
}

func TestTransactionEmptyMemoText(t *testing.T) {
	transaction := Transaction{
		MemoType: "text",
		Memo:     "",
	}
	marshaledTransaction, marshalErr := json.Marshal(transaction)
	assert.Nil(t, marshalErr)
	var result struct {
		Memo *string
	}
	json.Unmarshal(marshaledTransaction, &result)
	assert.NotNil(t, result.Memo, "memo field is present even if input memo was empty string")
}

func TestTransactionMemoTypeNone(t *testing.T) {
	transaction := Transaction{
		MemoType: "none",
	}
	marshaledTransaction, marshalErr := json.Marshal(transaction)
	assert.Nil(t, marshalErr)
	var result struct {
		Memo *string
	}
	json.Unmarshal(marshaledTransaction, &result)
	assert.Nil(t, result.Memo, "no memo field is present when memo input type was `none`")
}
