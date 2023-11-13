package keeper

import (
	"context"

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
