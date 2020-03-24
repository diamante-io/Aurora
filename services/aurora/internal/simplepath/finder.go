package simplepath

import (
	"github.com/go-errors/errors"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/paths"
	"github.com/hcnet/go/support/log"
	"github.com/hcnet/go/xdr"
)

// Finder implements the paths.Finder interface and searchs for
// payment paths using a simple breadth first search of the offers table of a hcnet-core.
//
// This implementation is not meant to be fast or to provide the lowest costs paths, but
// rather is meant to be a simple implementation that gives usable paths.
type Finder struct {
	Q *core.Q
}

// ensure the struct is paths.Finder compliant
var _ paths.Finder = &Finder{}

// Find performs a path find with the provided query.
func (f *Finder) Find(q paths.Query, maxLength uint) (result []paths.Path, err error) {
	log.WithField("source_assets", q.SourceAssets).
		WithField("destination_asset", q.DestinationAsset).
		WithField("destination_amount", q.DestinationAmount).
		Info("Starting pathfind")

	if len(q.SourceAssets) == 0 {
		err = errors.New("No source assets")
		return
	}

	if maxLength == 0 {
		maxLength = MaxPathLength
	}

	if maxLength < 2 || maxLength > MaxPathLength {
		err = errors.New("invalid value of maxLength")
		return
	}

	s := &search{
		Query:     q,
		Q:         &core.Q{f.Q.Clone()},
		MaxLength: maxLength,
	}

	s.Init()
	s.Run()

	result, err = s.Results, s.Err

	log.WithField("found", len(s.Results)).
		WithField("err", s.Err).
		Info("Finished pathfind")
	return
}

// FindFixedPaths will return an error because this implementation
// does not support this operation
func (f *Finder) FindFixedPaths(
	sourceAccount *xdr.AccountId,
	sourceAsset xdr.Asset,
	amountToSpend xdr.Int64,
	destinationAsset xdr.Asset,
	maxLength uint,
) ([]paths.Path, error) {
	return nil, errors.New("Not implemented")
}
