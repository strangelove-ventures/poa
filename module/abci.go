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
	sk := am.keeper.GetStakingKeeper()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	curHeight := sdkCtx.BlockHeight()
	bt := sdkCtx.BlockTime()
	logger := am.keeper.Logger()

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

		// .Error for viewability
		logger.Error("EndBlocker BeforeJailedValidators", "operator", valOperAddr, "height", height)

		valAddr, err := sk.ValidatorAddressCodec().StringToBytes(valOperAddr)
		if err != nil {
			return err
		}

		val, err := sk.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		consBz, err := val.GetConsAddr()
		if err != nil {
			return err
		}

		si, err := am.keeper.GetSlashingKeeper().GetValidatorSigningInfo(ctx, consBz)
		if err != nil {
			return err
		}

		if height == curHeight {
			// if si.JailedUntil.After(bt) is false, then we remove from the keeper store (false positive for the val modification wrt jailing from staking hook)
			if !si.JailedUntil.After(bt) {
				logger.Error("EndBlocker BeforeJailedValidators validator was not jailed, removing from cache",
					"height", height, "blockHeight", curHeight, "si.JailedUntil", si.JailedUntil, "block_time", bt, "operator", valOperAddr,
				)
				if err := am.keeper.CheckForJailedValidators.Remove(ctx, valOperAddr); err != nil {
					return err
				}

				continue
			}

			logger.Error("handleBeforeJailedValidators deleting jailed validator", "height", height, "blockHeight", curHeight)

			if err := sk.DeleteValidatorByPowerIndex(ctx, val); err != nil {
				return err
			}
			if err := sk.DeleteLastValidatorPower(ctx, valAddr); err != nil {
				return err
			}

			val.Jailed = false
			if err := sk.SetValidator(ctx, val); err != nil {
				return err
			}
		} else if height+2 == curHeight {
			// Why is staking / slashing not handling this for us anyways?
			logger.Error("handleBeforeJailedValidators setting val to jailed", "height", height, "blockHeight", curHeight)

			val.Jailed = true
			if err := sk.SetValidator(ctx, val); err != nil {
				return err
			}

			if err := am.keeper.CheckForJailedValidators.Remove(ctx, valOperAddr); err != nil {
				return err
			}
		}
	}

	return nil
}
