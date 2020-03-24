package federation

import "github.com/hcnet/go/support/errors"

// LookupReverseRecord implements `ReverseDriver` by performing
// `drv.LookupReverseRecordQuery` against `drv.DB` using the provided parameter
func (drv *ReverseSQLDriver) LookupReverseRecord(
	accountid string,
) (*ReverseRecord, error) {
	drv.initDB()
	var result ReverseRecord

	err := drv.db.GetRaw(&result, drv.LookupReverseRecordQuery, accountid)

	if drv.db.NoRows(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db get")
	}

	return &result, nil
}

var _ ReverseDriver = &ReverseSQLDriver{}
