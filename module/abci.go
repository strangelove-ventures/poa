package module

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (am AppModule) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	for _, v := range vals {
		switch v.GetStatus() {

		case stakingtypes.Unbonding:
			v.Status = stakingtypes.Unbonded
			if err := am.keeper.GetStakingKeeper().SetValidator(ctx, v); err != nil {
				return err
			}

		case stakingtypes.Unbonded:
			valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
			if err != nil {
				return err
			}

			if err := am.keeper.GetStakingKeeper().DeleteLastValidatorPower(ctx, valAddr); err != nil {
				return err
			}
		}
	}

	// if it is not a genTx, reset values to the expected
	if sdkCtx.BlockHeight() > 1 {
		if err := am.resetCachedTotalPower(ctx); err != nil {
			return err
		}

		if err := am.resetAbsoluteBlockPower(ctx); err != nil {
			return err
		}
	}

	return nil
}

// resetCachedTotalPower resets the cached total power to the new TotalPower index.
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
		return am.keeper.SetAbsoluteChangedInBlockPower(ctx, currValPower.Uint64()-prev)
	}

	return nil
}

// resetAbsoluteBlockPower resets the absolute block power to 0.
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
