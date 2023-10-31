package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func TestUpdateParams(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	testCases := []struct {
		name         string
		request      *poa.MsgUpdateParams
		expectErrMsg string
	}{
		{
			name: "set invalid authority (not an address)",
			request: &poa.MsgUpdateParams{
				Sender: "foo",
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "set invalid authority (not defined authority)",
			request: &poa.MsgUpdateParams{
				Sender: f.addrs[1].String(),
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "set valid params",
			request: &poa.MsgUpdateParams{
				Sender: f.govModAddr,
				Params: poa.Params{},
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestSetPower(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	totalBonded := math.ZeroInt()
	for _, val := range vals {
		totalBonded = totalBonded.Add(val.GetBondedTokens())
	}

	fmt.Println(totalBonded)

	testCases := []struct {
		name         string
		request      *poa.MsgSetPower
		expectErrMsg string
	}{
		{
			name: "set invalid authority (not an address)",
			request: &poa.MsgSetPower{
				Sender:           "foo",
				ValidatorAddress: vals[0].OperatorAddress,
				Power:            1,
				Unsafe:           false,
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "unsafe set",
			request: &poa.MsgSetPower{
				Sender:           f.addrs[0].String(),
				ValidatorAddress: vals[0].OperatorAddress,
				Power:            100_000_000_000,
				Unsafe:           false,
			},
			expectErrMsg: poa.ErrUnsafePower.Error(),
		},
		// TODO: add more test
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.SetPower(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}
