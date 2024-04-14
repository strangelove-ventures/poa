package module

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

func (am AppModule) EndBlocker(ctx context.Context) error {
	sk := am.keeper.GetStakingKeeper()

	// Front running x/staking maturity ?
	if err := sk.UnbondAllMatureValidators(ctx); err != nil {
		return err
	}

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
		fmt.Printf("UpdatedValidatorsCache: %s\n", valOperAddr)

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

	// get events
	events, err := am.keeper.GetStakingKeeper().GetValidatorUpdates(ctx)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	fmt.Println("\nBeginBlocker events:")
	for _, e := range events {
		fmt.Printf(" %s: %d\n", &e.PubKey, e.Power)
	}
	fmt.Println()

	return nil
}

func (am AppModule) handleBeforeJailedValidators(ctx context.Context) error {
	sk := am.keeper.GetStakingKeeper()

	iterator, err := am.keeper.BeforeJailedValidators.Iterate(ctx, nil)
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
		fmt.Printf("EndBlocker BeforeJailedValidators: %s\n", valOperAddr)

		valAddr, err := sk.ValidatorAddressCodec().StringToBytes(valOperAddr)
		if err != nil {
			return err
		}

		val, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		if err := sk.DeleteValidatorByPowerIndex(ctx, val); err != nil {
			return err
		}

		// TODO: If this is used here, it persist ABCI Updates. When removes, it looks like the validator gets slashed every block in x/staking? (when we do the hack and force set jailed = false)
		// if err := sk.DeleteLastValidatorPower(ctx, valAddr); err != nil {
		// 	return err
		// }

		// !IMPORTANT HACK: Set validator from jailed to not jailed to see what happens
		// Okay so this like kind of worked for a split second
		// Issue: the validator keeps trying to be converted to a jailed validator every single block when x/staking is calling it
		val.Jailed = false
		if err := sk.SetValidator(ctx, val); err != nil {
			return err
		}

		// remove it from persisting
		if err := am.keeper.BeforeJailedValidators.Remove(ctx, valOperAddr); err != nil {
			return err
		}
	}

	return nil
}
