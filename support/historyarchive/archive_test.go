// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/hcnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func GetTestS3Archive() *Archive {
	mx := big.NewInt(0xffffffff)
	r, e := rand.Int(rand.Reader, mx)
	if e != nil {
		panic(e)
	}
	return MustConnect(fmt.Sprintf("s3://history-stg.hcnet.org/dev/archivist/test-%s", r),
		ConnectOptions{S3Region: "eu-west-1"})
}

func GetTestMockArchive() *Archive {
	return MustConnect("mock://test", ConnectOptions{})
}

var tmpdirs []string

func GetTestFileArchive() *Archive {
	d, e := ioutil.TempDir("/tmp", "archivist")
	if e != nil {
		panic(e)
	}
	if tmpdirs == nil {
		tmpdirs = []string{d}
	} else {
		tmpdirs = append(tmpdirs, d)
	}
	return MustConnect("file://"+d, ConnectOptions{})
}

func cleanup() {
	for _, d := range tmpdirs {
		os.RemoveAll(d)
	}
}

func GetTestArchive() *Archive {
	ty := os.Getenv("ARCHIVIST_TEST_TYPE")
	if ty == "file" {
		return GetTestFileArchive()
	} else if ty == "s3" {
		return GetTestS3Archive()
	} else {
		return GetTestMockArchive()
	}
}

func (arch *Archive) AddRandomBucket() (Hash, error) {
	var h Hash
	buf := make([]byte, 1024)
	_, e := rand.Read(buf)
	if e != nil {
		return h, e
	}
	h = sha256.Sum256(buf)
	pth := BucketPath(h)
	e = arch.backend.PutFile(pth, ioutil.NopCloser(bytes.NewReader(buf)))
	return h, e
}

func (arch *Archive) AddRandomCheckpointFile(cat string, chk uint32) error {
	buf := make([]byte, 1024)
	_, e := rand.Read(buf)
	if e != nil {
		return e
	}
	pth := CategoryCheckpointPath(cat, chk)
	return arch.backend.PutFile(pth, ioutil.NopCloser(bytes.NewReader(buf)))
}

func (arch *Archive) AddRandomCheckpoint(chk uint32) error {
	opts := &CommandOptions{Force: true}
	for _, cat := range Categories() {
		if cat == "history" {
			var has HistoryArchiveState
			has.CurrentLedger = chk
			for i := 0; i < NumLevels; i++ {
				curr, e := arch.AddRandomBucket()
				if e != nil {
					return e
				}
				snap, e := arch.AddRandomBucket()
				if e != nil {
					return e
				}
				next, e := arch.AddRandomBucket()
				if e != nil {
					return e
				}
				has.CurrentBuckets[i].Curr = curr.String()
				has.CurrentBuckets[i].Snap = snap.String()
				has.CurrentBuckets[i].Next.Output = next.String()
			}
			arch.PutCheckpointHAS(chk, has, opts)
			arch.PutRootHAS(has, opts)
		} else {
			arch.AddRandomCheckpointFile(cat, chk)
		}
	}
	return nil
}

func (arch *Archive) PopulateRandomRange(rng Range) error {
	for chk := range rng.Checkpoints() {
		if e := arch.AddRandomCheckpoint(chk); e != nil {
			return e
		}
	}
	return nil
}

func testRange() Range {
	return Range{Low: 63, High: 0x3bf}
}

func testOptions() *CommandOptions {
	return &CommandOptions{Range: testRange(), Concurrency: 16}
}

func GetRandomPopulatedArchive() *Archive {
	a := GetTestArchive()
	a.PopulateRandomRange(testRange())
	return a
}

func TestScan(t *testing.T) {
	defer cleanup()
	opts := testOptions()
	GetRandomPopulatedArchive().Scan(opts)
}

func countMissing(arch *Archive, opts *CommandOptions) int {
	n := 0
	arch.Scan(opts)
	for _, missing := range arch.CheckCheckpointFilesMissing(opts) {
		n += len(missing)
	}
	n += len(arch.CheckBucketsMissing())
	return n
}

func TestMirror(t *testing.T) {
	defer cleanup()
	opts := testOptions()
	src := GetRandomPopulatedArchive()
	dst := GetTestArchive()
	Mirror(src, dst, opts)
	assert.Equal(t, 0, countMissing(dst, opts))
}

func copyFile(category string, checkpoint uint32, src *Archive, dst *Archive) {
	pth := CategoryCheckpointPath(category, checkpoint)
	rdr, err := src.backend.GetFile(pth)
	if err != nil {
		panic(err)
	}
	if err = dst.backend.PutFile(pth, rdr); err != nil {
		panic(err)
	}
}

func TestMirrorThenRepair(t *testing.T) {
	defer cleanup()
	opts := testOptions()
	src := GetRandomPopulatedArchive()
	dst := GetTestArchive()
	Mirror(src, dst, opts)
	assert.Equal(t, 0, countMissing(dst, opts))
	bad := opts.Range.Low + uint32(opts.Range.Size()/2)
	src.AddRandomCheckpoint(bad)
	copyFile("history", bad, src, dst)
	assert.NotEqual(t, 0, countMissing(dst, opts))
	Repair(src, dst, opts)
	assert.Equal(t, 0, countMissing(dst, opts))
}

func TestDryRunNoRepair(t *testing.T) {
	defer cleanup()
	opts := testOptions()
	src := GetRandomPopulatedArchive()
	dst := GetTestArchive()
	Mirror(src, dst, opts)
	assert.Equal(t, 0, countMissing(dst, opts))
	bad := opts.Range.Low + uint32(opts.Range.Size()/2)
	src.AddRandomCheckpoint(bad)
	copyFile("history", bad, src, dst)
	assert.NotEqual(t, 0, countMissing(dst, opts))
	opts.DryRun = true
	Repair(src, dst, opts)
	assert.NotEqual(t, 0, countMissing(dst, opts))
}

func TestXdrDecode(t *testing.T) {

	xdrbytes := []byte{

		0, 0, 0, 0, // entry type 0, liveentry

		0, 32, 223, 100, // lastmodified 2154340

		0, 0, 0, 0, // entry type 0, account

		0, 0, 0, 0, // key type 0
		23, 140, 68, 253, // ed25519 key (32 bytes)
		184, 162, 186, 195,
		118, 239, 158, 210,
		100, 241, 174, 254,
		108, 110, 165, 140,
		75, 76, 83, 141,
		104, 212, 227, 80,
		1, 214, 157, 7,

		0, 0, 0, 29, // 64bit balance: 125339976000
		46, 216, 65, 64,

		0, 0, 129, 170, // 64bit seqnum: 142567144423475
		0, 0, 0, 51,

		0, 0, 0, 1, // numsubentries: 1

		0, 0, 0, 1, // inflationdest type, populated

		0, 0, 0, 0, // key type 0
		87, 240, 19, 71, // ed25519 key (32 bytes)
		52, 91, 9, 62,
		213, 239, 178, 85,
		161, 119, 108, 251,
		168, 90, 76, 116,
		12, 48, 134, 248,
		115, 255, 117, 50,
		19, 18, 170, 203,

		0, 0, 0, 0, // flags

		0, 0, 0, 19, // homedomain: 19 bytes + 1 null padding
		99, 101, 110, 116, // "centaurus.xcoins.de"
		97, 117, 114, 117,
		115, 46, 120, 99,
		111, 105, 110, 115,
		46, 100, 101, 0,

		1, 0, 0, 0, // thresholds
		0, 0, 0, 0, // signers (null)

		0, 0, 0, 0, // entry.account.ext.v: 0

		0, 0, 0, 0, // entry.ext.v: 0
	}

	assert.Equal(t, len(xdrbytes), 152)

	var tmp xdr.BucketEntry
	n, err := xdr.Unmarshal(bytes.NewReader(xdrbytes[:]), &tmp)
	fmt.Printf("Decoded %d bytes\n", n)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(xdrbytes), n)

	var out bytes.Buffer
	n, err = xdr.Marshal(&out, &tmp)
	fmt.Printf("Encoded %d bytes\n", n)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, out.Len(), n)
	assert.Equal(t, out.Bytes(), xdrbytes)
}
