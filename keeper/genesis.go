package keeper

import (
	"context"

	"github.com/cosmosregistry/example"
)

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *example.GenesisState) {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *example.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return &example.GenesisState{
		Params: params,
	}
}
