package poaante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/poa"
)

type MsgStakingFilterDecorator struct {
}

func NewPOADisableStakingDecorator() MsgStakingFilterDecorator {
	return MsgStakingFilterDecorator{}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is blocked.
func (msfd MsgStakingFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	currHeight := ctx.BlockHeight()
	if currHeight <= 1 {
		// allow GenTx to pass
		return next(ctx, tx, simulate)
	}

	err := msfd.hasInvalidStakingMsg(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (msfd MsgStakingFilterDecorator) hasInvalidStakingMsg(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		// authz nested message check (recursive)
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			msgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}

			err = msfd.hasInvalidStakingMsg(msgs)
			if err != nil {
				return err
			}
		}

		switch msg.(type) {
		// POA wrapped messages
		case *stakingtypes.MsgCreateValidator, *stakingtypes.MsgUpdateParams,
			// Blocked entirely when POA is enabled
			*stakingtypes.MsgBeginRedelegate,
			*stakingtypes.MsgCancelUnbondingDelegation,
			*stakingtypes.MsgDelegate,
			*stakingtypes.MsgUndelegate:
			return poa.ErrStakingActionNotAllowed
		}

		// stakingtypes.MsgEditValidator is the only allowed message. We do not need to check for it.
	}

	return nil
}
