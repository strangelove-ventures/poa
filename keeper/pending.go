package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/poa"
)

// AddPendingValidator adds a validator to the pending set
func (k Keeper) AddPendingValidator(ctx context.Context, valAddr []byte, info string) error {
	return k.PendingValidators.Set(ctx, valAddr, info)
}

func (k Keeper) RemovePendingValidator(ctx context.Context, valAddr sdk.AccAddress) error {
	return k.PendingValidators.Remove(ctx, valAddr)
}

// GetPendingValidators
func (k Keeper) GetPendingValidators(ctx context.Context) ([]poa.PendingValidator, error) {
	iter, err := k.PendingValidators.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	pendingValidators := make([]poa.PendingValidator, 0)

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}

		val, err := iter.Value()
		if err != nil {
			return nil, err
		}

		pendingValidators = append(pendingValidators, poa.PendingValidator{
			Address: key,
			Info:    val,
		})
	}

	return pendingValidators, nil
}

func (k Keeper) GetPendingValidator(ctx context.Context, addr sdk.AccAddress) (poa.PendingValidator, error) {
	info, err := k.PendingValidators.Get(ctx, addr)
	if err != nil {
		return poa.PendingValidator{}, err
	}

	return poa.PendingValidator{
		Address: addr,
		Info:    info,
	}, nil
}

func (k Keeper) IsValidatorPending(ctx context.Context, operator sdk.AccAddress) (bool, error) {
	return k.PendingValidators.Has(ctx, operator)
}
