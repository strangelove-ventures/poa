package keeper

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/strangelove-ventures/poa"
)

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *poa.GenesisState) error {
	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	for _, val := range data.PendingValidators {
		stakingValidator := poa.ConvertPOAToStaking(val)

		pk, ok := stakingValidator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return fmt.Errorf("expecting cryptotypes.PubKey, got %T", pk)
		}

		if err := k.AddPendingValidator(ctx, stakingValidator, pk); err != nil {
			return err
		}
	}

	return nil
}

// InitStores sets the base cached values from the genesis state in relation to the validator set.
func (k *Keeper) InitCacheStores(ctx context.Context) error {
	currValPower, err := k.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

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

	pendingVals, err := k.GetPendingValidators(ctx)
	if err != nil {
		panic(err)
	}

	return &poa.GenesisState{
		Params:            params,
		PendingValidators: pendingVals.Validators,
	}
}
