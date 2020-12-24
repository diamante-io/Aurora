package simplepath

import (
	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/services/aurora/internal/paths"
	"github.com/diamnet/go/xdr"
)

// MaxPathLength is a maximum path length as defined in XDR file (includes source and
// destination assets).
const MaxPathLength uint = 7

// search represents a single query against the simple finder.  It provides
// a place to store the results of the query, mostly for the purposes of code
// clarity.
//
// The search struct is used as follows:
//
// 1.  Create an instance, ensuring the Query and Finder fields are set
// 2.  Call Init() to populate dependent fields in the struct with their initial values
// 3.  Call Run() to perform the search.
//
type search struct {
	Query     paths.Query
	Q         *core.Q
	MaxLength uint

	// Fields below are initialized by a call to Init() after
	// setting the fields above
	queue   []computedNode
	targets map[string]bool

	//This fields below are initialized after the search is run
	Err     error
	Results []paths.Path
}

// computedNode represents a pathNode with the computed cost
type computedNode struct {
	path pathNode
	cost xdr.Int64
}

func (c computedNode) asPath(destinationAmount xdr.Int64) paths.Path {
	return paths.Path{
		Path:              c.path.Path(),
		Source:            c.path.Source(),
		SourceAmount:      c.cost,
		Destination:       c.path.Destination(),
		DestinationAmount: destinationAmount,
	}
}

const maxResults = 20

// Init initialized the search, setting fields on the struct used to
// hold state needed during the actual search.
func (s *search) Init() {
	p0 := pathNode{
		Asset: s.Query.DestinationAsset,
		Tail:  nil,
		Q:     s.Q,
		Depth: 1,
	}
	var c0 xdr.Int64
	// `Cost` on destination node does not use DB connection.
	c0, s.Err = p0.Cost(s.Query.DestinationAmount)
	if s.Err != nil {
		return
	}

	s.queue = []computedNode{
		computedNode{
			path: p0,
			cost: c0,
		},
	}

	// build a map of asset's string representation to check if a given node
	// is one of the targets for our search.  Unfortunately, xdr.Asset is not suitable
	// for use as a map key, and so we use its string representation.
	s.targets = map[string]bool{}
	for _, a := range s.Query.SourceAssets {
		s.targets[a.String()] = true
	}

	s.Err = nil
	s.Results = nil
}

// Run triggers the search, which will populate the Results and Err
// field for the search after completion.
func (s *search) Run() {
	if s.Err != nil {
		return
	}

	s.Err = s.Q.Begin()
	if s.Err != nil {
		return
	}

	defer s.Q.Rollback()

	// We need REPEATABLE READ here to have a stable view of the offers
	// table. Without it, it's possible that search started in ledger X
	// and finished in ledger X+1 would give invalid results.
	//
	// https://www.postgresql.org/docs/9.1/static/transaction-iso.html
	// > Note that only updating transactions might need to be retried;
	// > read-only transactions will never have serialization conflicts.
	_, s.Err = s.Q.ExecRaw("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY")
	if s.Err != nil {
		return
	}

	for s.hasMore() {
		s.runOnce()
	}
}

// pop removes the head from the search queue, returning it to the caller
func (s *search) pop() computedNode {
	next := s.queue[0]
	s.queue = s.queue[1:]
	return next
}

// returns false if the search should stop.
func (s *search) hasMore() bool {
	if s.Err != nil {
		return false
	}

	if len(s.Results) >= maxResults {
		return false
	}

	return len(s.queue) > 0
}

// isTarget returns true if the asset id provided is one of the targets
// for this search (i.e. one of the requesting account's trusted assets)
func (s *search) isTarget(id string) bool {
	_, found := s.targets[id]
	return found
}

// runOnce processes the head of the search queue, findings results
// and extending the search as necessary.
func (s *search) runOnce() {
	cur := s.pop()
	id := cur.path.Asset.String()

	if s.isTarget(id) {
		s.Results = append(s.Results, cur.asPath(s.Query.DestinationAmount))
	}

	if cur.path.Depth == s.MaxLength {
		return
	}

	s.extendSearch(cur.path)
}

func (s *search) extendSearch(p pathNode) {
	// find connected assets
	var connected []xdr.Asset
	s.Err = s.Q.ConnectedAssets(&connected, p.Asset)
	if s.Err != nil {
		return
	}

	for _, a := range connected {
		// If asset already exists on the path, continue to the next one.
		// We don't want the same asset on the path twice as buying and
		// then selling the asset will be a bad deal in most cases
		// (especially A -> B -> A trades).
		if p.IsOnPath(a) {
			continue
		}

		// If the connected asset is not our target and the current length
		// of the path is MaxLength-1 then it does not make sense to extend
		// such path.
		if p.Depth == s.MaxLength-1 && !s.isTarget(a.String()) {
			continue
		}

		newPath := pathNode{
			Asset: a,
			Tail:  &p,
			Q:     s.Q,
			Depth: p.Depth + 1,
		}

		var hasEnough bool
		var cost xdr.Int64
		hasEnough, cost, s.Err = s.hasEnoughDepth(&newPath)
		if s.Err != nil {
			return
		}

		if !hasEnough {
			continue
		}

		s.queue = append(s.queue, computedNode{newPath, cost})
	}
}

func (s *search) hasEnoughDepth(path *pathNode) (bool, xdr.Int64, error) {
	cost, err := path.Cost(s.Query.DestinationAmount)
	if err == ErrNotEnough {
		return false, 0, nil
	}
	return true, cost, err
}
