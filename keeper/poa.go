package keeper

import (
	"context"
	"fmt"
	"math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

// UpdateValidatorSet updates a validator to their new share and consensus power, then updates the total power of the set.
func (k Keeper) UpdateValidatorSet(ctx context.Context, newShares, newConsensusPower int64, val stakingtypes.Validator, valAddr sdk.ValAddress) error {
	newShare := sdkmath.LegacyNewDec(newShares)
	newShareInt := sdkmath.NewIntFromUint64(uint64(newShares))

	delAddr := sdk.AccAddress(valAddr.Bytes())
	delegation := stakingtypes.Delegation{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: val.OperatorAddress,
		Shares:           newShare,
	}

	if err := k.stakingKeeper.SetDelegation(ctx, delegation); err != nil {
		return err
	}

	val.Tokens = newShareInt
	val.DelegatorShares = newShare
	val.Status = stakingtypes.Bonded
	if err := k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return err
	}

	if err := k.stakingKeeper.SetLastValidatorPower(ctx, valAddr, newConsensusPower); err != nil {
		return err
	}

	return k.updateTotalPower(ctx)
}

// SetPower sets a validator's self token delegation and the consensus power for the network.
// Then it updates the total validator set power using this new value.
// It:
// - removes all delegations (for safety)
// - sets a single delegation for POA power
// - updates the validator with the new shares, single delegation
// - sets the last validator power to the new value.
func (k Keeper) SetPOAPower(ctx context.Context, valOpBech32 string, newShares int64) (stakingtypes.Validator, error) {
	// 1 Consenus Power = 1_000_000 shares by default
	newBFTConsensusPower := k.stakingKeeper.TokensToConsensusPower(ctx, sdkmath.NewInt(newShares))

	valAddr, err := sdk.ValAddressFromBech32(valOpBech32)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	val, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	// Get the validators current power before we update it
	currentPower, err := k.stakingKeeper.GetLastValidatorPower(ctx, valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	// no need to process anything, same values
	if newBFTConsensusPower == currentPower {
		return val, fmt.Errorf("current power (%d) is the same as the new power (%d) for %s", currentPower, newBFTConsensusPower, valOpBech32)
	}

	// When we SetValidatorByPowerIndex, the Tokens are used to get the shares of power for CometBFT consensus (voting_power).
	// We don't `k.stakingKeeper.SetValidator` since we only use this for CometBFT consensus power.
	val.Tokens = sdkmath.NewIntFromUint64(uint64(newShares))

	// slash all the validator's tokens (100%)
	if newShares == 0 && currentPower > 0 {
		pk, ok := val.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return stakingtypes.Validator{}, fmt.Errorf("issue getting consensus pubkey for %s", valOpBech32)
		}

		consAddr := sdk.GetConsAddress(pk)

		height := sdk.UnwrapSDKContext(ctx).BlockHeight()

		normalizedToken := k.stakingKeeper.TokensFromConsensusPower(ctx, currentPower)

		if _, err := k.stakingKeeper.Slash(ctx, sdk.GetConsAddress(pk), height, normalizedToken.Int64(), sdkmath.LegacyOneDec()); err != nil {
			return stakingtypes.Validator{}, err
		}
		// TODO:
		if err := k.stakingKeeper.DeleteLastValidatorPower(ctx, valAddr); err != nil {
			return stakingtypes.Validator{}, err
		}
		if err := k.stakingKeeper.DeleteValidatorByPowerIndex(ctx, val); err != nil {
			return stakingtypes.Validator{}, err
		}
		if err := k.slashKeeper.DeleteMissedBlockBitmap(ctx, consAddr); err != nil {
			return stakingtypes.Validator{}, err
		}
	} else {
		// Sets the new consensus power for the validator (this is executed in the x/staking ApplyAndReturnValidatorUpdates method)
		if err := k.GetStakingKeeper().SetLastValidatorPower(ctx, valAddr, newBFTConsensusPower); err != nil {
			return stakingtypes.Validator{}, err
		}
		if err := k.GetStakingKeeper().SetValidatorByPowerIndex(ctx, val); err != nil {
			return stakingtypes.Validator{}, err
		}

		// A cache to handle updated validators power. Once set, the next begin block will remove it from the cache.
		// This allows us to Delete the validator index on the staking side, and ensures power updates do not persist over many blocks.
		// This multi block persistence would break ABCI updates via x/staking if the validator's power is updated again, or removed.
		if err := k.UpdatedValidatorsCache.Set(ctx, val.OperatorAddress); err != nil {
			return stakingtypes.Validator{}, err
		}
	}

	absPowerDiff := uint64(math.Abs(float64(newBFTConsensusPower - currentPower)))

	k.Logger().Debug("POA updatePOAPower",
		"valOpBech32", valOpBech32,
		"New Shares", newShares,
		"New Consensus Power", newBFTConsensusPower,
		"Previous Power", currentPower,
		"absPowerDiff", absPowerDiff,
	)

	if err := k.IncreaseAbsoluteChangedInBlockPower(ctx, absPowerDiff); err != nil {
		return stakingtypes.Validator{}, err
	}

	if err := k.UpdateValidatorSet(ctx, newShares, newBFTConsensusPower, val, valAddr); err != nil {
		return stakingtypes.Validator{}, err
	}

	return val, nil
}

// AcceptNewValidator accepts a new validator and pushes them into the actives set.
func (k Keeper) AcceptNewValidator(ctx context.Context, operatingAddress string, power uint64) error {
	poaVal, err := k.GetPendingValidator(ctx, operatingAddress)
	if err != nil {
		return err
	}

	// convert the pending POA validator into a staking module validator
	val := poa.ConvertPOAToStaking(poaVal)

	// setup the validator into the state and base power
	if err := k.setValidatorInternals(ctx, val); err != nil {
		return err
	}

	if err := k.stakingKeeper.SetNewValidatorByPowerIndex(ctx, val); err != nil {
		return err
	}

	// since the validator is set, remove it from the pending set
	if err := k.RemovePendingValidator(ctx, val.OperatorAddress); err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// set the slashing info for the validator
	if err := k.setSlashingInfo(sdkCtx, val); err != nil {
		return err
	}

	// The validator is actually created now, so emit the necessary events
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingtypes.EventTypeCreateValidator,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, val.OperatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprintf("%d", power)),
		),
	})

	return k.UpdateBondedPoolPower(ctx)
}

// setValidatorInternals sets the validator's:
// - app state
// - consensus state
// - create new power index (no power set yet)
// - Emit `ValidatorCreated` staking hook
func (k Keeper) setValidatorInternals(ctx context.Context, val stakingtypes.Validator) error {
	valAddr, err := k.GetValidatorAddressCodec().StringToBytes(val.OperatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return err
	}

	if err := k.stakingKeeper.SetValidatorByConsAddr(ctx, val); err != nil {
		return err
	}

	return k.stakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr)
}

// UpdateTotalPower sets the new LastTotalPower for the consensus power params.
// It is reduced by the power reduction fraction (default: 10^6) to fit within BFT consensus limits.
func (k Keeper) updateTotalPower(ctx context.Context) error {
	allVals, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	// summation of all validator tokens
	allTokens := sdkmath.ZeroInt()
	for _, val := range allVals {
		allTokens = allTokens.Add(val.Tokens)
	}

	// all tokens / 10^6 = new total power
	totalConsenusPower := allTokens.Quo(k.stakingKeeper.PowerReduction(ctx))
	if err := k.stakingKeeper.SetLastTotalPower(ctx, totalConsenusPower); err != nil {
		return err
	}

	return k.UpdateBondedPoolPower(ctx)
}
