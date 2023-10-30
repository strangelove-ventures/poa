package poaante

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

var (
	EmptyAnte = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}
)

func TestAnteCommissionRanges(t *testing.T) {
	ctx := sdk.Context{}

	testCases := []struct {
		name       string
		rateFloor  math.LegacyDec
		rateCeil   math.LegacyDec
		commission poa.CommissionRates
		expPass    bool
	}{
		{
			name:      "fail: rate < floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.10"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.01"),
			},
			expPass: false,
		},
		{
			name:      "fail: rate > ceil",
			rateFloor: math.LegacyMustNewDecFromStr("0.10"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.51"),
			},
			expPass: false,
		},
		{
			name:      "fail: rate != ceil and floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.15"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.14"),
			},
			expPass: false,
		},
		{
			name:      "pass: rate == ceil and floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.15"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.15"),
			},
			expPass: true,
		},
		{
			name:      "pass: rate == floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.15"),
			},
			expPass: true,
		},
		{
			name:      "pass: rate == ceil",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.50"),
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		mcl := NewMsgCommissionLimiterDecorator(true, tc.rateFloor, tc.rateCeil)

		// Creating a Validator
		_, err := mcl.AnteHandle(ctx, NewMockTx(&poa.MsgCreateValidator{
			Commission: tc.commission,
		}), false, EmptyAnte)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

		// Editing a Validator
		_, err = mcl.AnteHandle(ctx, NewMockTx(&stakingtypes.MsgEditValidator{
			CommissionRate: &tc.commission.Rate,
		}), false, EmptyAnte)
		if tc.expPass {
			require.NoError(t, err, tc.name+" (edit)")
		} else {
			require.Error(t, err, tc.name+" (edit)")
		}
	}
}

func TestAnteStakingFilter(t *testing.T) {
	ctx := sdk.Context{}
	sf := NewPOAStakingFilterDecorator()

	blockedMsgs := map[string]sdk.Msg{
		"CreateStakingValidator":    &stakingtypes.MsgCreateValidator{},
		"BeginRedelegate":           &stakingtypes.MsgBeginRedelegate{},
		"CancelUnbondingDelegation": &stakingtypes.MsgCancelUnbondingDelegation{},
		"Delegate":                  &stakingtypes.MsgDelegate{},
		"Undelegate":                &stakingtypes.MsgUndelegate{},
		"UpdateParams":              &stakingtypes.MsgUpdateParams{},
	}

	for k, msg := range blockedMsgs {
		tx := MockTx{
			msgs: []sdk.Msg{
				msg,
			},
		}

		t.Run(fmt.Sprintf("allow GenTx to pass (%s)", k), func(t *testing.T) {
			ctx = setBlockHeader(ctx, 1)
			_, err := sf.AnteHandle(ctx, tx, false, EmptyAnte)
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("fail: staking action not allowed after gentx (%s)", k), func(t *testing.T) {
			for h := uint64(2); h < 10; h++ {
				ctx = setBlockHeader(ctx, h)
				_, err := sf.AnteHandle(ctx, tx, false, EmptyAnte)
				require.Error(t, err)
			}
		})
	}
}

func setBlockHeader(ctx sdk.Context, height uint64) sdk.Context {
	h := ctx.BlockHeader()
	h.Height = int64(height)
	return ctx.WithBlockHeader(h)
}

type MockTx struct {
	msgs []sdk.Msg
}

func NewMockTx(msgs ...sdk.Msg) MockTx {
	return MockTx{
		msgs: msgs,
	}
}

func (tx MockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (tx MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}
