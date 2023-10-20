package poa

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Validate validates the MsgCreateValidator sdk msg.
func (msg MsgCreateValidator) Validate(ac address.Codec) error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := ac.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Pubkey == nil {
		return types.ErrEmptyValidatorPubKey
	}

	// if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
	// 	return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid delegation amount")
	// }

	if msg.Description == (Description{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.Commission == (CommissionRates{}) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty commission")
	}

	if err := msg.Commission.Validate(); err != nil {
		return err
	}

	if !msg.MinSelfDelegation.IsPositive() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"minimum self delegation must be a positive integer",
		)
	}

	// if msg.Value.Amount.LT(msg.MinSelfDelegation) {
	// 	return types.ErrSelfDelegationBelowMinimum
	// }

	return nil
}

// Validate performs basic sanity validation checks of initial commission
// parameters. If validation fails, an SDK error is returned.
func (cr CommissionRates) Validate() error {
	switch {
	case cr.MaxRate.IsNegative():
		// max rate cannot be negative
		return types.ErrCommissionNegative

	case cr.MaxRate.GT(math.LegacyOneDec()):
		// max rate cannot be greater than 1
		return types.ErrCommissionHuge

	case cr.Rate.IsNegative():
		// rate cannot be negative
		return types.ErrCommissionNegative

	case cr.Rate.GT(cr.MaxRate):
		// rate cannot be greater than the max rate
		return types.ErrCommissionGTMaxRate

	case cr.MaxChangeRate.IsNegative():
		// change rate cannot be negative
		return types.ErrCommissionChangeRateNegative

	case cr.MaxChangeRate.GT(cr.MaxRate):
		// change rate cannot be greater than the max rate
		return types.ErrCommissionChangeRateGTMaxRate
	}

	return nil
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, error) {
	if len(d.Moniker) > types.MaxMonikerLength {
		return d, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), types.MaxMonikerLength)
	}

	if len(d.Identity) > types.MaxIdentityLength {
		return d, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), types.MaxIdentityLength)
	}

	if len(d.Website) > types.MaxWebsiteLength {
		return d, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), types.MaxWebsiteLength)
	}

	if len(d.SecurityContact) > types.MaxSecurityContactLength {
		return d, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid security contact length; got: %d, max: %d", len(d.SecurityContact), types.MaxSecurityContactLength)
	}

	if len(d.Details) > types.MaxDetailsLength {
		return d, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), types.MaxDetailsLength)
	}

	return d, nil
}
