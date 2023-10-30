package poaante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/poa"
	poakeeper "github.com/strangelove-ventures/poa/keeper"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type MsgStakingFilterDecorator struct {
	poaKeeper poakeeper.Keeper
}

func NewPOAStakingFilterDecorator(poaKeeper poakeeper.Keeper) MsgStakingFilterDecorator {
	return MsgStakingFilterDecorator{
		poaKeeper: poaKeeper,
	}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is blocked.
func (msfd MsgStakingFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	currHeight := ctx.BlockHeight()
	if currHeight <= 1 {
		// allow GenTx to pass
		return next(ctx, tx, simulate)
	}

	invalid, err := msfd.hasInvalidStakingMsg(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	if invalid {
		return ctx, poa.ErrStakingActionNotAllowed
	}

	return next(ctx, tx, simulate)
}

func (msfd MsgStakingFilterDecorator) hasInvalidStakingMsg(msgs []sdk.Msg) (bool, error) {
	for _, msg := range msgs {
		switch msg.(type) {
		case *stakingtypes.MsgBeginRedelegate:
			return true, nil
		case *stakingtypes.MsgCancelUnbondingDelegation:
			return true, nil
		case *stakingtypes.MsgCreateValidator: // POA wraps this message.
			return true, nil
		case *stakingtypes.MsgDelegate:
			return true, nil
			// case *stakingtypes.MsgEditValidator: // Allowed
			// 	return true, nil
		case *stakingtypes.MsgUndelegate:
			return true, nil
		case *stakingtypes.MsgUpdateParams: // POA wraps this message.
			return true, nil
		}
	}
	return false, nil
}
