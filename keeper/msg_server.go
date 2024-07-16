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
	if isAdmin := ms.k.IsAdmin(ctx, msg.Sender); !isAdmin {
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
		params, err := ms.k.GetParams(ctx)
		if err != nil {
			return nil, err
		}

		// check if the sender is the validator being removed.
		hasPermission, err := ms.k.IsSenderValidator(ctx, msg.Sender, msg.ValidatorAddress)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, poa.ErrNotAnAuthority
		}

		if !params.AllowValidatorSelfExit {
			return nil, poa.ErrValidatorSelfRemoval
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

	return &poa.MsgRemovePendingResponse{}, ms.k.RemovePendingValidator(ctx, msg.ValidatorAddress)
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

	stakingParams := stakingtypes.Params{
		UnbondingTime:     msg.Params.UnbondingTime,
		MaxValidators:     msg.Params.MaxValidators,
		MaxEntries:        msg.Params.MaxEntries,
		HistoricalEntries: msg.Params.HistoricalEntries,
		BondDenom:         msg.Params.BondDenom,
		MinCommissionRate: msg.Params.MinCommissionRate,
	}

	return &poa.MsgUpdateStakingParamsResponse{}, ms.k.stakingKeeper.SetParams(ctx, stakingParams)
}
