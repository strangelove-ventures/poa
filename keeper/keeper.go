package keeper

import (
	"context"

	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/strangelove-ventures/poa"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	stakingKeeper *stakingkeeper.Keeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk *stakingkeeper.Keeper,
) Keeper {
	k := Keeper{
		cdc:           cdc,
		storeService:  storeService,
		stakingKeeper: sk,
	}

	return k
}

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

// GetAdmins returns the module's administrators with delegation of power control.
func (k Keeper) GetAdmins(ctx context.Context) []string {
	p, err := k.GetParams(ctx)
	if err != nil {
		// panic(err) ?
		return []string{}
	}

	return p.Admins
}
