package poaante

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
	poakeeper "github.com/strangelove-ventures/poa/keeper"
)

type MsgStakingFilterDecorator struct {
	PoaKeeper poakeeper.Keeper
}

func NewPOADisableStakingDecorator(pk poakeeper.Keeper) MsgStakingFilterDecorator {
	return MsgStakingFilterDecorator{
		PoaKeeper: pk,
	}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is blocked.
func (msfd MsgStakingFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	currHeight := ctx.BlockHeight()
	if currHeight <= 1 {
		// allow GenTx to pass
		return next(ctx, tx, simulate)
	}

	if err := msfd.hasAdminOnlyStakingMessage(ctx, tx.GetMsgs()); err != nil {
		// return ctx, poa.ErrStakingActionNotAllowed
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (msfd MsgStakingFilterDecorator) hasAdminOnlyStakingMessage(ctx context.Context, msgs []sdk.Msg) error {

	for _, msg := range msgs {
		// authz nested message check (recursive)
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			msgs, err := execMsg.GetMessages()
			if err != nil {
				return nil
			}

			if err := msfd.hasAdminOnlyStakingMessage(ctx, msgs); err != nil {
				return err
			}
		}

		switch m := msg.(type) {
		// on MsgCreateValidator, mint 1 token to the validator on creation
		case *stakingtypes.MsgCreateValidator:
			if m.MinSelfDelegation != sdkmath.NewInt(1) {
				return fmt.Errorf("min self delegation must be 1")
			}

			bondDenom, err := msfd.PoaKeeper.GetStakingKeeper().BondDenom(ctx)
			if err != nil {
				return nil
			}

			if m.Value.Amount.Equal(sdkmath.NewInt(1_000_000)) {
				return fmt.Errorf("self delegation amount must be 1000000" + bondDenom)
			}

			addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
			if err != nil {
				return nil
			}

			// verify they are allowed to create a validator
			isPending, err := msfd.PoaKeeper.IsValidatorPending(ctx, addr)
			if err != nil {
				return err
			} else if !isPending {
				return fmt.Errorf("validator is not pending, can not create validator")
			}

			// mint 1 full token to them
			coin := sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000)))
			if err := msfd.PoaKeeper.GetBankKeeper().MintCoins(ctx, poa.ModuleName, coin); err != nil {
				return err
			}

			if err := msfd.PoaKeeper.GetBankKeeper().SendCoinsFromModuleToAccount(ctx, poa.ModuleName, addr, coin); err != nil {
				return err
			}

			return nil

		case *stakingtypes.MsgUpdateParams:
			return nil

		// Blocked entirely when POA is enabled
		case *stakingtypes.MsgBeginRedelegate,
			*stakingtypes.MsgCancelUnbondingDelegation,
			*stakingtypes.MsgDelegate,
			*stakingtypes.MsgUndelegate:
			return nil
		}

		// stakingtypes.MsgEditValidator is the only allowed message. We do not need to check for it.
	}

	return nil
}
