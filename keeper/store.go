package keeper

import (
	"context"
	"encoding/binary"

	"github.com/strangelove-ventures/poa"
)

// SetCachedBlockPower sets the cached consensus power for the current block.
func (k Keeper) SetCachedBlockPower(ctx context.Context, power uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	// convert power to bytes
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, power)

	return store.Set(poa.CachedPreviousBlockPowerKey, bz)
}

// GetCachedBlockPower gets the cached previous block consensus power.
func (k Keeper) GetCachedBlockPower(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(poa.CachedPreviousBlockPowerKey)
	if err != nil || bz == nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(bz), nil
}

// SetAbsoluteChangedInBlockPower sets the absolute power changed in the current block (relative to last).
func (k Keeper) SetAbsoluteChangedInBlockPower(ctx context.Context, power uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	// convert power to bytes
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, power)

	return store.Set(poa.AbsoluteChangedInBlockPowerKey, bz)
}

// GetAbsoluteChangedInBlockPower gets the absolute power changed in the block.
func (k Keeper) GetAbsoluteChangedInBlockPower(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(poa.AbsoluteChangedInBlockPowerKey)
	if err != nil || bz == nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(bz), nil
}

// IncreaseAbsoluteChangedInBlockPower increases the absolute changed in block power by an amount.
func (k Keeper) IncreaseAbsoluteChangedInBlockPower(ctx context.Context, power uint64) error {
	getAbs, err := k.GetAbsoluteChangedInBlockPower(ctx)
	if err != nil {
		return err
	}

	return k.SetAbsoluteChangedInBlockPower(ctx, getAbs+power)
}
