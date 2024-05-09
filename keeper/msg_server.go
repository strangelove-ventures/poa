package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"

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

func (ms msgServer) SetPower(ctx context.Context, msg *poa.MsgSetPower) (*poa.MsgSetPowerResponse, error) {
	if isAdmin := ms.k.IsAdmin(ctx, msg.Sender); !isAdmin {
		return nil, errorsmod.Wrapf(poa.ErrNotAnAuthority, "sender %s is not an authority. allowed: %+v", msg.Sender, ms.k.GetAdmins(ctx))
	}

	if err := msg.Validate(ms.k.GetValidatorAddressCodec()); err != nil {
		return nil, err
	}

	// Accept a validator into the active set if they are pending approval.
	if isPending, err := ms.k.IsValidatorPending(ctx, msg.ValidatorAddress); err != nil {
		return nil, err
	} else if isPending {
		if err := ms.k.AcceptNewValidator(ctx, msg.ValidatorAddress, msg.Power); err != nil {
			return nil, err
		}
	}

	// Sets the new POA power to the validator.
	if _, err := ms.k.SetPOAPower(ctx, msg.ValidatorAddress, int64(msg.Power)); err != nil {
		return nil, err
	}

	// Check that the total power change of the block is not >=30% of the total power of the previous block.
	// Transactions tagged `unsafe` will not be checked.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !msg.Unsafe && sdkCtx.BlockHeight() > 1 {
		cachedPower, err := ms.k.GetCachedBlockPower(ctx)
		if err != nil {
			return nil, err
		}

		if cachedPower == 0 {
			return nil, errorsmod.Wrapf(poa.ErrUnsafePower, "cached power is 0 for block %d", sdkCtx.BlockHeight()-1)
		}

		totalChanged, err := ms.k.GetAbsoluteChangedInBlockPower(ctx)
		if err != nil {
			return nil, err
		}

		percent := (totalChanged * 100) / cachedPower

		ms.k.Logger().Debug("POA SetPower",
			"totalChanged", totalChanged,
			"cachedPower", cachedPower,
			"percent", fmt.Sprintf("%d%%", percent),
		)

		if percent >= 30 {
			return nil, poa.ErrUnsafePower
		}
	}

	return &poa.MsgSetPowerResponse{}, ms.k.UpdateBondedPoolPower(ctx)
}

func (ms msgServer) RemoveValidator(ctx context.Context, msg *poa.MsgRemoveValidator) (*poa.MsgRemoveValidatorResponse, error) {
	// Sender is not an admin. Check if the sender is the validator and that validator exist.
	if !ms.k.IsAdmin(ctx, msg.Sender) {
		// check if the sender is the validator being removed.
		hasPermission, err := ms.k.IsSenderValidator(ctx, msg.Sender, msg.ValidatorAddress)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, poa.ErrNotAnAuthority
		}
	}

	vals, err := ms.k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	if len(vals) == 1 {
		return nil, fmt.Errorf("cannot remove the last validator in the set")
	}

	// Ensure the validator exists and is bonded.
	found := false
	for _, val := range vals {
		if val.OperatorAddress == msg.ValidatorAddress {
			if !val.IsBonded() {
				return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s is not bonded", msg.ValidatorAddress)
			}

			found = true
			break
		}
	}

	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s does not exist", msg.ValidatorAddress)
	}

	// Remove the validator from the active set with 0 consensus power.
	val, err := ms.k.SetPOAPower(ctx, msg.ValidatorAddress, 0)
	if err != nil {
		return nil, err
	}

	// Remove missed blocks for the validator.
	if err := ms.k.clearSlashingInfo(ctx, val); err != nil {
		return nil, err
	}

	return &poa.MsgRemoveValidatorResponse{}, ms.k.UpdateBondedPoolPower(ctx)
}

func (ms msgServer) RemovePending(ctx context.Context, msg *poa.MsgRemovePending) (*poa.MsgRemovePendingResponse, error) {
	if isAdmin := ms.k.IsAdmin(ctx, msg.Sender); !isAdmin {
		return nil, errorsmod.Wrapf(poa.ErrNotAnAuthority, "sender %s is not an authority", msg.Sender)
	}

	valAddr, err := ms.k.GetValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	return &poa.MsgRemovePendingResponse{}, ms.k.RemovePendingValidator(ctx, valAddr)
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *poa.MsgUpdateParams) (*poa.MsgUpdateParamsResponse, error) {
	if ok := ms.k.IsAdmin(ctx, msg.Sender); !ok {
		return nil, poa.ErrNotAnAuthority
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	return &poa.MsgUpdateParamsResponse{}, ms.k.SetParams(ctx, msg.Params)
}
