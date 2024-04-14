package module

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

// We no longer have instant unbonding to support adding new validators into the set.
// This could be tweaked to properly handle jailing / instant unbonding, but not a requirement.
// They stop contributing to BFT consensus anyways.
func (am AppModule) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(poa.ModuleName, sdkCtx.BlockTime(), telemetry.MetricKeyBeginBlocker)

	if sdkCtx.BlockHeight() <= 1 {
		return nil
	}

	valUpdates, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}

	// if len(valUpdates) != 0 {
	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	// TODO: Do we want to run this all the time, or with only val updates of != 0?
	for _, val := range vals {
		// valBz, err := sdk.ValAddressFromBech32(val.OperatorAddress)
		// if err != nil {
		// 	return err
		// }

		if val.IsJailed() {
			// This fixes: "failed to apply block; error commit failed for application: changing validator set: duplicate entry"
			// This is why we require poa to be before staking in the SetOrderBeginBlocker array
			if err := am.keeper.GetStakingKeeper().DeleteValidatorByPowerIndex(ctx, val); err != nil {
				return err
			}
			// TODO: empty set issues on jail
		} else if val.Status == stakingtypes.Bonded && len(valUpdates) > 0 {
			if err := am.keeper.GetStakingKeeper().DeleteValidatorByPowerIndex(ctx, val); err != nil {
				return err
			}
		}
		// }

		// if val.Status == stakingtypes.Bonded {
		// 	lastPower, err := am.keeper.GetStakingKeeper().GetLastValidatorPower(ctx, valBz)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	// This fixes: "failed to apply block; error commit failed for application: changing validator set: duplicate entry"
		// 	// This is why we require poa to be before staking in the SetOrderBeginBlocker array
		// 	if err := am.keeper.GetStakingKeeper().DeleteValidatorByPowerIndex(ctx, val); err != nil {
		// 		return err
		// 	}

		// 	if err := am.keeper.GetStakingKeeper().SetLastValidatorPower(ctx, valBz, lastPower); err != nil {
		// 		return err
		// 	}

		// 	// `SetValidatorByPowerIndex` would forever persist if you do not DeleteValidatorByPowerIndex first.
		// 	// This is used as reference for any future code written as a reminder.
		// 	// Instead, staking handles it for us :)
		// 	// if err := am.keeper.GetStakingKeeper().SetValidatorByPowerIndex(ctx, val); err != nil {
		// 	// 	return err
		// 	// }
		// }
		// }
	}

	return nil
}

// BeginBlocker updates the validator set without applying updates.
// Since this module depends on staking, that module will `ApplyAndReturnValidatorSetUpdates` from x/staking.
func (am AppModule) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(poa.ModuleName, sdkCtx.BlockTime(), telemetry.MetricKeyEndBlocker)

	// vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	// if err != nil {
	// 	return err
	// }

	valUpdates, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}

	for _, valUpdate := range valUpdates {
		am.keeper.Logger().Info("POA EndBlocker ValUpdate", "pubkey", valUpdate.PubKey.String(), "power", valUpdate.Power)
	}

	// for _, v := range vals {
	// 	valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	switch v.GetStatus() {
	// 	case stakingtypes.Unbonding:
	// 		continue
	// 	case stakingtypes.Unbonded:
	// 		// if the validator is unbonded (above case), delete the last validator power. (H+2)
	// 		if err := am.keeper.GetStakingKeeper().DeleteLastValidatorPower(ctx, valAddr); err != nil {
	// 			return err
	// 		}

	// 	case stakingtypes.Unspecified, stakingtypes.Bonded:
	// 		if !v.DelegatorShares.IsZero() {
	// 			// TODO: new val cache still needed?
	// 			// if the validator is freshly created, then perform the validator update.
	// 			isNewVal, err := am.keeper.NewValidatorsCache.Has(ctx, v.GetOperator())
	// 			if err != nil {
	// 				return err
	// 			}

	// 			// power, err := am.keeper.GetStakingKeeper().GetLastValidatorPower(ctx, valAddr)
	// 			// if err != nil {
	// 			// 	return err
	// 			// }

	// 			if isNewVal {
	// 				am.keeper.Logger().Info("New Validator", "operator", v.GetOperator(), "bonded_tokens", v.GetBondedTokens())
	// 				if err := am.keeper.NewValidatorsCache.Remove(ctx, v.GetOperator()); err != nil {
	// 					return err
	// 				}
	// 			}
	// 		}
	// 		continue
	// 	}
	// }

	if sdkCtx.BlockHeight() > 1 {
		// non gentx messages reset the cached block powers for IBC validations.
		if err := am.resetCachedTotalPower(ctx); err != nil {
			return err
		}

		if err := am.resetAbsoluteBlockPower(ctx); err != nil {
			return err
		}
	}

	return err
}

// resetCachedTotalPower resets the block power index to the current total power.
func (am AppModule) resetCachedTotalPower(ctx context.Context) error {
	currValPower, err := am.keeper.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	prev, err := am.keeper.GetCachedBlockPower(ctx)
	if err != nil {
		return err
	}

	if currValPower.Uint64() != prev {
		return am.keeper.SetCachedBlockPower(ctx, currValPower.Uint64())
	}

	return nil
}

// resetAbsoluteBlockPower resets the absolute block power to 0 since updates per block have been executed upon.
func (am AppModule) resetAbsoluteBlockPower(ctx context.Context) error {
	var err error

	val, err := am.keeper.GetAbsoluteChangedInBlockPower(ctx)
	if err != nil {
		return err
	} else if val != 0 {
		return am.keeper.SetAbsoluteChangedInBlockPower(ctx, 0)
	}

	return err
}
