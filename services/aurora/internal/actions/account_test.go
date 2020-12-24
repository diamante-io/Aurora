package actions

import (
	"context"
	"testing"

	protocol "github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/db2"
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/db2/history"
	"github.com/diamnet/go/services/aurora/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestAccountInfo(t *testing.T) {
	tt := test.Start(t).Scenario("allow_trust")
	defer tt.Finish()

	account, err := AccountInfo(tt.Ctx, &core.Q{tt.CoreSession()}, "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU")
	tt.Assert.NoError(err)

	tt.Assert.Equal("8589934593", account.Sequence)
	tt.Assert.NotEqual(0, account.LastModifiedLedger)

	for _, balance := range account.Balances {
		if balance.Type == "native" {
			tt.Assert.Equal(uint32(0), balance.LastModifiedLedger)
		} else {
			tt.Assert.NotEqual(uint32(0), balance.LastModifiedLedger)
		}
	}
}

func TestAccountPageNoResults(t *testing.T) {
	mockQ := &history.MockQSigners{}

	mockQ.On("GetLastLedgerExpIngestNonBlocking").Return(uint32(10), nil).Once()

	mockQ.
		On(
			"AccountsForSigner",
			"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			db2.PageQuery{},
		).
		Return([]history.AccountSigner{}, nil).Once()

	page, err := AccountPage(
		context.Background(),
		mockQ,
		"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
		db2.PageQuery{},
	)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(page.Embedded.Records))
}

func TestAccountPageResults(t *testing.T) {
	mockQ := &history.MockQSigners{}

	mockQ.On("GetLastLedgerExpIngestNonBlocking").Return(uint32(10), nil).Once()

	pq := db2.PageQuery{
		Order: "asc",
		Limit: 100,
	}

	rows := []history.AccountSigner{
		history.AccountSigner{
			Account: "GABGMPEKKDWR2WFH5AJOZV5PDKLJEHGCR3Q24ALETWR5H3A7GI3YTS7V",
			Signer:  "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			Weight:  1,
		},
		history.AccountSigner{
			Account: "GADTXHUTHIAESMMQ2ZWSTIIGBZRLHUCBLCHPLLUEIAWDEFRDC4SYDKOZ",
			Signer:  "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			Weight:  2,
		},
		history.AccountSigner{
			Account: "GDP347UYM2ZKE6ED6T5OM3BQ5IAS76NKRVEUPNB5PCQ26Z5D7Q7PJOMI",
			Signer:  "GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			Weight:  3,
		},
	}

	mockQ.
		On(
			"AccountsForSigner",
			"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
			pq,
		).
		Return(rows, nil).Once()

	page, err := AccountPage(
		context.Background(),
		mockQ,
		"GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU",
		pq,
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(page.Embedded.Records))

	for i, row := range rows {
		result := page.Embedded.Records[i].(protocol.AccountSigner)
		assert.Equal(t, row.Account, result.AccountID)
		assert.Equal(t, row.Signer, result.Signer.Key)
		assert.Equal(t, row.Weight, result.Signer.Weight)
	}
}
