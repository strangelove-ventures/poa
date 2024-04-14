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
		if err := am.keeper.ResetCachedTotalPower(ctx); err != nil {
			return err
		}

		if err := am.keeper.ResetAbsoluteBlockPower(ctx); err != nil {
			return err
		}
	}

	return nil
}
