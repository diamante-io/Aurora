package history

import (
	"testing"

	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/test"
)

func TestQueryEmptyAccountSigners(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	signer := "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO0"
	results, err := q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 0)
}

func TestInsertAccountSigner(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	account := "GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2"
	signer := "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4"
	weight := int32(123)
	tt.Assert.NoError(q.CreateAccountSigner(account, signer, weight))

	expected := AccountSigner{
		Account: account,
		Signer:  signer,
		Weight:  weight,
	}
	results, err := q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 1)
	tt.Assert.Equal(expected, results[0])

	weight = 321
	tt.Assert.NoError(q.CreateAccountSigner(account, signer, weight))

	expected.Weight = weight
	results, err = q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 1)
	tt.Assert.Equal(expected, results[0])
}

func TestMultipleAccountsForSigner(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	account := "GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH1"
	signer := "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO2"
	weight := int32(123)
	tt.Assert.NoError(q.CreateAccountSigner(account, signer, weight))

	anotherAccount := "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H"
	anotherWeight := int32(321)
	tt.Assert.NoError(q.CreateAccountSigner(anotherAccount, signer, anotherWeight))

	expected := []AccountSigner{
		AccountSigner{
			Account: account,
			Signer:  signer,
			Weight:  weight,
		},
		AccountSigner{
			Account: anotherAccount,
			Signer:  signer,
			Weight:  anotherWeight,
		},
	}
	results, err := q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 2)
	tt.Assert.Equal(expected, results)
}

func TestRemoveNonExistantAccountSigner(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	account := "GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH3"
	signer := "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO5"
	tt.Assert.NoError(q.RemoveAccountSigner(account, signer))
}

func TestRemoveAccountSigner(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	account := "GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH6"
	signer := "GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO7"
	weight := int32(123)
	tt.Assert.NoError(q.CreateAccountSigner(account, signer, weight))

	expected := AccountSigner{
		Account: account,
		Signer:  signer,
		Weight:  weight,
	}
	results, err := q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 1)
	tt.Assert.Equal(expected, results[0])

	tt.Assert.NoError(q.RemoveAccountSigner(account, signer))

	results, err = q.AccountsForSigner(signer, db2.PageQuery{Order: "asc", Limit: 10})
	tt.Assert.NoError(err)
	tt.Assert.Len(results, 0)
}
