package federation

import "github.com/hcnet/go/support/db"
import "github.com/hcnet/go/support/errors"

// LookupRecord implements `Driver` by performing `drv.LookupRecordQuery`
// against `drv.DB` using the provided parameters
func (drv *SQLDriver) LookupRecord(name, domain string) (*Record, error) {
	drv.initDB()
	var result Record

	err := drv.db.GetRaw(&result, drv.LookupRecordQuery, name, domain)

	if drv.db.NoRows(err) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "db get")
	}

	return &result, nil
}

var _ Driver = &SQLDriver{}

func (drv *SQLDriver) initDB() {
	drv.init.Do(func() {
		if drv.Dialect == "" {
			panic("no dialect specified")
		}

		drv.db = db.Wrap(drv.DB, drv.Dialect)
	})
}
