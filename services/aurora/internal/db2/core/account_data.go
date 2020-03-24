package core

import (
	"encoding/base64"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/hcnet/go/support/errors"
)

// Raw returns the decoded, raw value of the account data
func (ad AccountData) Raw() ([]byte, error) {
	return base64.StdEncoding.DecodeString(ad.Value)
}

// AccountDataByKey loads a row from `accountdata`, by key
func (q *Q) AccountDataByKey(dest *AccountData, addy string, key string) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	queryKey := key
	if schemaVersion >= 9 {
		// Since schema version 9, keys are base64 encoded.
		queryKey = base64.StdEncoding.EncodeToString([]byte(key))
	}

	sql := selectAccountData.Limit(1).
		Where("accountid = ?", addy).
		Where("dataname = ?", queryKey)

	err = q.Get(dest, sql)
	if err != nil {
		return err
	}

	if schemaVersion >= 9 {
		dest.Key = key
	}

	return nil
}

// AllDataByAddress loads all data for `addy`
func (q *Q) AllDataByAddress(dest interface{}, addy string) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	sql := selectAccountData.Where("accountid = ?", addy)
	err = q.Select(dest, sql)
	if err != nil {
		return err
	}

	if schemaVersion >= 9 {
		// Since schema version 9, keys are base64 encoded.
		d, ok := dest.(*[]AccountData)
		if !ok {
			return errors.New("Cannot ensure []AccountData type")
		}

		for i, val := range *d {
			decoded, err := base64.StdEncoding.DecodeString(val.Key)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error decoding data entry: %s", val.Key))
			}
			(*d)[i].Key = string(decoded)
		}
	}

	return nil
}

var selectAccountData = sq.Select(
	"ad.accountid",
	"ad.dataname",
	"ad.datavalue",
).From("accountdata ad")
