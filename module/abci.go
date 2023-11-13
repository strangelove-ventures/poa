package module

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (am AppModule) BeginBlocker(ctx context.Context) error {
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

	return nil
}
