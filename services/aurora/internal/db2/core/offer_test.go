package core

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestOffersByAddress(t *testing.T) {
	tt := test.Start(t).Scenario("trades")
	defer tt.Finish()
	q := &Q{tt.CoreSession()}

	var offers []Offer

	load := func(addy, cursor, order string, limit uint64) bool {
		offers = []Offer{}
		pq, err := db2.NewPageQuery(cursor, true, order, limit)
		if !tt.Assert.NoError(err) {
			return false
		}

		err = q.OffersByAddress(&offers, addy, pq)
		if !tt.Assert.NoError(err) {
			return false
		}
		return true
	}

	// Works for native offers
	if load("GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU", "", "asc", db2.DefaultPageSize) {
		tt.Assert.Len(offers, 1)
		tt.Assert.Equal(int64(4), offers[0].OfferID)
	}

	// Filters properly
	if load("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H", "", "asc", db2.DefaultPageSize) {
		tt.Assert.Len(offers, 0)
	}

	if load("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", "", "asc", db2.DefaultPageSize) {
		tt.Assert.Len(offers, 3)
	}

	// limits properly
	if load("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", "", "asc", 2) {
		tt.Assert.Len(offers, 2)
	}

	// ordering works
	if load("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", "", "desc", db2.DefaultPageSize) {
		for i := range offers {
			// if there is no next element, break
			if i+1 == len(offers) {
				break
			}
			tt.Assert.True(offers[i].OfferID > offers[i+1].OfferID, "Results are not in order")
		}
	}

	// cursor works
	if load("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", "1", "asc", db2.DefaultPageSize) {
		tt.Assert.Len(offers, 2)
		tt.Assert.Equal(int64(2), offers[0].OfferID)
	}
	if load("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2", "3", "desc", db2.DefaultPageSize) {
		tt.Assert.Len(offers, 2)
		tt.Assert.Equal(int64(2), offers[0].OfferID)
	}
}
