package keeper

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

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
	if !ms.k.IsAdmin(ctx, msg.Sender) {
		return nil, errorsmod.Wrapf(poa.ErrNotAnAuthority, "sender %s is not an authority. allowed: %+v", msg.Sender, ms.k.GetAdmin(ctx))
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

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	val, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		// validator not found in the set.
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s does not exist", msg.ValidatorAddress)
	}
	// Validator must exist and be bonded for us to set to remove it from the set
	if !val.IsBonded() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s is not bonded", msg.ValidatorAddress)
	}

	// Remove the validator from the active set with 0 consensus power.
	val, err = ms.k.SetPOAPower(ctx, msg.ValidatorAddress, 0)
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

	return &poa.MsgRemovePendingResponse{}, ms.k.RemovePendingValidator(ctx, msg.ValidatorAddress)
}

// CreateValidator is from the x/staking module.
// POA changes:
// - MinSelfDelegation is force set to 1.
// - Create hook logic removed (this is done after acceptance).
// - Valiadtor is added to the pending queue (AddPendingValidator).
func (ms msgServer) CreateValidator(ctx context.Context, msg *poa.MsgCreateValidator) (*poa.MsgCreateValidatorResponse, error) {
	valAddr, err := ms.k.GetValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(ms.k.GetValidatorAddressCodec()); err != nil {
		return nil, err
	}

	minCommRate, err := ms.k.stakingKeeper.MinCommissionRate(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Commission.Rate.LT(minCommRate) {
		return nil, errorsmod.Wrapf(stakingtypes.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", minCommRate)
	}

	// check to see if the pubkey or sender has been registered before
	if _, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr); err == nil {
		return nil, stakingtypes.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "CreateValidator expecting cryptotypes.PubKey, got %T. developer note: make sure to impl codectypes.UnpackInterfacesMessage", pk)
	}

	if _, err := ms.k.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); err == nil {
		return nil, stakingtypes.ErrValidatorPubKeyExists
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true

				break
			}
		}
		if !hasKeyType {
			return nil, errorsmod.Wrapf(
				stakingtypes.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := stakingtypes.NewValidator(msg.ValidatorAddress, pk, stakingtypes.Description{
		Moniker:         msg.Description.Moniker,
		Identity:        msg.Description.Identity,
		Website:         msg.Description.Website,
		SecurityContact: msg.Description.SecurityContact,
		Details:         msg.Description.Details,
	})
	if err != nil {
		return nil, err
	}

	commission := stakingtypes.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, sdkCtx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = sdkmath.NewInt(1)

	// appends the validator to a queue to wait for approval from an admin.
	if err := ms.k.AddPendingValidator(ctx, validator, pk); err != nil {
		return nil, err
	}

	return &poa.MsgCreateValidatorResponse{}, ms.k.UpdateBondedPoolPower(ctx)
}

// UpdateStakingParams wraps the x/staking module's UpdateStakingParams method so that only POA admins can invoke it.
func (ms msgServer) UpdateStakingParams(ctx context.Context, msg *poa.MsgUpdateStakingParams) (*poa.MsgUpdateStakingParamsResponse, error) {
	if ok := ms.k.IsAdmin(ctx, msg.Sender); !ok {
		return nil, poa.ErrNotAnAuthority
	}

	prevStakingParams, err := ms.k.stakingKeeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// https://github.com/liftedinit/cosmos-sdk/blob/a877e3e8048a5acb07a0bff92bd8498cd24d1a01/x/staking/keeper/msg_server.go#L619-L642
	// when min commission rate is updated, we need to update the commission rate of all validators
	if !prevStakingParams.MinCommissionRate.Equal(msg.Params.MinCommissionRate) {
		minRate := msg.Params.MinCommissionRate

		vals, err := ms.k.stakingKeeper.GetAllValidators(ctx)
		if err != nil {
			return nil, err
		}

		blockTime := sdk.UnwrapSDKContext(ctx).BlockHeader().Time

		for _, val := range vals {
			// set the commission rate to min rate
			if val.Commission.CommissionRates.Rate.LT(minRate) {
				val.Commission.CommissionRates.Rate = minRate
				// set the max rate to minRate if it is less than min rate
				if val.Commission.CommissionRates.MaxRate.LT(minRate) {
					val.Commission.CommissionRates.MaxRate = minRate
				}

				val.Commission.UpdateTime = blockTime
				if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
					return nil, fmt.Errorf("failed to set validator after MinCommissionRate param change: %w", err)
				}
			}
		}
	}

	stakingParams := stakingtypes.Params{
		UnbondingTime:     msg.Params.UnbondingTime,
		MaxValidators:     msg.Params.MaxValidators,
		MaxEntries:        msg.Params.MaxEntries,
		HistoricalEntries: msg.Params.HistoricalEntries,
		BondDenom:         msg.Params.BondDenom,
		MinCommissionRate: msg.Params.MinCommissionRate,
	}

	if err := stakingParams.Validate(); err != nil {
		return nil, err
	}

	return &poa.MsgUpdateStakingParamsResponse{}, ms.k.stakingKeeper.SetParams(ctx, stakingParams)
}
