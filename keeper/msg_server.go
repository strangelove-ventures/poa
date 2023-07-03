package keeper

import (
	"context"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (ms msgServer) CreateValidator(goCtx context.Context, msg *poa.MsgCreateValidator) (*poa.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address as bech32: %w", err)
	}

	if _, found := ms.k.GetValidator(ctx, valAddr); found {
		return nil, sdkerrors.Wrap(poa.ErrBadValidatorAddr, fmt.Sprintf("%s validator already exists: %T", poa.ModuleName, msg))
	}

	validator := &poa.Validator{
		Description: msg.Description,
		Address:     valAddr,
		Pubkey:      msg.Pubkey,
	}

	ms.k.SaveValidator(ctx, validator)

	// Validators vouch for themselves
	ms.k.SetVouch(ctx, &poa.Vouch{
		VoucherAddress:   valAddr,
		CandidateAddress: valAddr,
		InFavor:          true,
	})

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.GetOperator())

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBadValidatorPubKey, err.Error())
	}

	k.AfterValidatorBonded(ctx, consAddr, validator.GetOperator())

	// ctx.EventManager().EmitEvents(sdk.Events{
	// 	sdk.NewEvent(
	// 		stakingtypes.EventTypeCreateValidator,
	// 		sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.Address.String()),
	// 		sdk.NewAttribute(sdk.AttributeKeySender, msg.Owner.String()),
	// 	),
	// })

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgCreateValidatorResponse{}, err

	return nil, nil
}

func (ms msgServer) VouchValidator(context.Context, *poa.MsgVouchValidator) (*poa.MsgVouchValidatorResponse, error) {
	return nil, nil
}

// // IncrementCounter defines the handler for the MsgIncrementCounter message.
// func (ms msgServer) IncrementCounter(ctx context.Context, msg *poa.MsgIncrementCounter) (*poa.MsgIncrementCounterResponse, error) {
// 	if _, err := ms.k.addressCodec.StringToBytes(msg.Sender); err != nil {
// 		return nil, fmt.Errorf("invalid sender address: %w", err)
// 	}

// 	counter, err := ms.k.Counter.Get(ctx, msg.Sender)
// 	if err != nil && !errors.Is(err, collections.ErrNotFound) {
// 		return nil, err
// 	}

// 	counter++

// 	if err := ms.k.Counter.Set(ctx, msg.Sender, counter); err != nil {
// 		return nil, err
// 	}

// 	return &poa.MsgIncrementCounterResponse{}, nil
// }

// // UpdateParams params is defining the handler for the MsgUpdateParams message.
// func (ms msgServer) UpdateParams(ctx context.Context, msg *poa.MsgUpdateParams) (*poa.MsgUpdateParamsResponse, error) {
// 	if _, err := ms.k.addressCodec.StringToBytes(msg.Authority); err != nil {
// 		return nil, fmt.Errorf("invalid authority address: %w", err)
// 	}

// 	if authority := ms.k.GetAuthority(); !strings.EqualFold(msg.Authority, authority) {
// 		return nil, fmt.Errorf("unauthorized, authority does not match the module's authority: got %s, want %s", msg.Authority, authority)
// 	}

// 	if err := msg.Params.Validate(); err != nil {
// 		return nil, err
// 	}

// 	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
// 		return nil, err
// 	}

// 	return &example.MsgUpdateParamsResponse{}, nil
// }
