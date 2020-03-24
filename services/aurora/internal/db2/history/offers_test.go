package history

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/hcnet/go/xdr"
)

var (
	issuer   = xdr.MustAddress("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
	usdAsset = xdr.Asset{
		Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
		AlphaNum4: &xdr.AssetAlphaNum4{
			AssetCode: [4]byte{'u', 's', 'd', 0},
			Issuer:    issuer,
		},
	}

	nativeAsset = xdr.Asset{
		Type: xdr.AssetTypeAssetTypeNative,
	}

	eurAsset = xdr.Asset{
		Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
		AlphaNum4: &xdr.AssetAlphaNum4{
			AssetCode: [4]byte{'e', 'u', 'r', 0},
			Issuer:    issuer,
		},
	}
	eurOffer = xdr.OfferEntry{
		SellerId: issuer,
		OfferId:  xdr.Int64(4),
		Buying:   eurAsset,
		Selling:  nativeAsset,
		Price: xdr.Price{
			N: 1,
			D: 1,
		},
		Flags:  1,
		Amount: xdr.Int64(500),
	}
	twoEurOffer = xdr.OfferEntry{
		SellerId: issuer,
		OfferId:  xdr.Int64(5),
		Buying:   eurAsset,
		Selling:  nativeAsset,
		Price: xdr.Price{
			N: 2,
			D: 1,
		},
		Flags:  2,
		Amount: xdr.Int64(500),
	}
)

func assertOfferEntryMatchesDBOffer(t *testing.T, offerEntry xdr.OfferEntry, offer Offer, lastModifiedLedger xdr.Uint32) {
	if offerEntry.SellerId.Address() != offer.SellerID {
		t.Fatalf(
			"seller id in offer entry %v does not equal to seller id in offer from db %v",
			offerEntry.SellerId.Address(),
			offer.SellerID,
		)
	}
	if offerEntry.OfferId != offer.OfferID {
		t.Fatalf(
			"offer id in offer entry %v does not equal to offer id in offer from db %v",
			offerEntry.OfferId,
			offer.OfferID,
		)
	}
	if offerEntry.Selling.String() != offer.SellingAsset.String() {
		t.Fatalf(
			"selling asset in offer entry %v does not equal to selling asset in offer from db %v",
			offerEntry.Selling.String(),
			offer.SellingAsset.String(),
		)
	}
	if offerEntry.Buying.String() != offer.BuyingAsset.String() {
		t.Fatalf(
			"buying asset in offer entry %v does not equal to buying asset in offer from db %v",
			offerEntry.Buying.String(),
			offer.BuyingAsset.String(),
		)
	}
	if offerEntry.Amount != offer.Amount {
		t.Fatalf(
			"amount in offer entry %v does not equal to amount in offer from db %v",
			offerEntry.Amount,
			offer.Amount,
		)
	}
	if offerEntry.Price.N != xdr.Int32(offer.Pricen) {
		t.Fatalf(
			"price numerator in offer entry %v does not equal to price numerator in offer from db %v",
			offerEntry.Price.N,
			offer.Pricen,
		)
	}
	if offerEntry.Price.D != xdr.Int32(offer.Priced) {
		t.Fatalf(
			"price denominator in offer entry %v does not equal to price denominator in offer from db %v",
			offerEntry.Price.D,
			offer.Priced,
		)
	}
	if offerEntry.Flags != xdr.Uint32(offer.Flags) {
		t.Fatalf(
			"flags in offer entry %v does not equal to flags in offer from db %v",
			offerEntry.Flags,
			offer.Flags,
		)
	}
	if lastModifiedLedger != xdr.Uint32(offer.LastModifiedLedger) {
		t.Fatalf(
			"last_modified_ledger %v does not equal last_modified_ledger %v in offer from DB",
			lastModifiedLedger,
			offer,
		)
	}
}

func TestGetOfferByID(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	tt.Assert.NoError(q.UpsertOffer(eurOffer, 1234))
	offer, err := q.GetOfferByID(int64(eurOffer.OfferId))
	tt.Assert.NoError(err)
	assertOfferEntryMatchesDBOffer(t, eurOffer, offer, 1234)
}

func TestGetNonExistentOfferByID(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	_, err := q.GetOfferByID(12345)
	tt.Assert.True(q.NoRows(err))
}

func TestQueryEmptyOffers(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	offers, err := q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 0)
}

func TestInsertOffers(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	tt.Assert.NoError(q.UpsertOffer(eurOffer, 1234))
	tt.Assert.NoError(q.UpsertOffer(twoEurOffer, 1235))

	offers, err := q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 2)

	offersByID := map[xdr.Int64]Offer{
		offers[0].OfferID: offers[0],
		offers[1].OfferID: offers[1],
	}

	assertOfferEntryMatchesDBOffer(t, eurOffer, offersByID[eurOffer.OfferId], 1234)
	assertOfferEntryMatchesDBOffer(t, twoEurOffer, offersByID[twoEurOffer.OfferId], 1235)
}

func TestUpdateOffer(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	tt.Assert.NoError(q.UpsertOffer(eurOffer, 1234))

	offers, err := q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 1)

	assertOfferEntryMatchesDBOffer(t, eurOffer, offers[0], 1234)

	modifiedEurOffer := eurOffer
	modifiedEurOffer.Amount -= 10

	tt.Assert.NoError(q.UpsertOffer(modifiedEurOffer, 1235))

	offers, err = q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 1)

	assertOfferEntryMatchesDBOffer(t, modifiedEurOffer, offers[0], 1235)
}

func TestRemoveNonExistantOffer(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	tt.Assert.NoError(q.RemoveOffer(xdr.Int64(12345)))
}

func TestRemoveOffer(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.AuroraSession()}

	tt.Assert.NoError(q.UpsertOffer(eurOffer, 1234))
	offers, err := q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 1)
	assertOfferEntryMatchesDBOffer(t, eurOffer, offers[0], 1234)

	tt.Assert.NoError(q.RemoveOffer(eurOffer.OfferId))

	offers, err = q.GetAllOffers()
	tt.Assert.NoError(err)
	tt.Assert.Len(offers, 0)
}
