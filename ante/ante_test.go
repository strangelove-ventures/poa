package poaante

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
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
		mcl := NewCommissionLimitDecorator(true, tc.rateFloor, tc.rateCeil)

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

func TestAnteNested(t *testing.T) {
	ctx := sdk.Context{}
	ctx = setBlockHeader(ctx, 2)

	const invalidRequestErr = "messages contains *types.Any which is not a sdk.MsgRequest"
	cases := []struct {
		name      string
		decorator sdk.AnteDecorator
		msg       proto.Message
		err       string
	}{
		{
			name:      "fail: commission nested rate < floor",
			decorator: NewCommissionLimitDecorator(true, math.LegacyMustNewDecFromStr("0.10"), math.LegacyMustNewDecFromStr("0.50")),
			msg: &poa.MsgCreateValidator{
				Commission: poa.CommissionRates{
					Rate: math.LegacyMustNewDecFromStr("0.09"),
				},
			},
			err: "rate 0.090000000000000000 is not between 0.100000000000000000 and 0.500000000000000000",
		},
		{
			name:      "fail: commission nested rate > ceil",
			decorator: NewCommissionLimitDecorator(true, math.LegacyMustNewDecFromStr("0.10"), math.LegacyMustNewDecFromStr("0.50")),
			msg: &poa.MsgCreateValidator{
				Commission: poa.CommissionRates{
					Rate: math.LegacyMustNewDecFromStr("0.51"),
				},
			},
			err: "rate 0.510000000000000000 is not between 0.100000000000000000 and 0.500000000000000000",
		},
		{
			name:      "fail: commission nested rate != ceil and floor",
			decorator: NewCommissionLimitDecorator(true, math.LegacyMustNewDecFromStr("0.14"), math.LegacyMustNewDecFromStr("0.14")),
			msg: &poa.MsgCreateValidator{
				Commission: poa.CommissionRates{
					Rate: math.LegacyMustNewDecFromStr("0.1"),
				},
			},
			err: "rate 0.100000000000000000 is not equal to 0.140000000000000000",
		},
		{
			name:      "failed: commission rate msg is nil",
			decorator: NewCommissionLimitDecorator(true, math.LegacyMustNewDecFromStr("0.10"), math.LegacyMustNewDecFromStr("0.50")),
			msg:       nil,
			err:       invalidRequestErr,
		},
		{
			name:      "failed: staking action not allowed",
			decorator: NewPOADisableStakingDecorator(),
			msg:       &stakingtypes.MsgCreateValidator{},
			err:       poa.ErrStakingActionNotAllowed.Error(),
		},
		{
			name:      "failed: staking filter nil msg",
			decorator: NewPOADisableStakingDecorator(),
			msg:       nil,
			err:       invalidRequestErr,
		},
		{
			name:      "failed: withdraw rewards not allowed",
			decorator: NewPOADisableWithdrawDelegatorRewards(),
			msg:       &distrtypes.MsgWithdrawDelegatorReward{},
			err:       poa.ErrWithdrawDelegatorRewardsNotAllowed.Error(),
		},
		{
			name:      "failed: withdraw rewards nil msg",
			decorator: NewPOADisableWithdrawDelegatorRewards(),
			msg:       nil,
			err:       invalidRequestErr,
		},
	}

	for _, tc := range cases {
		tc := tc

		var anyMsg *types.Any
		var err error
		if tc.msg != nil {
			anyMsg, err = types.NewAnyWithValue(tc.msg)
			require.NoError(t, err)
		} else {
			anyMsg = &types.Any{
				TypeUrl: "",
				Value:   nil,
			}
		}

		nestedTx := NewMockTx(&authz.MsgExec{
			Grantee: "",
			Msgs:    []*types.Any{anyMsg},
		})

		_, err = tc.decorator.AnteHandle(ctx, nestedTx, false, EmptyAnte)
		require.Error(t, err)

		if tc.err != "" {
			require.ErrorContains(t, err, tc.err)
		}
	}
}

func TestAnteStakingFilter(t *testing.T) {
	ctx := sdk.Context{}
	sf := NewPOADisableStakingDecorator()

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

func TestAnteDisableWithdrawRewards(t *testing.T) {
	ctx := sdk.Context{}
	dwr := NewPOADisableWithdrawDelegatorRewards()

	blockedMsgs := map[string]sdk.Msg{
		"WithdrawDelegatorReward": &distrtypes.MsgWithdrawDelegatorReward{},
	}

	for k, msg := range blockedMsgs {
		tx := MockTx{
			msgs: []sdk.Msg{
				msg,
			},
		}

		t.Run(fmt.Sprintf("allow GenTx to pass (%s)", k), func(t *testing.T) {
			ctx = setBlockHeader(ctx, 1)
			_, err := dwr.AnteHandle(ctx, tx, false, EmptyAnte)
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("fail: withdraw rewards not allowed after gentx (%s)", k), func(t *testing.T) {
			for h := uint64(2); h < 10; h++ {
				ctx = setBlockHeader(ctx, h)
				_, err := dwr.AnteHandle(ctx, tx, false, EmptyAnte)
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
