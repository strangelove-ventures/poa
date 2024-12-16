package keeper

import (
	"context"
	"fmt"

	"github.com/strangelove-ventures/poa"
)

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *poa.GenesisState) error {
	found := make(map[string]bool)
	for _, vals := range data.Vals {
		if found[vals.OperatorAddress] {
			return fmt.Errorf("duplicate validator found in genesis state: %s", vals.OperatorAddress)
		}
		found[vals.OperatorAddress] = true
	}

	if err := k.PendingValidators.Set(ctx, poa.Validators{
		Validators: data.Vals,
	}); err != nil {
		return err
	}

	if err := k.CachedBlockPower.Set(ctx, poa.PowerCache{
		Power: 0,
	}); err != nil {
		return err
	}

	return k.AbsoluteChangedInBlockPower.Set(ctx, poa.PowerCache{
		Power: 0,
	})
}

// InitStores sets the `AbsoluteChangedBlock` and `PreviousBlockPower` as a cache into the poa store.
func (k *Keeper) InitCacheStores(ctx context.Context) error {
	currValPower, err := k.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	// Set the current block (last power) as a H-1 cache.
	if err := k.SetCachedBlockPower(ctx, currValPower.Uint64()); err != nil {
		return err
	}

	if err := k.SetAbsoluteChangedInBlockPower(ctx, 0); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *poa.GenesisState {
	vals, err := k.PendingValidators.Get(ctx)
	if err != nil {
		panic(err)
	}

	return &poa.GenesisState{
		Vals: vals.Validators,
	}
}
