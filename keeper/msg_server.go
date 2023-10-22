package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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

// ! IMPORTANT: if set to low the chain halts (delegations must be above 1_000_000stake atm)
func (ms msgServer) SetPower(ctx context.Context, msg *poa.MsgSetPower) (*poa.MsgSetPowerResponse, error) {
	if ok, err := ms.isAdmin(ctx, msg.FromAddress); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not an authority")
	}

	if err := msg.Validate(ms.k.validatorAddressCodec); err != nil {
		return nil, err
	}

	isPending, err := ms.k.IsValidatorPending(ctx, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if isPending {
		if err := ms.acceptNewValidator(ctx, msg.ValidatorAddress, msg.Power); err != nil {
			return nil, err
		}
	}

	// if msg.Unsafe is true, we don't check the total power and set it.
	if !msg.Unsafe {
		totalPower, err := ms.getTotalChainPower(ctx)
		if err != nil {
			return nil, err
		}

		if msg.Power > totalPower.Mul(math.NewInt(30)).Quo(math.NewInt(100)).Uint64() {
			return nil, fmt.Errorf("unsafe: msg.Power is >30%% of total power, set unsafe=true to override")
		}
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("ValAddressFromBech32 failed: %w", err)
	}

	val, err := ms.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, fmt.Errorf("GetValidator failed: %w", err)
	}

	// clean delegations up
	delegations, err := ms.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return nil, err
	}
	for _, del := range delegations {
		if err := ms.k.stakingKeeper.RemoveDelegation(ctx, del); err != nil {
			return nil, err
		}
	}

	delegation := stakingtypes.Delegation{
		DelegatorAddress: sdk.AccAddress(valAddr.Bytes()).String(),
		ValidatorAddress: val.OperatorAddress,
		Shares:           math.LegacyNewDec(int64(msg.Power)),
	}

	// TODO: Do not allow setting lower than 1_000_000 ?
	// TODO: does this cause any invariance issues?
	val.DelegatorShares = delegation.Shares
	val.Tokens = math.NewIntFromUint64(msg.Power)
	val.Status = stakingtypes.Bonded

	if err := ms.k.stakingKeeper.SetDelegation(ctx, delegation); err != nil {
		return nil, err
	}

	if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return nil, err
	}

	delegations, err = ms.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	if len(delegations) != 1 {
		return nil, fmt.Errorf("delegation error, expected 1, got %d", len(delegations))
	}

	return &poa.MsgSetPowerResponse{}, nil
}

func (ms msgServer) getTotalChainPower(ctx context.Context) (math.Int, error) {
	delSum := math.ZeroInt()
	allDelegations, err := ms.k.stakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}

	for _, del := range allDelegations {
		delSum = delSum.Add(del.Shares.TruncateInt())
	}

	return delSum, nil
}

func (ms msgServer) RemoveValidator(ctx context.Context, msg *poa.MsgRemoveValidator) (*poa.MsgRemoveValidatorResponse, error) {
	if ok, err := ms.isAdmin(ctx, msg.FromAddress); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("not an authority")
	}

	// Ensure we do not remove the last validator in the set.
	allValidators, err := ms.k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllValidators failed: %w", err)
	}
	if len(allValidators) == 1 {
		return nil, fmt.Errorf("cannot remove the last validator")
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

// CreateValidator implements poa.MsgServer.
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

	validator.MinSelfDelegation = math.NewInt(1)

	// the validator is now pending approval to be let into the set.
	// Until then, they are not apart of the set.
	if err := ms.k.AddPendingValidator(ctx, validator, pk); err != nil {
		return nil, err
	}

	return &poa.MsgCreateValidatorResponse{}, nil
}

// takes in a validator address & sees if they are pending approval.
// if so, we create them now.
// TODO: use stakingtypes.Validator in GetPendingValidator?
func (ms msgServer) acceptNewValidator(ctx context.Context, operatingAddress string, power uint64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	poaVal, err := ms.k.GetPendingValidator(ctx, operatingAddress)
	if err != nil {
		return err
	}

	// ideally we just save the type
	val := poa.ConvertPOAToStaking(poaVal)

	valAddr, err := ms.k.validatorAddressCodec.StringToBytes(val.OperatorAddress) // this is the same as the ValidatorAddress yes?
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

	cons, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	err = ms.k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{
		Address:             sdk.ConsAddress(cons).String(),
		StartHeight:         sdkCtx.BlockHeight(),
		IndexOffset:         0,
		JailedUntil:         sdkCtx.BlockHeader().Time,
		Tombstoned:          false,
		MissedBlocksCounter: 0,
	})
	if err != nil {
		return err
	}

	// call the after-creation hook
	if err := ms.k.stakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return err
	}

	if err := ms.k.RemovePendingValidator(ctx, val.OperatorAddress); err != nil {
		return err
	}

	// TODO: set validator signing info (slashing) - how does staking do it?

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingtypes.EventTypeCreateValidator,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, val.OperatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprintf("%d", power)),
		),
	})

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

	// val.Status = stakingtypes.Unbonding // this removes the validator from singing BUT is not fully unbonded yet.
	val.Tokens = math.ZeroInt()
	val.DelegatorShares = math.LegacyNewDecFromInt(math.ZeroInt())
	if err := ms.k.stakingKeeper.SetValidator(ctx, val); err != nil {
		return fmt.Errorf("SetValidator failed: %w", err)
	}

	// Set their power to 0 they do no longer propose any blocks
	if err := ms.k.stakingKeeper.SetLastValidatorPower(ctx, valAddr, 0); err != nil {
		return fmt.Errorf("SetLastValidatorPower failed: %w", err)
	}

	// clear missed blocks
	cons, err := val.GetConsAddr()
	if err != nil {
		return fmt.Errorf("GetConsAddr failed: %w", err)
	}
	if err := ms.k.slashKeeper.DeleteMissedBlockBitmap(ctx, sdk.ConsAddress(cons)); err != nil {
		return err
	}
	if err := ms.k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{}); err != nil {
		return err
	}

	// Do we handle? or does the sdk do this (may need to wait until the next block?)
	// validator record not found for address: 67AE8730FE9C4A8E67FB699F61EEA7F90627B34F\n
	// if err := ms.k.stakingKeeper.RemoveValidator(ctx, valAddr); err != nil {
	// 	return fmt.Errorf("removevalidator failed: %w", err)
	// }

	return nil
}

func (ms msgServer) isAdmin(ctx context.Context, fromAddr string) (bool, error) {
	for _, auth := range ms.k.GetAdmins(ctx) {
		if auth == fromAddr {
			return true, nil
		}
	}

	return false, nil
}
