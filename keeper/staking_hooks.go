package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Before a validator is jailed, we must delete it from the power index. else:
// - ERR CONSENSUS FAILURE!!! err="should never retrieve a jailed validator from the power store"

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wrapper struct for staking keeper
type Hooks struct {
	k Keeper
}

// Hooks return the mesh-security hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// BeforeValidatorModified implements sdk.StakingHooks.
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	h.k.logger.Info("BeforeValidatorModified", "valAddr", valAddr.String())

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// we increase by 5 for when we actually check this in the end blocker
	return h.k.CheckForJailedValidators.Set(ctx, valAddr.String(), sdkCtx.BlockHeight()+5)
}

// BeforeValidatorSlashed implements sdk.StakingHooks.
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	h.k.logger.Info("BeforeValidatorSlashed", valAddr.String(), fraction.String())
	return nil
}

// ----------------------------

// AfterDelegationModified implements sdk.StakingHooks.
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterUnbondingInitiated implements sdk.StakingHooks.
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}

// AfterValidatorBeginUnbonding implements sdk.StakingHooks.
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements sdk.StakingHooks.
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements sdk.StakingHooks.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved implements sdk.StakingHooks.
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationCreated implements sdk.StakingHooks.
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved implements sdk.StakingHooks.
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified implements sdk.StakingHooks.
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
