// Copyright 2016 DiamNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/diamnet/go/support/errors"
)

func Mirror(src *Archive, dst *Archive, opts *CommandOptions) error {
	rootHAS, e := src.GetRootHAS()
	if e != nil {
		return e
	}

	opts.Range = opts.Range.clamp(rootHAS.Range())

	log.Printf("copying range %s\n", opts.Range)

	// Make a bucket-fetch map that shows which buckets are
	// already-being-fetched
	bucketFetch := make(map[Hash]bool)
	var bucketFetchMutex sync.Mutex

	var errs uint32
	tick := makeTicker(func(ticks uint) {
		bucketFetchMutex.Lock()
		sz := opts.Range.Size()
		log.Printf("Copied %d/%d checkpoints (%f%%), %d buckets",
			ticks, sz,
			100.0*float64(ticks)/float64(sz),
			len(bucketFetch))
		bucketFetchMutex.Unlock()
	})

	var wg sync.WaitGroup
	checkpoints := opts.Range.Checkpoints()
	wg.Add(opts.Concurrency)
	for i := 0; i < opts.Concurrency; i++ {
		go func() {
			for {
				ix, ok := <-checkpoints
				if !ok {
					break
				}
				has, err := src.GetCheckpointHAS(ix)
				if err != nil {
					atomic.AddUint32(&errs, noteError(err))
					continue
				}

				buckets, err := has.Buckets()
				if err != nil {
					panic(errors.Wrap(err, "error getting buckets"))
				}

				for _, bucket := range buckets {
					alreadyFetching := false
					bucketFetchMutex.Lock()
					_, alreadyFetching = bucketFetch[bucket]
					if !alreadyFetching {
						bucketFetch[bucket] = true
					}
					bucketFetchMutex.Unlock()
					if !alreadyFetching {
						pth := BucketPath(bucket)
						err = copyPath(src, dst, pth, opts)
						atomic.AddUint32(&errs, noteError(err))
					}
				}

				for _, cat := range Categories() {
					pth := CategoryCheckpointPath(cat, ix)
					err = copyPath(src, dst, pth, opts)
					if err != nil && !categoryRequired(cat) {
						continue
					}
					atomic.AddUint32(&errs, noteError(err))
				}
				tick <- true
			}
			wg.Done()
		}()
	}

	wg.Wait()
	log.Printf("Copied %d checkpoints, %d buckets",
		opts.Range.Size(), len(bucketFetch))
	close(tick)
	e = dst.PutRootHAS(rootHAS, opts)
	errs += noteError(e)
	if errs != 0 {
		return fmt.Errorf("%d errors while mirroring", errs)
	}
	return nil
}
