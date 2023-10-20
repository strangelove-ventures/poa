package poa

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ codectypes.UnpackInterfacesMessage = (*Validator)(nil)
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

func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}

func NewCommissionRates(rate, maxRate, maxChangeRate math.LegacyDec) CommissionRates {
	return CommissionRates{
		Rate:          rate,
		MaxRate:       maxRate,
		MaxChangeRate: maxChangeRate,
	}
}

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr string, pubKey cryptotypes.PubKey, description Description, commission CommissionRates, minSelfDelegation math.Int,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateValidator{
		Description:       description,
		ValidatorAddress:  valAddr,
		Pubkey:            pkAny,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (v Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(v.ConsensusPubkey, &pk)
}

// TODO: ideally we should remove this need?
func ConvertPOAToStaking(poa Validator) types.Validator {
	return types.Validator{
		OperatorAddress: poa.OperatorAddress,
		ConsensusPubkey: poa.ConsensusPubkey,
		Jailed:          poa.Jailed,
		Status:          types.BondStatus(poa.Status),
		Tokens:          poa.Tokens,
		DelegatorShares: poa.DelegatorShares,
		Description: types.NewDescription(
			poa.Description.Moniker,
			poa.Description.Identity,
			poa.Description.Website,
			poa.Description.SecurityContact,
			poa.Description.Details,
		),
		UnbondingHeight: poa.UnbondingHeight,
		UnbondingTime:   poa.UnbondingTime,
		Commission: types.NewCommission(
			poa.Commission.CommissionRates.Rate,
			poa.Commission.CommissionRates.MaxRate,
			poa.Commission.CommissionRates.MaxChangeRate,
		),
		MinSelfDelegation:       poa.MinSelfDelegation,
		UnbondingOnHoldRefCount: poa.UnbondingOnHoldRefCount,
		UnbondingIds:            poa.UnbondingIds,
	}
}

func ConvertStakingToPOA(val types.Validator) *Validator {
	return &Validator{
		OperatorAddress: val.OperatorAddress,
		ConsensusPubkey: val.ConsensusPubkey,
		Jailed:          val.Jailed,
		Status:          BondStatus(val.Status),
		Tokens:          val.Tokens,
		DelegatorShares: val.DelegatorShares,
		Description: Description{
			Moniker:         val.Description.Moniker,
			Identity:        val.Description.Identity,
			Website:         val.Description.Website,
			SecurityContact: val.Description.SecurityContact,
			Details:         val.Description.Details,
		},
		UnbondingHeight: val.UnbondingHeight,
		UnbondingTime:   val.UnbondingTime,
		Commission: Commission{
			CommissionRates: CommissionRates{
				Rate:          val.Commission.Rate,
				MaxRate:       val.Commission.MaxRate,
				MaxChangeRate: val.Commission.MaxChangeRate,
			},
		},
		MinSelfDelegation:       val.MinSelfDelegation,
		UnbondingOnHoldRefCount: val.UnbondingOnHoldRefCount,
		UnbondingIds:            val.UnbondingIds,
	}
}
