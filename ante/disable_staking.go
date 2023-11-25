package poaante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

type MsgStakingFilterDecorator struct {
}

func NewPOAStakingFilterDecorator() MsgStakingFilterDecorator {
	return MsgStakingFilterDecorator{}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is blocked.
func (msfd MsgStakingFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	currHeight := ctx.BlockHeight()
	if currHeight <= 1 {
		// allow GenTx to pass
		return next(ctx, tx, simulate)
	}

	if msfd.hasInvalidStakingMsg(tx.GetMsgs()) {
		return ctx, poa.ErrStakingActionNotAllowed
	}

	return next(ctx, tx, simulate)
}

func (msfd MsgStakingFilterDecorator) hasInvalidStakingMsg(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *stakingtypes.MsgBeginRedelegate:
			return true
		case *stakingtypes.MsgCancelUnbondingDelegation:
			return true
		case *stakingtypes.MsgCreateValidator: // POA wraps this message.
			return true
		case *stakingtypes.MsgDelegate:
			return true
		// case *stakingtypes.MsgEditValidator: // Allowed
		// 	return true
		case *stakingtypes.MsgUndelegate:
			return true
		case *stakingtypes.MsgUpdateParams: // POA wraps this message.
			return true
		}
	}

	return false
}
