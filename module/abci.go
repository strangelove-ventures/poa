package module

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

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

	if len(valUpdates) != 0 {
		vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
		if err != nil {
			return err
		}

		for _, val := range vals {
			if val.Status == stakingtypes.Bonded {
				valBz, err := sdk.ValAddressFromBech32(val.OperatorAddress)
				if err != nil {
					return err
				}

				lastPower, err := am.keeper.GetStakingKeeper().GetLastValidatorPower(ctx, valBz)
				if err != nil {
					return err
				}

				// This fixes: "failed to apply block; error commit failed for application: changing validator set: duplicate entry"
				if err := am.keeper.GetStakingKeeper().DeleteValidatorByPowerIndex(ctx, val); err != nil {
					return err
				}

				if err := am.keeper.GetStakingKeeper().SetLastValidatorPower(ctx, valBz, lastPower); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// BeginBlocker updates the validator set without applying updates.
// Since this module depends on staking, that module will `ApplyAndReturnValidatorSetUpdates` from x/staking.
func (am AppModule) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(poa.ModuleName, sdkCtx.BlockTime(), telemetry.MetricKeyEndBlocker)

	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	valUpdates, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("\nENDBLOCK POA valUpdates before:\n")
	for _, valUpdate := range valUpdates {
		fmt.Printf(" - %v: %d\n", valUpdate.PubKey.String(), valUpdate.Power)
	}

	for _, v := range vals {
		valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			return err
		}

		switch v.GetStatus() {
		case stakingtypes.Unbonding:
			// if the validator is unbonding, force it to be unbonded. (H+1)
			v.Status = stakingtypes.Unbonded
			if err := am.keeper.GetStakingKeeper().SetValidator(ctx, v); err != nil {
				return err
			}

		case stakingtypes.Unbonded:
			// if the validator is unbonded (above case), delete the last validator power. (H+2)
			if err := am.keeper.GetStakingKeeper().DeleteLastValidatorPower(ctx, valAddr); err != nil {
				return err
			}

		case stakingtypes.Unspecified, stakingtypes.Bonded:
			if !v.DelegatorShares.IsZero() {
				// if the validator is freshly created, then perform the validator update.

				// TODO: get last validator power from x/staking here instead? (then we can remove the cache)
				isNewVal, err := am.keeper.NewValidatorsCache.Has(ctx, v.GetOperator())
				if err != nil {
					return err
				}

				power, err := am.keeper.GetStakingKeeper().GetLastValidatorPower(ctx, valAddr)
				fmt.Println("\nisNewVal", isNewVal)
				fmt.Println("power", power)
				fmt.Println("err", err)

				if isNewVal {
					if err := am.keeper.NewValidatorsCache.Remove(ctx, v.GetOperator()); err != nil {
						return err
					}
				}
			}
			continue
		}
	}

	if sdkCtx.BlockHeight() > 1 {
		// non gentx messages reset the cached block powers for IBC validations.
		if err := am.resetCachedTotalPower(ctx); err != nil {
			return err
		}

		if err := am.resetAbsoluteBlockPower(ctx); err != nil {
			return err
		}
	}

	fmt.Printf("POA valUpdates after:\n")
	for _, valUpdate := range valUpdates {
		fmt.Printf(" - %v: %d\n", valUpdate.PubKey.String(), valUpdate.Power)
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
