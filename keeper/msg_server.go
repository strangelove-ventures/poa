package keeper

import (
	"context"
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
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

func (ms msgServer) SetPower(ctx context.Context, msg *poa.MsgSetPower) (*poa.MsgSetPowerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if ok := ms.isAdmin(ctx, msg.Sender); !ok {
		return nil, poa.ErrNotAnAuthority
	}

	if err := msg.Validate(ms.k.validatorAddressCodec); err != nil {
		return nil, err
	}

	// accepts a validator into the active set if they are pending approval.
	if isPending, err := ms.k.IsValidatorPending(ctx, msg.ValidatorAddress); err != nil {
		return nil, err
	} else if isPending {
		if err := ms.acceptNewValidator(ctx, msg.ValidatorAddress, msg.Power); err != nil {
			return nil, err
		}
	}

	// sets the new POA power for the validator
	if _, err := ms.updatePOAPower(ctx, msg.ValidatorAddress, int64(msg.Power)); err != nil {
		return nil, err
	}

	// Check that the total power change of the block is not >=30% of the total power of the previous block
	if !msg.Unsafe && sdkCtx.BlockHeight() > 1 {
		// Get Cached GetLastTotalPower
		var cachedPower uint64

		cachedPower, err := ms.k.GetCachedBlockPower(ctx)
		if err != nil {
			return nil, err
		}

		if cachedPower == 0 {
			sdkPower, err := ms.k.stakingKeeper.GetLastTotalPower(ctx)
			if err != nil {
				return nil, err
			}
			cachedPower = sdkPower.Uint64()
		}

		totalChanged, err := ms.k.GetAbsoluteChangedInBlockPower(ctx)
		if err != nil {
			return nil, err
		}

		amt := (totalChanged / cachedPower) * 100
		fmt.Printf("\ntotalChanged: %d, cachedPower: %d, amt: %d\n", totalChanged, cachedPower, amt)
		if amt >= 30 {
			return nil, poa.ErrUnsafePower
		}
	}

	return &poa.MsgSetPowerResponse{}, nil
}

func (ms msgServer) RemoveValidator(ctx context.Context, msg *poa.MsgRemoveValidator) (*poa.MsgRemoveValidatorResponse, error) {
	if ok := ms.isAdmin(ctx, msg.Sender); !ok {
		return nil, poa.ErrNotAnAuthority
	}

	// Ensure we do not remove the last validator in the set.
	allValidators, err := ms.k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}
	if len(allValidators) == 1 {
		return nil, fmt.Errorf("cannot remove the last validator")
	}

	val, err := ms.updatePOAPower(ctx, msg.ValidatorAddress, 0)
	if err != nil {
		return nil, err
	}

	// clear missed blocks (is this needed?)
	cons, err := val.GetConsAddr()
	if err != nil {
		return nil, err
	}
	if err := ms.k.slashKeeper.DeleteMissedBlockBitmap(ctx, sdk.ConsAddress(cons)); err != nil {
		return nil, err
	}
	if err := ms.k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{}); err != nil {
		return nil, err
	}

	return &poa.MsgRemoveValidatorResponse{}, nil
}

// pulled from x/staking
func (ms msgServer) CreateValidator(ctx context.Context, msg *poa.MsgCreateValidator) (*poa.MsgCreateValidatorResponse, error) {
	valAddr, err := ms.k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(ms.k.validatorAddressCodec); err != nil {
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

	return &poa.MsgCreateValidatorResponse{}, nil
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *poa.MsgUpdateParams) (*poa.MsgUpdateParamsResponse, error) {
	if ok := ms.isAdmin(ctx, msg.Sender); !ok {
		return nil, poa.ErrNotAnAuthority
	}

	return &poa.MsgUpdateParamsResponse{}, ms.k.SetParams(ctx, msg.Params)
}

// UpdateStakingParams implements poa.MsgServer.
func (ms msgServer) UpdateStakingParams(ctx context.Context, msg *poa.MsgUpdateStakingParams) (*poa.MsgUpdateStakingParamsResponse, error) {
	if ok := ms.isAdmin(ctx, msg.Sender); !ok {
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

// takes in a validator address & sees if they are pending approval.
func (ms msgServer) acceptNewValidator(ctx context.Context, operatingAddress string, power uint64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get the validator configuration from their CreateValidator message in the past.
	poaVal, err := ms.k.GetPendingValidator(ctx, operatingAddress)
	if err != nil {
		return err
	}

	val := poa.ConvertPOAToStaking(poaVal)

	valAddr, err := ms.k.validatorAddressCodec.StringToBytes(val.OperatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	err = ms.k.stakingKeeper.SetValidator(ctx, val)
	if err != nil {
		return err
	}

	err = ms.k.stakingKeeper.SetValidatorByConsAddr(ctx, val)
	if err != nil {
		return err
	}

	err = ms.k.stakingKeeper.SetNewValidatorByPowerIndex(ctx, val)
	if err != nil {
		return err
	}

	// sets validator slashing defaults (useful for downtime jailing)
	cons, err := val.GetConsAddr()
	if err != nil {
		return err
	}
	if err := ms.k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{
		Address:             sdk.ConsAddress(cons).String(),
		StartHeight:         sdkCtx.BlockHeight(),
		IndexOffset:         0,
		JailedUntil:         sdkCtx.BlockHeader().Time,
		Tombstoned:          false,
		MissedBlocksCounter: 0,
	}); err != nil {
		return err
	}

	if err := ms.k.stakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return err
	}

	if err := ms.k.RemovePendingValidator(ctx, val.OperatorAddress); err != nil {
		return err
	}

	// The validator is actually created now, so emit the necessary events
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingtypes.EventTypeCreateValidator,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, val.OperatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprintf("%d", power)),
		),
	})

	return nil
}

// updatePOAPower removes all delegations, sets a single delegation for POA power, updates the validator with the new shares
// and sets the last validator power to the new value.
func (ms msgServer) updatePOAPower(ctx context.Context, valOpBech32 string, newPower int64) (stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(valOpBech32)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	val, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	previousPower, err := ms.k.stakingKeeper.GetLastValidatorPower(ctx, valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}

	if err := ms.k.stakingKeeper.SetLastValidatorPower(ctx, valAddr, newPower); err != nil {
		return stakingtypes.Validator{}, err
	}

	absPowerDiff := uint64(math.Abs(float64(newPower - previousPower)))

	// print absPowerDiff
	fmt.Printf("\n\n\nvalOpBech32: %s\n", valOpBech32)
	fmt.Printf("New Power: %d\n", newPower)
	fmt.Printf("Prev Power: %d\n", previousPower)
	fmt.Printf("absPowerDiff: %d\n\n\n", absPowerDiff)

	if err := ms.k.IncreaseAbsoluteChangedInBlockPower(ctx, absPowerDiff); err != nil {
		return stakingtypes.Validator{}, err
	}

	if err := ms.updateValidatorSet(ctx, newPower, val, valAddr); err != nil {
		return stakingtypes.Validator{}, err
	}

	return val, nil
}

func (ms msgServer) updateValidatorSet(ctx context.Context, newPower int64, val stakingtypes.Validator, valAddr sdk.ValAddress) error {
	sdkContext := sdk.UnwrapSDKContext(ctx)

	newShare := sdkmath.LegacyNewDec(newPower)
	newShareInt := sdkmath.NewIntFromUint64(uint64(newPower))

	delegation := stakingtypes.Delegation{
		DelegatorAddress: sdk.AccAddress(valAddr.Bytes()).String(),
		ValidatorAddress: val.OperatorAddress,
		Shares:           newShare,
	}
	if err := ms.k.stakingKeeper.SetDelegation(ctx, delegation); err != nil {
		return err
	}

	if newPower == 0 && sdkContext.BlockHeight() > 1 {
		currPower, err := ms.k.stakingKeeper.GetLastValidatorPower(ctx, valAddr)
		if err != nil {
			return err
		}

		val.MinSelfDelegation = sdkmath.NewIntFromUint64(uint64(currPower) + 1)
	}

	val.Tokens = newShareInt
	val.DelegatorShares = newShare
	val.Status = stakingtypes.Bonded
	if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return err
	}

	if err := ms.k.stakingKeeper.SetLastValidatorPower(ctx, valAddr, newPower); err != nil {
		return err
	}

	return ms.updateTotalPower(ctx)
}

func (ms msgServer) updateTotalPower(ctx context.Context) error {
	allVals, err := ms.k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	allTokens := sdkmath.ZeroInt()
	for _, val := range allVals {
		allTokens = allTokens.Add(val.Tokens)
	}

	if err := ms.k.stakingKeeper.SetLastTotalPower(ctx, allTokens); err != nil {
		return err
	}

	return nil
}

func (ms msgServer) isAdmin(ctx context.Context, fromAddr string) bool {
	for _, auth := range ms.k.GetAdmins(ctx) {
		if auth == fromAddr {
			return true
		}
	}

	return false
}
