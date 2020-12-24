package txnbuild

import (
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/xdr"
)

// AccountFlag represents the bitmask flags used to set and clear account authorization options.
type AccountFlag uint32

// AuthRequired is a flag that requires the issuing account to give other accounts
// permission before they can hold the issuing account's credit.
const AuthRequired = AccountFlag(xdr.AccountFlagsAuthRequiredFlag)

// AuthRevocable is a flag that allows the issuing account to revoke its credit
// held by other accounts.
const AuthRevocable = AccountFlag(xdr.AccountFlagsAuthRevocableFlag)

// AuthImmutable is a flag that if set prevents any authorization flags from being
// set, and prevents the account from ever being merged (deleted).
const AuthImmutable = AccountFlag(xdr.AccountFlagsAuthImmutableFlag)

// Threshold is the datatype for MasterWeight, Signer.Weight, and Thresholds. Each is a number
// between 0-255 inclusive.
type Threshold uint8

// Signer represents the Signer in a SetOptions operation.
// If the signer already exists, it is updated.
// If the weight is 0, the signer is deleted.
type Signer struct {
	Address string
	Weight  Threshold
}

// NewHomeDomain is syntactic sugar that makes instantiating SetOptions more convenient.
func NewHomeDomain(hd string) *string {
	return &hd
}

// NewThreshold is syntactic sugar that makes instantiating SetOptions more convenient.
func NewThreshold(t Threshold) *Threshold {
	return &t
}

// NewInflationDestination is syntactic sugar that makes instantiating SetOptions more convenient.
func NewInflationDestination(ai string) *string {
	return &ai
}

// SetOptions represents the DiamNet set options operation. See
// https://www.diamnet.org/developers/guides/concepts/list-of-operations.html
type SetOptions struct {
	InflationDestination *string
	SetFlags             []AccountFlag
	ClearFlags           []AccountFlag
	MasterWeight         *Threshold
	LowThreshold         *Threshold
	MediumThreshold      *Threshold
	HighThreshold        *Threshold
	HomeDomain           *string
	Signer               *Signer
	xdrOp                xdr.SetOptionsOp
	SourceAccount        Account
}

// BuildXDR for SetOptions returns a fully configured XDR Operation.
func (so *SetOptions) BuildXDR() (xdr.Operation, error) {
	err := so.handleInflation()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set inflation destination address")
	}

	so.handleClearFlags()
	so.handleSetFlags()
	so.handleMasterWeight()
	so.handleLowThreshold()
	so.handleMediumThreshold()
	so.handleHighThreshold()
	err = so.handleHomeDomain()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set home domain")
	}
	err = so.handleSigner()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set signer")
	}

	opType := xdr.OperationTypeSetOptions
	body, err := xdr.NewOperationBody(opType, so.xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}

	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, so.SourceAccount)
	return op, nil
}

// handleInflation for SetOptions sets the XDR inflation destination.
// Once set, a new address can be set, but there's no way to ever unset.
func (so *SetOptions) handleInflation() (err error) {
	if so.InflationDestination != nil {
		var xdrAccountID xdr.AccountId
		err = xdrAccountID.SetAddress(*so.InflationDestination)
		if err != nil {
			return
		}
		so.xdrOp.InflationDest = &xdrAccountID
	}
	return
}

// handleSetFlags for SetOptions sets XDR account flags (represented as a bitmask).
// See https://www.diamnet.org/developers/guides/concepts/accounts.html
func (so *SetOptions) handleSetFlags() {
	var flags xdr.Uint32
	for _, flag := range so.SetFlags {
		flags = flags | xdr.Uint32(flag)
	}
	if len(so.SetFlags) > 0 {
		so.xdrOp.SetFlags = &flags
	}
}

// handleClearFlags for SetOptions unsets XDR account flags (represented as a bitmask).
// See https://www.diamnet.org/developers/guides/concepts/accounts.html
func (so *SetOptions) handleClearFlags() {
	var flags xdr.Uint32
	for _, flag := range so.ClearFlags {
		flags = flags | xdr.Uint32(flag)
	}
	if len(so.ClearFlags) > 0 {
		so.xdrOp.ClearFlags = &flags
	}
}

// handleMasterWeight for SetOptions sets the XDR weight of the master signing key.
// See https://www.diamnet.org/developers/guides/concepts/multi-sig.html
func (so *SetOptions) handleMasterWeight() {
	if so.MasterWeight != nil {
		xdrWeight := xdr.Uint32(*so.MasterWeight)
		so.xdrOp.MasterWeight = &xdrWeight
	}
}

// handleLowThreshold for SetOptions sets the XDR value of the account's "low" threshold.
// See https://www.diamnet.org/developers/guides/concepts/multi-sig.html
func (so *SetOptions) handleLowThreshold() {
	if so.LowThreshold != nil {
		xdrThreshold := xdr.Uint32(*so.LowThreshold)
		so.xdrOp.LowThreshold = &xdrThreshold
	}
}

// handleMediumThreshold for SetOptions sets the XDR value of the account's "medium" threshold.
// See https://www.diamnet.org/developers/guides/concepts/multi-sig.html
func (so *SetOptions) handleMediumThreshold() {
	if so.MediumThreshold != nil {
		xdrThreshold := xdr.Uint32(*so.MediumThreshold)
		so.xdrOp.MedThreshold = &xdrThreshold
	}
}

// handleHighThreshold for SetOptions sets the XDR value of the account's "high" threshold.
// See https://www.diamnet.org/developers/guides/concepts/multi-sig.html
func (so *SetOptions) handleHighThreshold() {
	if so.HighThreshold != nil {
		xdrThreshold := xdr.Uint32(*so.HighThreshold)
		so.xdrOp.HighThreshold = &xdrThreshold
	}
}

// handleHomeDomain for SetOptions sets the XDR value of the account's home domain.
// https://www.diamnet.org/developers/guides/concepts/federation.html
func (so *SetOptions) handleHomeDomain() error {
	if so.HomeDomain != nil {
		if len(*so.HomeDomain) > 32 {
			return errors.New("homeDomain must be 32 characters or less")
		}
		xdrHomeDomain := xdr.String32(*so.HomeDomain)
		so.xdrOp.HomeDomain = &xdrHomeDomain
	}

	return nil
}

// handleSigner for SetOptions sets the XDR value of a signer for the account.
// See https://www.diamnet.org/developers/guides/concepts/multi-sig.html
func (so *SetOptions) handleSigner() (err error) {
	if so.Signer != nil {
		var xdrSigner xdr.Signer
		xdrWeight := xdr.Uint32(so.Signer.Weight)
		xdrSigner.Weight = xdrWeight
		err = xdrSigner.Key.SetAddress(so.Signer.Address)
		if err != nil {
			return
		}

		so.xdrOp.Signer = &xdrSigner
	}
	return nil
}
