package orderbook

import (
	"math/big"
	"sort"

	"github.com/hcnet/go/xdr"
)

// edgeSet maintains a maping of strings to sorted lists of offers.
// The offers are sorted by price (in terms of the buying asset)
// from cheapest to expensive
type edgeSet map[string][]xdr.OfferEntry

// add will insert the given offer into the edge set
func (e edgeSet) add(key string, offer xdr.OfferEntry) {
	// offers is sorted by cheapest to most expensive price to convert buyingAsset to sellingAsset
	offers := e[key]
	if len(offers) == 0 {
		e[key] = []xdr.OfferEntry{offer}
		return
	}

	// find the smallest i such that Price of offers[i] >  Price of offer
	insertIndex := sort.Search(len(offers), func(i int) bool {
		return big.NewRat(int64(offers[i].Price.N), int64(offers[i].Price.D)).
			Cmp(big.NewRat(int64(offer.Price.N), int64(offer.Price.D))) > 0
	})

	offers = append(offers, offer)
	last := len(offers) - 1
	for insertIndex < last {
		offers[insertIndex], offers[last] = offers[last], offers[insertIndex]
		insertIndex++
	}
	e[key] = offers
}

// remove will delete the given offer from the edge set
func (e edgeSet) remove(offerID xdr.Int64, key string) bool {
	edges := e[key]
	if len(edges) == 0 {
		return false
	}
	contains := false

	for i := 0; i < len(edges); i++ {
		if edges[i].OfferId == offerID {
			contains = true
			for j := i + 1; j < len(edges); j++ {
				edges[i] = edges[j]
				i++
			}
			edges = edges[0 : len(edges)-1]
			break
		}
	}
	if contains {
		if len(edges) == 0 {
			delete(e, key)
		} else {
			e[key] = edges
		}
	}

	return contains
}
