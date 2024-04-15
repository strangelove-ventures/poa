package module

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

func (am AppModule) EndBlocker(ctx context.Context) error {
	if err := am.handleBeforeJailedValidators(ctx); err != nil {
		return err
	}

	return nil
}

// BeginBlocker updates the validator set without applying updates.
// Since this module depends on staking, that module will `ApplyAndReturnValidatorSetUpdates` from x/staking.
func (am AppModule) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(poa.ModuleName, sdkCtx.BlockTime(), telemetry.MetricKeyBeginBlocker)

	iterator, err := am.keeper.UpdatedValidatorsCache.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iterator.Close()

	sk := am.keeper.GetStakingKeeper()

	for ; iterator.Valid(); iterator.Next() {
		valOperAddr, err := iterator.Key()
		if err != nil {
			return err
		}
		am.keeper.Logger().Info("UpdatedValidatorsCache: %s\n", valOperAddr)

		valAddr, err := sk.ValidatorAddressCodec().StringToBytes(valOperAddr)
		if err != nil {
			return err
		}

		val, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		// Remove it from persisting across many blocks
		if err := sk.DeleteValidatorByPowerIndex(ctx, val); err != nil {
			return err
		}

		if err := am.keeper.UpdatedValidatorsCache.Remove(ctx, valOperAddr); err != nil {
			return err
		}
	}

	// reset caches
	if sdkCtx.BlockHeight() > 1 {
		// non gentx messages reset the cached block powers for IBC validations.
		if err := am.keeper.ResetCachedTotalPower(ctx); err != nil {
			return err
		}

		if err := am.keeper.ResetAbsoluteBlockPower(ctx); err != nil {
			return err
		}
	}

	// Event Debugging
	events, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	am.keeper.Logger().Info("BeginBlocker events:\n")
	for _, e := range events {
		e := e
		am.keeper.Logger().Info(fmt.Sprintf("PubKey: %s, Power: %d", &e.PubKey, e.Power))
	}
	am.keeper.Logger().Info("\n")

	return nil
}

func (am AppModule) handleBeforeJailedValidators(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sk := am.keeper.GetStakingKeeper()

	curHeight := sdkCtx.BlockHeight()

	iterator, err := am.keeper.CheckForJailedValidators.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iterator.Close()

	// Why? we don't want it in the store w/ the val state change in x/staking
	for ; iterator.Valid(); iterator.Next() {
		valOperAddr, err := iterator.Key()
		if err != nil {
			return err
		}

		height, err := iterator.Value()
		if err != nil {
			return err
		}

		am.keeper.Logger().Info("EndBlocker BeforeJailedValidators", valOperAddr, height, "\n")

		valAddr, err := sk.ValidatorAddressCodec().StringToBytes(valOperAddr)
		if err != nil {
			return err
		}

		val, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		// !GOAL: Jail a validator properly without `CONSENSUS FAILURE!!! err="should never retrieve a jailed validator from the power store"`  (x/staking/keeper/val_state_change.go) being triggered

		// We only want to perform height logic after jailing stuff has persisted. So we attempt in future blocks.
		// TODO: use x/evidence or slashing instead to pull if the validator is really jailed?
		if height == curHeight {
			fmt.Printf("height: %d, blockHeight: %d\n", height, curHeight)

			if err := sk.DeleteLastValidatorPower(ctx, valAddr); err != nil {
				return err
			}
			if err := sk.DeleteValidatorByPowerIndex(ctx, val); err != nil {
				return err
			}

		} else if height+5 == curHeight {
			// we wait 2 blocks so that the delete last val power has cleared through x/staking
			val.Jailed = true
			if err := sk.SetValidator(ctx, val); err != nil {
				return err
			}
			// issue: still does not like - CONSENSUS FAILURE!!! err="should never retrieve a jailed validator from the power store"
		}

	}

	return nil
}
