package module

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

// BeginBlocker updates the validator set without applying updates.
// Since this module depends on staking, that module will `ApplyAndReturnValidatorSetUpdates` from x/staking.
func (am AppModule) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(poa.ModuleName, sdkCtx.BlockTime(), telemetry.MetricKeyBeginBlocker)

	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	for _, v := range vals {
		switch v.GetStatus() {
		case stakingtypes.Unbonding:
			// if the validator is unbonding, force it to be unbonded. (H+1)
			v.Status = stakingtypes.Unbonded
			if err := am.keeper.GetStakingKeeper().SetValidator(ctx, v); err != nil {
				return err
			}

		case stakingtypes.Unbonded:
			// if the validator is unbonded (above case), delete the last validator power. (H+2)
			valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
			if err != nil {
				return err
			}

			if err := am.keeper.GetStakingKeeper().DeleteLastValidatorPower(ctx, valAddr); err != nil {
				return err
			}

		case stakingtypes.Unspecified, stakingtypes.Bonded:
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

	return nil
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
