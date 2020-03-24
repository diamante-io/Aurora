// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const hexPrefixPat = "/[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{2}/"
const rootHASPath = ".well-known/hcnet-history.json"

type CommandOptions struct {
	Concurrency int
	Range       Range
	DryRun      bool
	Force       bool
	Verify      bool
	Thorough    bool
}

type ConnectOptions struct {
	S3Region         string
	S3Endpoint       string
	UnsignedRequests bool
}

type ArchiveBackend interface {
	Exists(path string) (bool, error)
	GetFile(path string) (io.ReadCloser, error)
	PutFile(path string, in io.ReadCloser) error
	ListFiles(path string) (chan string, chan error)
	CanListFiles() bool
}

type ArchiveInterface interface {
	GetPathHAS(path string) (HistoryArchiveState, error)
	PutPathHAS(path string, has HistoryArchiveState, opts *CommandOptions) error
	BucketExists(bucket Hash) (bool, error)
	CategoryCheckpointExists(cat string, chk uint32) (bool, error)
	GetRootHAS() (HistoryArchiveState, error)
	GetCheckpointHAS(chk uint32) (HistoryArchiveState, error)
	PutCheckpointHAS(chk uint32, has HistoryArchiveState, opts *CommandOptions) error
	PutRootHAS(has HistoryArchiveState, opts *CommandOptions) error
	ListBucket(dp DirPrefix) (chan string, chan error)
	ListAllBuckets() (chan string, chan error)
	ListAllBucketHashes() (chan Hash, chan error)
	ListCategoryCheckpoints(cat string, pth string) (chan uint32, chan error)
	GetXdrStreamForHash(hash Hash) (*XdrStream, error)
	GetXdrStream(pth string) (*XdrStream, error)
}

var _ ArchiveInterface = &Archive{}

type Archive struct {
	mutex             sync.Mutex
	checkpointFiles   map[string](map[uint32]bool)
	allBuckets        map[Hash]bool
	referencedBuckets map[Hash]bool

	expectLedgerHashes      map[uint32]Hash
	actualLedgerHashes      map[uint32]Hash
	expectTxSetHashes       map[uint32]Hash
	actualTxSetHashes       map[uint32]Hash
	expectTxResultSetHashes map[uint32]Hash
	actualTxResultSetHashes map[uint32]Hash

	invalidBuckets int

	invalidLedgers      int
	invalidTxSets       int
	invalidTxResultSets int

	backend ArchiveBackend
}

func (a *Archive) GetPathHAS(path string) (HistoryArchiveState, error) {
	var has HistoryArchiveState
	rdr, err := a.backend.GetFile(path)
	if err != nil {
		return has, err
	}
	defer rdr.Close()
	dec := json.NewDecoder(rdr)
	err = dec.Decode(&has)
	return has, err
}

func (a *Archive) PutPathHAS(path string, has HistoryArchiveState, opts *CommandOptions) error {
	exists, err := a.backend.Exists(path)
	if err != nil {
		return err
	}
	if exists && !opts.Force {
		log.Printf("skipping existing " + path)
		return nil
	}
	buf, err := json.MarshalIndent(has, "", "    ")
	if err != nil {
		return err
	}
	return a.backend.PutFile(path,
		ioutil.NopCloser(bytes.NewReader(buf)))
}

func (a *Archive) BucketExists(bucket Hash) (bool, error) {
	return a.backend.Exists(BucketPath(bucket))
}

func (a *Archive) CategoryCheckpointExists(cat string, chk uint32) (bool, error) {
	return a.backend.Exists(CategoryCheckpointPath(cat, chk))
}

func (a *Archive) GetRootHAS() (HistoryArchiveState, error) {
	return a.GetPathHAS(rootHASPath)
}

func (a *Archive) GetCheckpointHAS(chk uint32) (HistoryArchiveState, error) {
	return a.GetPathHAS(CategoryCheckpointPath("history", chk))
}

func (a *Archive) PutCheckpointHAS(chk uint32, has HistoryArchiveState, opts *CommandOptions) error {
	return a.PutPathHAS(CategoryCheckpointPath("history", chk), has, opts)
}

func (a *Archive) PutRootHAS(has HistoryArchiveState, opts *CommandOptions) error {
	force := opts.Force
	opts.Force = true
	e := a.PutPathHAS(rootHASPath, has, opts)
	opts.Force = force
	return e
}

func (a *Archive) ListBucket(dp DirPrefix) (chan string, chan error) {
	return a.backend.ListFiles(path.Join("bucket", dp.Path()))
}

func (a *Archive) ListAllBuckets() (chan string, chan error) {
	return a.backend.ListFiles("bucket")
}

func (a *Archive) ListAllBucketHashes() (chan Hash, chan error) {
	sch, errs := a.backend.ListFiles("bucket")
	ch := make(chan Hash)
	rx := regexp.MustCompile("bucket" + hexPrefixPat + "bucket-([0-9a-f]{64})\\.xdr\\.gz$")
	errs = makeErrorPump(errs)
	go func() {
		for s := range sch {
			m := rx.FindStringSubmatch(s)
			if m != nil {
				ch <- MustDecodeHash(m[1])
			}
		}
		close(ch)
	}()
	return ch, errs
}

func (a *Archive) ListCategoryCheckpoints(cat string, pth string) (chan uint32, chan error) {
	ext := categoryExt(cat)
	rx := regexp.MustCompile(cat + hexPrefixPat + cat +
		"-([0-9a-f]{8})\\." + regexp.QuoteMeta(ext) + "$")
	sch, errs := a.backend.ListFiles(path.Join(cat, pth))
	ch := make(chan uint32)
	errs = makeErrorPump(errs)

	go func() {
		for s := range sch {
			m := rx.FindStringSubmatch(s)
			if m != nil {
				i, e := strconv.ParseUint(m[1], 16, 32)
				if e == nil {
					ch <- uint32(i)
				} else {
					errs <- errors.New("decoding checkpoint number in filename " + s)
				}
			}
		}
		close(ch)
	}()
	return ch, errs
}

func (a *Archive) GetXdrStreamForHash(hash Hash) (*XdrStream, error) {
	path := fmt.Sprintf(
		"bucket/%s/bucket-%s.xdr.gz",
		HashPrefix(hash).Path(),
		hash.String(),
	)

	return a.GetXdrStream(path)
}

func (a *Archive) GetXdrStream(pth string) (*XdrStream, error) {
	if !strings.HasSuffix(pth, ".xdr.gz") {
		return nil, errors.New("File has non-.xdr.gz suffix: " + pth)
	}
	rdr, err := a.backend.GetFile(pth)
	if err != nil {
		return nil, err
	}
	return NewXdrGzStream(rdr)
}

func Connect(u string, opts ConnectOptions) (*Archive, error) {
	arch := Archive{
		checkpointFiles:         make(map[string](map[uint32]bool)),
		allBuckets:              make(map[Hash]bool),
		referencedBuckets:       make(map[Hash]bool),
		expectLedgerHashes:      make(map[uint32]Hash),
		actualLedgerHashes:      make(map[uint32]Hash),
		expectTxSetHashes:       make(map[uint32]Hash),
		actualTxSetHashes:       make(map[uint32]Hash),
		expectTxResultSetHashes: make(map[uint32]Hash),
		actualTxResultSetHashes: make(map[uint32]Hash),
	}
	for _, cat := range Categories() {
		arch.checkpointFiles[cat] = make(map[uint32]bool)
	}

	if u == "" {
		return &arch, errors.New("URL is empty")
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return &arch, err
	}
	pth := parsed.Path
	if parsed.Scheme == "s3" {
		// Inside s3, all paths start _without_ the leading /
		if len(pth) > 0 && pth[0] == '/' {
			pth = pth[1:]
		}
		arch.backend, err = makeS3Backend(parsed.Host, pth, opts)
	} else if parsed.Scheme == "file" {
		pth = path.Join(parsed.Host, pth)
		arch.backend = makeFsBackend(pth, opts)
	} else if parsed.Scheme == "http" || parsed.Scheme == "https" {
		arch.backend = makeHttpBackend(parsed, opts)
	} else if parsed.Scheme == "mock" {
		arch.backend = makeMockBackend(opts)
	} else {
		err = errors.New("unknown URL scheme: '" + parsed.Scheme + "'")
	}
	return &arch, err
}

func MustConnect(u string, opts ConnectOptions) *Archive {
	arch, err := Connect(u, opts)
	if err != nil {
		log.Fatal(err)
	}
	return arch
}
