package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/math"
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

// BeforeValidatorModified implements types.StakingHooks.
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr types.ValAddress) error {
	h.k.logger.Info("BeforeValidatorModified", "valAddr", valAddr.String())
	return h.k.BeforeJailedValidators.Set(ctx, valAddr.String())
}

// BeforeValidatorSlashed implements types.StakingHooks.
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr types.ValAddress, fraction math.LegacyDec) error {
	h.k.logger.Info("BeforeValidatorSlashed", valAddr.String(), fraction.String())
	return nil
}

// ----------------------------

// AfterDelegationModified implements types.StakingHooks.
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterUnbondingInitiated implements types.StakingHooks.
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}

// AfterValidatorBeginUnbonding implements types.StakingHooks.
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorBonded implements types.StakingHooks.
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorCreated implements types.StakingHooks.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr types.ValAddress) error {
	return nil
}

// AfterValidatorRemoved implements types.StakingHooks.
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr types.ConsAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationCreated implements types.StakingHooks.
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved implements types.StakingHooks.
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified implements types.StakingHooks.
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr types.AccAddress, valAddr types.ValAddress) error {
	return nil
}
