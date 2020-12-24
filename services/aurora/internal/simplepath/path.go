package simplepath

import (
	"bytes"
	"fmt"

	"github.com/diamnet/go/services/aurora/internal/db2/core"
	"github.com/diamnet/go/xdr"
)

// pathNode represents a path as a linked list pointing from source to destination
type pathNode struct {
	Asset      xdr.Asset
	Tail       *pathNode
	Q          *core.Q
	CachedCost *xdr.Int64
	Depth      uint
}

func (p *pathNode) String() string {
	if p == nil {
		return ""
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "%v", p.Asset)

	cur := p.Tail

	for cur != nil {
		fmt.Fprintf(&out, " -> %v", cur.Asset)
		cur = cur.Tail
	}

	return out.String()
}

// Destination returns the destination of the pathNode
func (p *pathNode) Destination() xdr.Asset {
	cur := p
	for cur.Tail != nil {
		cur = cur.Tail
	}
	return cur.Asset
}

// IsOnPath returns true if a given asset is in the path.
func (p *pathNode) IsOnPath(asset xdr.Asset) bool {
	cur := p
	for cur.Tail != nil {
		if asset.Equals(cur.Asset) {
			return true
		}
		cur = cur.Tail
	}

	return asset.Equals(cur.Asset)
}

// Source returns the source asset in the pathNode
func (p *pathNode) Source() xdr.Asset {
	// the destination for path is the head of the linked list
	return p.Asset
}

// Path returns the path of the list excluding the source and destination assets
func (p *pathNode) Path() []xdr.Asset {
	path := p.Flatten()

	if len(path) < 2 {
		return nil
	}

	// return the flattened slice without the first and last elements
	// which are the source and the destination assets
	return path[1 : len(path)-1]
}

// Cost computes the units of the source asset needed to send the amount in the destination asset
// This is an expensive operation so callers should reuse the result where appropriate
func (p *pathNode) Cost(amount xdr.Int64) (xdr.Int64, error) {
	if p.Tail == nil {
		return amount, nil
	}

	if p.CachedCost != nil {
		return *p.CachedCost, nil
	}

	// The first element of the current path is the current source asset.
	// The last element (with `Tail` = nil) of the current path is the destination
	// asset. To make the calculations correct we start by selling destination
	// asset to the second from the end asset and continue until we reach the current
	// source asset.
	cur := p
	stack := make([]*pathNode, 0, p.Depth)
	for cur.Tail != nil {
		stack = append(stack, cur)
		cur = cur.Tail
	}

	var err error
	result := amount

	for i := len(stack) - 1; i >= 0; i-- {
		cur = stack[i]

		if cur.CachedCost != nil {
			result = *cur.CachedCost
			continue
		}

		ob := cur.OrderBook()
		result, err = ob.CostToConsumeLiquidity(result)
		if err != nil {
			return result, err
		}
	}

	// Cache the result
	cur.CachedCost = &result

	return result, nil
}

func (p *pathNode) OrderBook() *orderBook {
	if p.Tail == nil {
		return nil
	}

	return &orderBook{
		Selling: p.Tail.Asset, // offer is selling this asset
		Buying:  p.Asset,      // offer is buying this asset
		Q:       p.Q,
	}
}

// Flatten walks the list and returns a slice of assets
func (p *pathNode) Flatten() []xdr.Asset {
	result := []xdr.Asset{}
	cur := p
	for cur != nil {
		result = append(result, cur.Asset)
		cur = cur.Tail
	}
	return result
}
