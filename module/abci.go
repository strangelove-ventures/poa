package module

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

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
