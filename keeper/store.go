package keeper

import (
	"context"
	"encoding/binary"

	"github.com/strangelove-ventures/poa"
)

// SetParams sets the module parameters.
func (k Keeper) SetParams(ctx context.Context, p poa.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)

	return store.Set(poa.ParamsKey, bz)
}

// GetParams returns the current module parameters.
func (k Keeper) GetParams(ctx context.Context) (poa.Params, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(poa.ParamsKey)
	if err != nil || bz == nil {
		return poa.DefaultParams(), err
	}

	var p poa.Params
	k.cdc.MustUnmarshal(bz, &p)

	return p, nil
}

// SetCachedBlockPower sets the cached block power.
func (k Keeper) SetCachedBlockPower(ctx context.Context, power uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	// convert power to bytes
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, power)

	return store.Set(poa.CachedPreviousBlockPowerKey, bz)
}

// GetCachedBlockPower gets the cached block power.
func (k Keeper) GetCachedBlockPower(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(poa.CachedPreviousBlockPowerKey)
	if err != nil || bz == nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(bz), nil
}

// SetAbsoluteChangedInBlockPower sets the absolute changed in block power.
func (k Keeper) SetAbsoluteChangedInBlockPower(ctx context.Context, power uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	// convert power to bytes
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, power)

	return store.Set(poa.AbsoluteChangedInBlockPowerKey, bz)
}

// GetAbsoluteChangedInBlockPower gets the absolute changed in block power.
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
