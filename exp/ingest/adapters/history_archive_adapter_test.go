package adapters

import (
	"fmt"
	stdio "io"
	"testing"

	"github.com/hcnet/go/exp/ingest/io"
	"github.com/hcnet/go/support/historyarchive"
	"github.com/stretchr/testify/assert"
)

// commented out this test for now
// func TestGetLatestLedgerSequence(t *testing.T) {
// 	archive, e := getTestArchive()
// 	if !assert.NoError(t, e) {
// 		return
// 	}

// 	haa := MakeHistoryArchiveAdapter(archive)
// 	seq, e := haa.GetLatestLedgerSequence()
// 	if !assert.NoError(t, e) {
// 		return
// 	}
// 	assert.Equal(t, uint32(931455), seq)
// }

func TestGetState_Sequence(t *testing.T) {
	archive, e := getTestArchive()
	if !assert.NoError(t, e) {
		return
	}
	haa := MakeHistoryArchiveAdapter(archive)

	seq, e := haa.GetLatestLedgerSequence()
	if !assert.NoError(t, e) {
		return
	}

	sr, e := haa.GetState(seq, &io.MemoryTempSet{})
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, sr.GetSequence(), seq)
}

func TestGetState_Read(t *testing.T) {
	archive, e := getTestArchive()
	if !assert.NoError(t, e) {
		return
	}
	haa := MakeHistoryArchiveAdapter(archive)

	sr, e := haa.GetState(21686847, &io.MemoryTempSet{})
	if !assert.NoError(t, e) {
		return
	}

	lec, e := sr.Read()
	if !assert.NoError(t, e) {
		return
	}
	assert.NotEqual(t, e, stdio.EOF)

	if !assert.NotNil(t, lec) {
		return
	}
	assert.Equal(t, "GAFBQT4VRORLEVEECUYDQGWNVQ563ZN76LGRJR7T7KDL32EES54UOQST", lec.State.Data.Account.AccountId.Address())
}

func getTestArchive() (*historyarchive.Archive, error) {
	return historyarchive.Connect(
		fmt.Sprintf("s3://history.hcnet.org/prd/core-live/core_live_001/"),
		historyarchive.ConnectOptions{
			S3Region:         "eu-west-1",
			UnsignedRequests: true,
		},
	)
}
