package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

type msgServer struct {
	k Keeper
}

var _ poa.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) poa.MsgServer {
	return &msgServer{k: keeper}
}

// ! IMPORTANT: if set to low the chain halts (delegations must be above 1_000_000stake atm)
func (ms msgServer) SetPower(ctx context.Context, msg *poa.MsgSetPower) (*poa.MsgSetPowerResponse, error) {
	if ok, err := ms.isAdmin(ctx, msg.FromAddress); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not an authority")
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("ValAddressFromBech32 failed: %w", err)
	}

	val, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, fmt.Errorf("GetValidator failed: %w", err)
	}

	delegations, err := ms.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// this should never happen, make sure of it even if something goes wrong
	if len(delegations) != 1 {
		return nil, fmt.Errorf("delegations should only be len of 1: %+v", delegations)
	}

	del := delegations[0]
	decAmt := math.LegacyNewDecFromInt(math.NewIntFromUint64(msg.Power))

	// TODO: Do not allow setting lower than 1_000_000 ?
	// TODO: does this cause any invariance issues?
	del.Shares = decAmt
	val.DelegatorShares = decAmt
	val.Tokens = math.NewIntFromUint64(msg.Power)

	if err := ms.k.stakingKeeper.SetDelegation(ctx, del); err != nil {
		return nil, err
	}

	if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return nil, err
	}

	return &poa.MsgSetPowerResponse{}, nil
}

func (ms msgServer) RemoveValidator(ctx context.Context, msg *poa.MsgRemoveValidator) (*poa.MsgRemoveValidatorResponse, error) {
	if ok, err := ms.isAdmin(ctx, msg.FromAddress); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not an authority")
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("ValAddressFromBech32 failed: %w", err)
	}

	if err := ms.clearValidator(ctx, valAddr); err != nil {
		return nil, fmt.Errorf("clearValidator failed: %w", err)
	}

	return &poa.MsgRemoveValidatorResponse{}, nil
}

func (ms msgServer) clearValidator(ctx context.Context, valAddr sdk.ValAddress) error {
	val, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("GetValidator failed: %w", err)
	}

	delegations, err := ms.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("GetValidatorDelegations failed: %w", err)
	}

	for _, del := range delegations {
		if err := ms.k.stakingKeeper.RemoveDelegation(ctx, del); err != nil {
			return fmt.Errorf("RemoveDelegation failed: %w", err)
		}
	}

	val.Status = stakingtypes.Unbonded
	val.Tokens = math.ZeroInt()
	val.DelegatorShares = math.LegacyNewDecFromInt(math.ZeroInt())
	if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return fmt.Errorf("SetValidator failed: %w", err)
	}

	// Do we handle? or does the sdk do this (may need to wait until the next block?)
	// validator record not found for address: 67AE8730FE9C4A8E67FB699F61EEA7F90627B34F\n
	// if err := ms.k.stakingKeeper.RemoveValidator(ctx, valAddr); err != nil {
	// 	return fmt.Errorf("removevalidator failed: %w", err)
	// }

	return nil
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *poa.MsgUpdateParams) (*poa.MsgUpdateParamsResponse, error) {
	if ok, err := ms.isAdmin(ctx, msg.FromAddress); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not an authority")
	}

	return &poa.MsgUpdateParamsResponse{}, ms.k.SetParams(ctx, msg.Params)
}

func (ms msgServer) isAdmin(ctx context.Context, fromAddr string) (bool, error) {
	for _, auth := range ms.k.GetAdmins(ctx) {
		if auth == fromAddr {
			return true, nil
		}
	}

	return false, nil
}
