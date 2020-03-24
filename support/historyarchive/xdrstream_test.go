// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"testing"

	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXdrStreamHash(t *testing.T) {
	bucketEntry := xdr.BucketEntry{
		Type: xdr.BucketEntryTypeLiveentry,
		LiveEntry: &xdr.LedgerEntry{
			Data: xdr.LedgerEntryData{
				Type: xdr.LedgerEntryTypeAccount,
				Account: &xdr.AccountEntry{
					AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
					Balance:   xdr.Int64(200000000),
				},
			},
		},
	}
	stream := createXdrStream(bucketEntry)

	// Stream hash should be equal sha256 hash of concatenation of:
	// - uint32 representing a number o bytes of a structure,
	// - xdr-encoded `BucketEntry` above.
	b := &bytes.Buffer{}
	err := WriteFramedXdr(b, bucketEntry)
	require.NoError(t, err)

	expectedHash := sha256.Sum256(b.Bytes())
	stream.SetExpectedHash(expectedHash)

	var readBucketEntry xdr.BucketEntry
	err = stream.ReadOne(&readBucketEntry)
	require.NoError(t, err)
	assert.Equal(t, bucketEntry, readBucketEntry)

	assert.NoError(t, stream.Close())
}

func createXdrStream(entries ...xdr.BucketEntry) *XdrStream {
	b := &bytes.Buffer{}
	for _, e := range entries {
		err := WriteFramedXdr(b, e)
		if err != nil {
			panic(err)
		}
	}

	return NewXdrStream(ioutil.NopCloser(b))
}
