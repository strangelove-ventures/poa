package poaante

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/poa"
)

// MsgCommissionLimiterDecorator limits commission rates for validators between 2 ranges.
// if both ranges are the same, the rate change must only be the value both.
type MsgCommissionLimiterDecorator struct {
	// if true, gentxs are also required to be within the commission limit on network start.
	DoGenTxRateValidation bool

	// the validator's set commission rate.
	RateFloor math.LegacyDec
	RateCeil  math.LegacyDec
}

func NewMsgCommissionLimiterDecorator(doGenTxRateValidation bool, rateFloor, rateCiel math.LegacyDec) MsgCommissionLimiterDecorator {
	return MsgCommissionLimiterDecorator{
		DoGenTxRateValidation: doGenTxRateValidation,
		RateFloor:             rateFloor,
		RateCeil:              rateCiel,
	}
}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message that is not within the commission limit.
func (mcl MsgCommissionLimiterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !mcl.DoGenTxRateValidation && ctx.BlockHeight() <= 1 {
		return next(ctx, tx, simulate)
	}

	err := mcl.hasInvalidCommissionRange(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (mcl MsgCommissionLimiterDecorator) hasInvalidCommissionRange(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		// authz nested message check (recursive)
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			msgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}

			err = mcl.hasInvalidCommissionRange(msgs)
			if err != nil {
				return err
			}
		}

		switch msg := msg.(type) {
		// Create Validator POA wrapper
		case *poa.MsgCreateValidator:
			return rateCheck(msg.Commission.Rate, mcl.RateFloor, mcl.RateCeil)
		// Editing the validator through staking (no POA edit)
		case *stakingtypes.MsgEditValidator:
			return rateCheck(*msg.CommissionRate, mcl.RateFloor, mcl.RateCeil)
		}

	}
	return nil
}

func rateCheck(source math.LegacyDec, low math.LegacyDec, high math.LegacyDec) error {
	if low.Equal(high) && !source.Equal(low) {
		return fmt.Errorf("rate %v is not equal to %v", source, low)
	}

	if source.GT(high) || source.LT(low) {
		return fmt.Errorf("rate %v is not between %v and %v", source, low, high)
	}

	return nil
}
