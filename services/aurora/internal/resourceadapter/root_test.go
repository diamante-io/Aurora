package resourceadapter

import (
	"context"
	"net/url"
	"testing"

	"github.com/hcnet/go/clients/aurora"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/stretchr/testify/assert"
)

func TestPopulateRoot(t *testing.T) {
	res := &aurora.Root{}
	PopulateRoot(context.Background(),
		res,
		ledger.State{CoreLatest: 1, HistoryLatest: 3, HistoryElder: 2},
		"hVersion",
		"cVersion",
		"passphrase",
		100,
		101,
		urlMustParse(t, "https://friendbot.example.com"))

	assert.Equal(t, int32(1), res.CoreSequence)
	assert.Equal(t, int32(2), res.HistoryElderSequence)
	assert.Equal(t, int32(3), res.AuroraSequence)
	assert.Equal(t, "hVersion", res.AuroraVersion)
	assert.Equal(t, "cVersion", res.HcNetCoreVersion)
	assert.Equal(t, "passphrase", res.NetworkPassphrase)
	assert.Equal(t, "https://friendbot.example.com/{?addr}", res.Links.Friendbot.Href)

	// Without testbot
	res = &aurora.Root{}
	PopulateRoot(context.Background(),
		res,
		ledger.State{CoreLatest: 1, HistoryLatest: 3, HistoryElder: 2},
		"hVersion",
		"cVersion",
		"passphrase",
		100,
		101,
		nil)

	assert.Equal(t, int32(1), res.CoreSequence)
	assert.Equal(t, int32(2), res.HistoryElderSequence)
	assert.Equal(t, int32(3), res.AuroraSequence)
	assert.Equal(t, "hVersion", res.AuroraVersion)
	assert.Equal(t, "cVersion", res.HcNetCoreVersion)
	assert.Equal(t, "passphrase", res.NetworkPassphrase)
	assert.Empty(t, res.Links.Friendbot)
}

func urlMustParse(t *testing.T, s string) *url.URL {
	if u, err := url.Parse(s); err != nil {
		t.Fatalf("Unable to parse URL: %s/%v", s, err)
		return nil
	} else {
		return u
	}
}
