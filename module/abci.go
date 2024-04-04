package module

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

	vals, err := am.keeper.GetStakingKeeper().GetAllValidators(ctx)
	if err != nil {
		return err
	}

	valUpdates, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}

	if len(valUpdates) == 0 {
		return nil
	}

	for _, valUpdate := range valUpdates {
		pk := valUpdate.PubKey
		for _, val := range vals {
			valPk, ok := val.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
			if !ok {
				return fmt.Errorf("issue getting consensus pubkey for %s", val.GetOperator())
			}

			if pk.Sum.Compare(valPk.Bytes()) == 1 {
				// delete the validator index
				if err := am.keeper.GetStakingKeeper().DeleteValidatorByPowerIndex(ctx, val); err != nil {
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
	fmt.Printf("\nPOA valUpdates before:")
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

	// remove any duplicate updates from
	rmDups := make(map[string]struct{})
	uniqueValUpdates := []abci.ValidatorUpdate{}
	for _, valUpdate := range valUpdates {
		if _, ok := rmDups[valUpdate.PubKey.String()]; !ok {
			uniqueValUpdates = append(uniqueValUpdates, valUpdate)
			rmDups[valUpdate.PubKey.String()] = struct{}{}
		}
	}

	fmt.Printf("POA valUpdates after:")
	for _, valUpdate := range uniqueValUpdates {
		fmt.Printf(" - %v: %d\n", valUpdate.PubKey.String(), valUpdate.Power)
	}

	if err := am.keeper.GetStakingKeeper().SetValidatorUpdates(ctx, uniqueValUpdates); err != nil {
		return err
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
