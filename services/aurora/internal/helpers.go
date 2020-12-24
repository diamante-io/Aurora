package aurora

import (
	"encoding/hex"

	"github.com/diamnet/go/support/errors"
)

func countNonEmpty(params ...interface{}) (int, error) {
	count := 0

	for _, param := range params {
		switch param := param.(type) {
		default:
			return 0, errors.Errorf("unexpected type %T", param)
		case int32:
			if param != int32(0) {
				count++
			}
		case int64:
			if param != int64(0) {
				count++
			}
		case string:
			if param != "" {
				count++
			}
		}
	}

	return count, nil
}

func isValidTransactionHash(hash string) bool {
	decoded, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	return len(decoded) == 32
}
