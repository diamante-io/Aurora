package meta_test

import (
	. "github.com/diamnet/go/meta"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/diamnet/go/xdr"
)

var _ = Describe("meta.Bundle", func() {
	var createAccount = bundle(
		"AAAAAgAAAAMAAAABAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAACAAAAAAAAAABi/B0L0JGythwN1lY0aypo19NHxvLCyO5tBEcCVvwF9w3gtrOnY/+cAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==",
		"AAAAAAAAAAEAAAACAAAAAAAAAAIAAAAAAAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAADuaygAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAIAAAAAAAAAAGL8HQvQkbK2HA3WVjRrKmjX00fG8sLI7m0ERwJW/AX3DeC2s2vJNNQAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAA",
	)

	var removeTrustline = bundle(
		"AAAAAgAAAAMAAAAEAAAAAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAJUC+M4AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAFAAAAAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAJUC+LUAAAAAgAAAAMAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==",
		"AAAAAAAAAAEAAAADAAAAAQAAAAUAAAAAAAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAAlQL4tQAAAACAAAAAwAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAwAAAAQAAAABAAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAAVVTRAAAAAAAtbgXR6E7oDL0LQ+wYSC9zXvXVT3xiPiYuSb1DvmQLe8AAAAAAAAAAAAAAAlQL5AAAAAAAQAAAAAAAAAAAAAAAgAAAAEAAAAArqN6LeOagjxMaUP96Bzfs9e0corNZXzBWJkFoK7kvkwAAAABVVNEAAAAAAC1uBdHoTugMvQtD7BhIL3Ne9dVPfGI+Ji5JvUO+ZAt7w==",
	)

	var updateTrustline = bundle(
		"AAAAAgAAAAMAAAADAAAAAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAJUC+OcAAAAAgAAAAEAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAAAAEAAAAAAAAAACuo3ot45qCPExpQ/3oHN+z17Ryis1lfMFYmQWgruS+TAAAAAJUC+M4AAAAAgAAAAIAAAABAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==",
		"AAAAAAAAAAEAAAACAAAAAwAAAAMAAAABAAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAAVVTRAAAAAAAtbgXR6E7oDL0LQ+wYSC9zXvXVT3xiPiYuSb1DvmQLe8AAAAAAAAAAH//////////AAAAAQAAAAAAAAAAAAAAAQAAAAQAAAABAAAAAK6jei3jmoI8TGlD/egc37PXtHKKzWV8wViZBaCu5L5MAAAAAVVTRAAAAAAAtbgXR6E7oDL0LQ+wYSC9zXvXVT3xiPiYuSb1DvmQLe8AAAAAAAAAAAAAAAlQL5AAAAAAAQAAAAAAAAAA",
	)
	// var mergeAccount = nil //TODO

	var newAccount xdr.AccountId
	var masterAccount xdr.AccountId
	var nonexistantAccount xdr.AccountId
	var gatewayAccount xdr.AccountId

	BeforeEach(func() {
		err := newAccount.SetAddress("GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU")
		Expect(err).ToNot(HaveOccurred())
		err = masterAccount.SetAddress("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
		Expect(err).ToNot(HaveOccurred())
		err = nonexistantAccount.SetAddress("GDGAWQZT2RALG2XBEESTMA7PHDASK4EZGXWGBCIHZRSGGLZOGZGV5JL3")
		Expect(err).ToNot(HaveOccurred())
		err = gatewayAccount.SetAddress("GC23QF2HUE52AMXUFUH3AYJAXXGXXV2VHXYYR6EYXETPKDXZSAW67XO4")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("InitialState", func() {

		It("errors when `key` does not appear in the bundle", func() {
			_, err := createAccount.InitialState(nonexistantAccount.LedgerKey())
			Expect(err).To(MatchError("meta: no changes found"))
		})

		It("returns nil if `key` gets created within the bundle", func() {
			found, err := createAccount.InitialState(newAccount.LedgerKey())
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeNil())
		})

		It("returns the state if found", func() {
			found, err := createAccount.InitialState(masterAccount.LedgerKey())
			Expect(err).ToNot(HaveOccurred())
			Expect(found).ToNot(BeNil())
			Expect(found.Data.Type).To(Equal(xdr.LedgerEntryTypeAccount))

			account := found.Data.MustAccount().AccountId
			Expect(account.Equals(masterAccount)).To(BeTrue())
		})
	})

	Describe("StateAfter", func() {
		It("returns newly created entries correctly", func() {
			state, err := createAccount.StateAfter(newAccount.LedgerKey(), 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(state).ToNot(BeNil())

			account := state.Data.MustAccount()
			Expect(account.Balance).To(Equal(xdr.Int64(1000000000)))
		})
	})

	Describe("StateBefore", func() {
		Context("Accounts", func() {
			It("return nil when the account was created in the operation", func() {
				state, err := createAccount.StateBefore(newAccount.LedgerKey(), 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(state).To(BeNil())
			})

			It("passes a sanity test", func() {
				before, err := createAccount.StateBefore(masterAccount.LedgerKey(), 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(before).ToNot(BeNil())
				after, err := createAccount.StateAfter(masterAccount.LedgerKey(), 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(after).ToNot(BeNil())
				Expect(before.Data.MustAccount().Balance).To(BeNumerically(">", after.Data.MustAccount().Balance))
			})
		})

		Context("Trustlines", func() {
			var tlkey xdr.LedgerKey
			var line xdr.Asset
			BeforeEach(func() {
				line.SetCredit("USD", gatewayAccount)
				tlkey.SetTrustline(newAccount, line)
			})

			It("properly returns the state of a trustlines that gets removed", func() {
				before, err := removeTrustline.StateBefore(tlkey, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(before).ToNot(BeNil())

				tl := before.Data.MustTrustLine()
				Expect(tl.Limit).To(Equal(xdr.Int64(40000000000)))
			})

			It("properly returns the state of a trustlines that gets removed", func() {
				before, err := updateTrustline.StateBefore(tlkey, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(before).ToNot(BeNil())
				tl := before.Data.MustTrustLine()
				Expect(tl.Limit).To((BeNumerically(">", 40000000000)))
			})
		})
	})
})

func bundle(feeMeta, resultMeta string) (ret Bundle) {
	err := xdr.SafeUnmarshalBase64(feeMeta, &ret.FeeMeta)
	if err != nil {
		panic(err)
	}
	err = xdr.SafeUnmarshalBase64(resultMeta, &ret.TransactionMeta)
	if err != nil {
		panic(err)
	}
	return
}
