package poaante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/strangelove-ventures/poa"
)

type MsgDisableWithdrawDelegatorRewards struct {
}

func NewPOADisableWithdrawDelegatorRewards() MsgDisableWithdrawDelegatorRewards {
	return MsgDisableWithdrawDelegatorRewards{}
}

func (mdwr MsgDisableWithdrawDelegatorRewards) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	currHeight := ctx.BlockHeight()
	if currHeight <= 1 {
		// allow GenTx to pass
		return next(ctx, tx, simulate)
	}

	if mdwr.hasWithdrawDelegatorRewardsMsg(tx.GetMsgs()) {
		return ctx, poa.ErrWithdrawDelegatorRewardsNotAllowed
	}

	return next(ctx, tx, simulate)
}

func (mdwr MsgDisableWithdrawDelegatorRewards) hasWithdrawDelegatorRewardsMsg(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if _, ok := msg.(*distrtypes.MsgWithdrawDelegatorReward); ok {
			return true
		}
	}
	return false
}
