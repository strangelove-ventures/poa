package keeper

import (
	"context"

	"github.com/strangelove-ventures/poa"
)

// SetCachedBlockPower sets the cached consensus power for the current block.
func (k Keeper) SetCachedBlockPower(ctx context.Context, power uint64) error {
	return k.CachedBlockPower.Set(ctx, poa.PowerCache{
		Power: power,
	})
}

// GetCachedBlockPower gets the cached previous block consensus power.
func (k Keeper) GetCachedBlockPower(ctx context.Context) (uint64, error) {
	cached, err := k.CachedBlockPower.Get(ctx)
	if err != nil {
		return 0, nil
	}

	return cached.Power, nil
}

// SetAbsoluteChangedInBlockPower sets the absolute power changed in the current block (relative to last).
func (k Keeper) SetAbsoluteChangedInBlockPower(ctx context.Context, power uint64) error {
	return k.AbsoluteChangedInBlockPower.Set(ctx, poa.PowerCache{
		Power: power,
	})
}

// GetAbsoluteChangedInBlockPower gets the absolute power changed in the block.
func (k Keeper) GetAbsoluteChangedInBlockPower(ctx context.Context) (uint64, error) {
	cached, err := k.AbsoluteChangedInBlockPower.Get(ctx)
	if err != nil {
		return 0, nil
	}

	return cached.Power, nil
}

// IncreaseAbsoluteChangedInBlockPower increases the absolute changed in block power by an amount.
func (k Keeper) IncreaseAbsoluteChangedInBlockPower(ctx context.Context, power uint64) error {
	getAbs, err := k.GetAbsoluteChangedInBlockPower(ctx)
	if err != nil {
		return err
	}

	return k.SetAbsoluteChangedInBlockPower(ctx, getAbs+power)
}
