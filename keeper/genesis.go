package keeper

import (
	"context"

	"github.com/strangelove-ventures/poa"
)

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *poa.GenesisState) error {
	return k.SetParams(ctx, data.Params)
}

// InitStores sets the `AbsoluteChangedBlock` and `PreviousBlockPower` as a cache into the poa store.
func (k *Keeper) InitCacheStores(ctx context.Context) error {
	currValPower, err := k.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	// consPower converts the current validator power to a smaller number (divide by 10^6).
	// This allows the power to match the expected consensus power (Int64 / 8) range.
	consPower := currValPower.Quo(k.GetStakingKeeper().PowerReduction(ctx))
	if err := k.SetCachedBlockPower(ctx, consPower.Uint64()); err != nil {
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
