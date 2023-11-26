package keeper

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

// DefaultPendingValidators returns an empty pending validators set
func DefaultPendingValidators() poa.Validators {
	return poa.Validators{
		Validators: []poa.Validator{},
	}
}

// AddPendingValidator adds a validator to the pending set
func (k Keeper) AddPendingValidator(ctx context.Context, newVal stakingtypes.Validator, pubKey cryptotypes.PubKey) error {
	store := k.storeService.OpenKVStore(ctx)

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return err
	}

	newVal.ConsensusPubkey = pkAny
	stdVal := poa.ConvertStakingToPOA(newVal)

	vals, err := k.GetPendingValidators(ctx)
	if err != nil {
		return err
	}

	vals.Validators = append(vals.Validators, stdVal)

	bz := k.cdc.MustMarshal(&vals)

	return store.Set(poa.PendingValidatorsKey, bz)
}

func (k Keeper) RemovePendingValidator(ctx context.Context, valOpAddr string) error {
	store := k.storeService.OpenKVStore(ctx)

	pending, err := k.GetPendingValidators(ctx)
	if err != nil {
		return err
	}

	vals := pending.Validators

	for i, val := range vals {
		if val.OperatorAddress == valOpAddr {
			vals = append(vals[:i], vals[i+1:]...)
			pending.Validators = vals
			break
		}
	}

	bz := k.cdc.MustMarshal(&pending)

	return store.Set(poa.PendingValidatorsKey, bz)
}

// GetPendingValidators
func (k Keeper) GetPendingValidators(ctx context.Context) (poa.Validators, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(poa.PendingValidatorsKey)
	if err != nil || bz == nil {
		return DefaultPendingValidators(), err
	}

	var pv poa.Validators
	k.cdc.MustUnmarshal(bz, &pv)

	return pv, nil
}

func (k Keeper) GetPendingValidator(ctx context.Context, operatorAddr string) (poa.Validator, error) {
	pending, err := k.GetPendingValidators(ctx)
	if err != nil {
		return poa.Validator{}, err
	}

	for _, val := range pending.Validators {
		if val.OperatorAddress == operatorAddr {
			// required to unpack the pubKey properly
			if err := val.UnpackInterfaces(k.cdc); err != nil {
				return poa.Validator{}, err
			}

			return val, nil
		}
	}

	return poa.Validator{}, nil
}

func (k Keeper) IsValidatorPending(ctx context.Context, operatorAddr string) (bool, error) {
	pending, err := k.GetPendingValidator(ctx, operatorAddr)
	if err != nil {
		return false, err
	}

	return pending.OperatorAddress == operatorAddr, nil
}
