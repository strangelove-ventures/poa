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
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return err
	}

	newVal.ConsensusPubkey = pkAny
	poaVal := poa.ConvertStakingToPOA(newVal)

	vals, err := k.GetPendingValidators(ctx)
	if err != nil {
		return err
	}

	for _, val := range vals.Validators {
		if val.OperatorAddress == poaVal.OperatorAddress {
			return poa.ErrValidatorAlreadyPending
		}
	}

	vals.Validators = append(vals.Validators, poaVal)

	return k.PendingValidators.Set(ctx, vals)
}

func (k Keeper) RemovePendingValidator(ctx context.Context, valOpAddr string) error {
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

	return k.PendingValidators.Set(ctx, pending)
}

// GetPendingValidators
func (k Keeper) GetPendingValidators(ctx context.Context) (poa.Validators, error) {
	pending, err := k.PendingValidators.Get(ctx)
	if err != nil {
		return DefaultPendingValidators(), err
	}

	return pending, nil
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
