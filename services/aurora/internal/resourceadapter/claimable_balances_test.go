package resourceadapter

import (
	"encoding/json"
	"testing"

	"github.com/guregu/null"
	. "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/support/test"
	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestPopulateClaimableBalance(t *testing.T) {
	tt := assert.New(t)
	ctx, _ := test.ContextWithLogBuffer()
	resource := ClaimableBalance{}

	balanceID := xdr.ClaimableBalanceId{
		Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
		V0:   &xdr.Hash{1, 2, 3},
	}
	id, err := xdr.MarshalHex(&balanceID)
	tt.NoError(err)
	claimableBalance := history.ClaimableBalance{
		BalanceID: id,
		Asset:     xdr.MustNewNativeAsset(),
		Amount:    100000000,
		Sponsor:   null.StringFrom("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
		Claimants: history.Claimants{
			{
				Destination: "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML",
				Predicate: xdr.ClaimPredicate{
					Type: xdr.ClaimPredicateTypeClaimPredicateUnconditional,
				},
			},
		},
		LastModifiedLedger: 123,
		Flags:              uint32(xdr.ClaimableBalanceFlagsClaimableBalanceClawbackEnabledFlag),
	}

	err = PopulateClaimableBalance(ctx, &resource, claimableBalance, nil)
	tt.NoError(err)

	tt.Equal("000000000102030000000000000000000000000000000000000000000000000000000000", resource.BalanceID)
	tt.Equal(claimableBalance.Asset.StringCanonical(), resource.Asset)
	tt.Equal("10.0000000", resource.Amount)
	tt.Equal("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML", resource.Sponsor)
	tt.Equal(uint32(123), resource.LastModifiedLedger)
	tt.Len(resource.Claimants, 1)
	tt.Equal("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML", resource.Claimants[0].Destination)
	tt.Equal("123-000000000102030000000000000000000000000000000000000000000000000000000000", resource.PagingToken())
	tt.True(resource.Flags.ClawbackEnabled)

	links, err := json.Marshal(resource.Links)
	tt.NoError(err)
	want := `
	{
	  "self": {
		"href": "/claimable_balances/000000000102030000000000000000000000000000000000000000000000000000000000"
	  },
	  "operations": {
		"href": "/claimable_balances/000000000102030000000000000000000000000000000000000000000000000000000000/operations{?cursor,limit,order}",
        "templated": true
	  },
	  "transactions": {
		"href": "/claimable_balances/000000000102030000000000000000000000000000000000000000000000000000000000/transactions{?cursor,limit,order}",
        "templated": true
	  }
	}
	`
	tt.JSONEq(want, string(links))

	predicate, err := json.Marshal(resource.Claimants[0].Predicate)
	tt.NoError(err)
	tt.JSONEq(`{"unconditional":true}`, string(predicate))
}
