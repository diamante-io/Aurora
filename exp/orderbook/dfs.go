package orderbook

import (
	"github.com/hcnet/go/price"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// Path represents a payment path from a source asset to some destination asset
type Path struct {
	SourceAmount xdr.Int64
	SourceAsset  xdr.Asset
	// sourceAssetString is included as an optimization to improve the performance
	// of sorting paths by avoiding serializing assets to strings repeatedly
	sourceAssetString string
	InteriorNodes     []xdr.Asset
	DestinationAsset  xdr.Asset
	DestinationAmount xdr.Int64
}

type searchState interface {
	isTerminalNode(
		currentAsset string,
		currentAssetAmount xdr.Int64,
	) bool

	appendToPaths(
		updatedVisitedList []xdr.Asset,
		currentAsset string,
		currentAssetAmount xdr.Int64,
	)

	edges(currentAssetString string) edgeSet

	consumeOffers(
		currentAssetAmount xdr.Int64,
		offers []xdr.OfferEntry,
	) (xdr.Asset, xdr.Int64, error)
}

func dfs(
	state searchState,
	maxPathLength int,
	visited map[string]bool,
	visitedList []xdr.Asset,
	currentAssetString string,
	currentAsset xdr.Asset,
	currentAssetAmount xdr.Int64,
) error {
	if currentAssetAmount <= 0 {
		return nil
	}
	if visited[currentAssetString] {
		return nil
	}
	if len(visitedList) > maxPathLength {
		return nil
	}
	visited[currentAssetString] = true
	defer func() {
		visited[currentAssetString] = false
	}()

	updatedVisitedList := append(visitedList, currentAsset)
	if state.isTerminalNode(currentAssetString, currentAssetAmount) {
		state.appendToPaths(
			updatedVisitedList,
			currentAssetString,
			currentAssetAmount,
		)
	}

	for nextAssetString, offers := range state.edges(currentAssetString) {
		if len(offers) == 0 {
			continue
		}

		nextAsset, nextAssetAmount, err := state.consumeOffers(currentAssetAmount, offers)
		if err != nil {
			return err
		}
		if nextAssetAmount <= 0 {
			continue
		}

		err = dfs(
			state,
			maxPathLength,
			visited,
			updatedVisitedList,
			nextAssetString,
			nextAsset,
			nextAssetAmount,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// sellingGraphSearchState configures a DFS on the orderbook graph
// where only edges in `graph.edgesForSellingAsset` are traversed.
// The DFS maintains the following invariants:
// no node is repeated
// no offers are consumed from the `ignoreOffersFrom` account
// each payment path must begin with an asset in `targetAssets`
// also, the required source asset amount cannot exceed the balance in `targetAssets`
type sellingGraphSearchState struct {
	graph                  *OrderBookGraph
	destinationAsset       xdr.Asset
	destinationAssetAmount xdr.Int64
	ignoreOffersFrom       xdr.AccountId
	targetAssets           map[string]xdr.Int64
	paths                  []Path
}

func (state *sellingGraphSearchState) isTerminalNode(
	currentAsset string,
	currentAssetAmount xdr.Int64,
) bool {
	targetAssetBalance, ok := state.targetAssets[currentAsset]
	return ok && targetAssetBalance >= currentAssetAmount
}

func (state *sellingGraphSearchState) appendToPaths(
	updatedVisitedList []xdr.Asset,
	currentAsset string,
	currentAssetAmount xdr.Int64,
) {
	var interiorNodes []xdr.Asset
	if len(updatedVisitedList) > 2 {
		// reverse updatedVisitedList and skip the first and last elements
		interiorNodes = make([]xdr.Asset, len(updatedVisitedList)-2)
		position := 0
		for i := len(updatedVisitedList) - 2; i >= 1; i-- {
			interiorNodes[position] = updatedVisitedList[i]
			position++
		}
	} else {
		interiorNodes = []xdr.Asset{}
	}

	state.paths = append(state.paths, Path{
		sourceAssetString: currentAsset,
		SourceAmount:      currentAssetAmount,
		SourceAsset:       updatedVisitedList[len(updatedVisitedList)-1],
		InteriorNodes:     interiorNodes,
		DestinationAsset:  state.destinationAsset,
		DestinationAmount: state.destinationAssetAmount,
	})
}

func (state *sellingGraphSearchState) edges(currentAssetString string) edgeSet {
	return state.graph.edgesForSellingAsset[currentAssetString]
}

func (state *sellingGraphSearchState) consumeOffers(
	currentAssetAmount xdr.Int64,
	offers []xdr.OfferEntry,
) (xdr.Asset, xdr.Int64, error) {
	var nextAsset xdr.Asset
	nextAmount, err := consumeOffersForSellingAsset(offers, state.ignoreOffersFrom, currentAssetAmount)
	if err == nil {
		nextAsset = offers[0].Buying
	}

	return nextAsset, nextAmount, err
}

// buyingGraphSearchState configures a DFS on the orderbook graph
// where only edges in `graph.edgesForBuyingAsset` are traversed.
// The DFS maintains the following invariants:
// no node is repeated
// no offers are consumed from the `ignoreOffersFrom` account
// each payment path must terminate with an asset in `targetAssets`
// each payment path must begin with `sourceAsset`
type buyingGraphSearchState struct {
	graph             *OrderBookGraph
	sourceAsset       xdr.Asset
	sourceAssetAmount xdr.Int64
	ignoreOffersFrom  *xdr.AccountId
	targetAssets      map[string]bool
	paths             []Path
}

func (state *buyingGraphSearchState) isTerminalNode(
	currentAsset string,
	currentAssetAmount xdr.Int64,
) bool {
	return state.targetAssets[currentAsset]
}

func (state *buyingGraphSearchState) appendToPaths(
	updatedVisitedList []xdr.Asset,
	currentAsset string,
	currentAssetAmount xdr.Int64,
) {
	var interiorNodes []xdr.Asset
	if len(updatedVisitedList) > 2 {
		// skip the first and last elements
		interiorNodes = make([]xdr.Asset, len(updatedVisitedList)-2)
		copy(interiorNodes, updatedVisitedList[1:len(updatedVisitedList)-1])
	} else {
		interiorNodes = []xdr.Asset{}
	}

	state.paths = append(state.paths, Path{
		SourceAmount:      state.sourceAssetAmount,
		SourceAsset:       state.sourceAsset,
		InteriorNodes:     interiorNodes,
		DestinationAsset:  updatedVisitedList[len(updatedVisitedList)-1],
		DestinationAmount: currentAssetAmount,
	})
}

func (state *buyingGraphSearchState) edges(currentAsset string) edgeSet {
	return state.graph.edgesForBuyingAsset[currentAsset]
}

func (state *buyingGraphSearchState) consumeOffers(
	currentAssetAmount xdr.Int64,
	offers []xdr.OfferEntry,
) (xdr.Asset, xdr.Int64, error) {
	var nextAsset xdr.Asset
	nextAmount, err := consumeOffersForBuyingAsset(offers, state.ignoreOffersFrom, currentAssetAmount)
	if err == nil {
		nextAsset = offers[0].Selling
	}

	return nextAsset, nextAmount, err
}

func consumeOffersForSellingAsset(
	offers []xdr.OfferEntry,
	ignoreOffersFrom xdr.AccountId,
	currentAssetAmount xdr.Int64,
) (xdr.Int64, error) {
	totalConsumed := xdr.Int64(0)

	if len(offers) == 0 {
		return totalConsumed, errEmptyOffers
	}

	if currentAssetAmount == 0 {
		return totalConsumed, errAssetAmountIsZero
	}

	for _, offer := range offers {
		if offer.SellerId.Equals(ignoreOffersFrom) {
			continue
		}
		if offer.Price.D == 0 {
			return -1, errOfferPriceDenominatorIsZero
		}

		buyingUnitsFromOffer, sellingUnitsFromOffer, err := price.ConvertToBuyingUnits(
			int64(offer.Amount),
			int64(currentAssetAmount),
			int64(offer.Price.N),
			int64(offer.Price.D),
		)
		if err != nil {
			return -1, errors.Wrap(err, "could not determine buying units")
		}

		totalConsumed += xdr.Int64(buyingUnitsFromOffer)
		currentAssetAmount -= xdr.Int64(sellingUnitsFromOffer)

		if currentAssetAmount <= 0 {
			return totalConsumed, nil
		}
	}

	return -1, nil
}

func consumeOffersForBuyingAsset(
	offers []xdr.OfferEntry,
	ignoreOffersFrom *xdr.AccountId,
	currentAssetAmount xdr.Int64,
) (xdr.Int64, error) {
	totalConsumed := xdr.Int64(0)

	if len(offers) == 0 {
		return totalConsumed, errEmptyOffers
	}

	if currentAssetAmount == 0 {
		return totalConsumed, errAssetAmountIsZero
	}

	for _, offer := range offers {
		if ignoreOffersFrom != nil && ignoreOffersFrom.Equals(offer.SellerId) {
			continue
		}
		if offer.Price.D == 0 {
			return -1, errOfferPriceDenominatorIsZero
		}

		n := int64(offer.Price.N)
		d := int64(offer.Price.D)

		// check if we can spend all of currentAssetAmount on the current offer
		// otherwise consume entire offer and move on to the next one
		amountSold, err := price.MulFractionRoundDown(int64(currentAssetAmount), d, n)
		if err != nil {
			return -1, errors.Wrap(err, "could not determine selling units needed")
		}
		if amountSoldXDR := xdr.Int64(amountSold); amountSoldXDR <= offer.Amount {
			totalConsumed += amountSoldXDR
			return totalConsumed, nil
		}

		buyingUnitsFromOffer, sellingUnitsFromOffer, err := price.ConvertToBuyingUnits(
			int64(offer.Amount),
			int64(offer.Amount),
			n,
			d,
		)
		if err != nil {
			return -1, errors.Wrap(err, "could not determine selling units")
		}

		totalConsumed += xdr.Int64(sellingUnitsFromOffer)
		currentAssetAmount -= xdr.Int64(buyingUnitsFromOffer)

		if currentAssetAmount <= 0 {
			return totalConsumed, nil
		}
	}

	return -1, nil
}
