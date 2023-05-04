package history

import (
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestSequenceProviderEmptyDB(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)
	q := &Q{tt.AuroraSession()}

	addresses := []string{
		"GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB",
		"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
	}
	results, err := q.GetSequenceNumbers(tt.Ctx, addresses)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSequenceProviderGet(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)
	q := &Q{tt.AuroraSession()}

	err := q.UpsertAccounts(tt.Ctx, []AccountEntry{account1, account2})
	assert.NoError(t, err)

	results, err := q.GetSequenceNumbers(tt.Ctx, []string{
		"GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB",
		"GCT2NQM5KJJEF55NPMY444C6M6CA7T33HRNCMA6ZFBIIXKNCRO6J25K7",
		"GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, uint64(account1.SequenceNumber), results["GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB"])
	assert.Equal(t, uint64(account2.SequenceNumber), results["GCT2NQM5KJJEF55NPMY444C6M6CA7T33HRNCMA6ZFBIIXKNCRO6J25K7"])
}
