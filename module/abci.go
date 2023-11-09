package module

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (am AppModule) BeginBlocker(ctx context.Context) error {
	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	// TODO: is this needed? / apply to the others?

	if sdk.UnwrapSDKContext(ctx).BlockHeight() == 2 {
		am.updateLastPower(ctx, vals)
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
			am.updateLastPower(ctx, vals)
		}
	}

	return nil
}

func (am AppModule) updateLastPower(ctx context.Context, vals []stakingtypes.Validator) error {
	totalPower := math.ZeroInt()
	for _, v := range vals {
		totalPower = totalPower.Add(v.Tokens)

		valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			return err
		}

		if err := am.keeper.GetStakingKeeper().SetLastValidatorPower(ctx, valAddr, v.Tokens.Int64()); err != nil {
			return err
		}
	}

	if err := am.keeper.GetStakingKeeper().SetLastTotalPower(ctx, totalPower); err != nil {
		return err
	}

	return nil
}
