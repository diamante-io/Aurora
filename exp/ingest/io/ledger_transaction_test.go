package io

import (
	"testing"

	"github.com/diamnet/go/xdr"
	"github.com/stretchr/testify/assert"
)

func TestChangeAccountSignersChangedInvalidType(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeOffer,
	}

	assert.Panics(t, func() {
		change.AccountSignersChanged()
	})
}

func TestChangeAccountSignersChangedNoPre(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre:  nil,
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedNoPostMasterKey(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{1, 1, 1, 1},
			},
		},
		Post: nil,
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedNoPostNoMasterKey(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{0, 1, 1, 1},
			},
		},
		Post: nil,
	}

	assert.False(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedMasterKeyRemoved(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{1, 1, 1, 1},
			},
		},
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{0, 1, 1, 1},
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedMasterKeyAdded(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{0, 1, 1, 1},
			},
		},
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				// Master weight = 1
				Thresholds: [4]byte{1, 1, 1, 1},
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedSignerAdded(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers:   []xdr.Signer{},
			},
		},
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers: []xdr.Signer{
					xdr.Signer{
						Key:    xdr.MustSigner("GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX"),
						Weight: 1,
					},
				},
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedSignerRemoved(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers: []xdr.Signer{
					xdr.Signer{
						Key:    xdr.MustSigner("GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX"),
						Weight: 1,
					},
				},
			},
		},
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers:   []xdr.Signer{},
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}

func TestChangeAccountSignersChangedSignerWeightChanged(t *testing.T) {
	change := Change{
		Type: xdr.LedgerEntryTypeAccount,
		Pre: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers: []xdr.Signer{
					xdr.Signer{
						Key:    xdr.MustSigner("GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX"),
						Weight: 1,
					},
				},
			},
		},
		Post: &xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeAccount,
			Account: &xdr.AccountEntry{
				AccountId: xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
				Signers: []xdr.Signer{
					xdr.Signer{
						Key:    xdr.MustSigner("GCCCU34WDY2RATQTOOQKY6SZWU6J5DONY42SWGW2CIXGW4LICAGNRZKX"),
						Weight: 2,
					},
				},
			},
		},
	}

	assert.True(t, change.AccountSignersChanged())
}
