package history

import (
	"testing"

	"github.com/diamnet/go/services/aurora/internal/test"
)

func TestAccountQueries(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	// Test Accounts()
	acs := []Account{}
	err := q.Accounts().Select(&acs)

	if tt.Assert.NoError(err) {
		tt.Assert.Len(acs, 4)
	}
}
