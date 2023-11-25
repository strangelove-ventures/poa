package keeper

import (
	"context"

	"github.com/strangelove-ventures/poa"
)

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *poa.GenesisState) error {
	return k.SetParams(ctx, data.Params)
}

// InitStores sets the base cached values from the genesis state in relation to the validator set.
func (k *Keeper) InitCacheStores(ctx context.Context) error {
	// currValPower will be 0 if height is 0 or 1 (gentx)
	currValPower, err := k.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	currValPower = currValPower.Quo(k.GetStakingKeeper().PowerReduction(ctx))

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
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &poa.GenesisState{
		Params: params,
	}
}
