package keeper_test

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func TestPendingValidatorsQuery(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// create many validators and query all
	numVals := 10
	for i := 0; i < numVals; i++ {
		f.CreatePendingValidator(fmt.Sprintf("val-%d", i), 1_000_000)
	}

	r, err := f.queryServer.PendingValidators(f.ctx, &poa.QueryPendingValidatorsRequest{})
	require.NoError(err)
	require.EqualValues(numVals, len(r.Pending))

	// get validator 0, SetPower, increase, and query again. There should only be numVals-1 now
	valAddr := r.Pending[0].OperatorAddress
	_, err = f.msgServer.SetPower(f.ctx, &poa.MsgSetPower{
		Sender:           f.addrs[0].String(),
		ValidatorAddress: valAddr,
		Power:            1_000_000,
		Unsafe:           true,
	})
	require.NoError(err)
	if _, err := f.IncreaseBlock(1); err != nil {
		panic(err)
	}

	r, err = f.queryServer.PendingValidators(f.ctx, &poa.QueryPendingValidatorsRequest{})
	require.NoError(err)
	require.EqualValues(numVals-1, len(r.Pending))

	for _, val := range r.Pending {
		require.NotEqual(valAddr, val.OperatorAddress)
	}
}

func TestParamsQuery(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	testCases := []struct {
		name     string
		request  *poa.MsgUpdateParams
		expected poa.Params
	}{
		{
			name: "default",
			request: &poa.MsgUpdateParams{
				Sender: f.addrs[0].String(),
				Params: poa.DefaultParams(),
			},
			expected: poa.DefaultParams(),
		},
		{
			name: "two admins",
			request: &poa.MsgUpdateParams{
				Sender: f.govModAddr,
				Params: poa.Params{
					Admins: []string{f.govModAddr, f.addrs[0].String()},
				},
			},
			expected: poa.Params{
				Admins: []string{f.govModAddr, f.addrs[0].String()},
			},
		},
		{
			name: "duplicate admins",
			request: &poa.MsgUpdateParams{
				Sender: f.govModAddr,
				Params: poa.Params{
					Admins: []string{f.govModAddr, f.govModAddr},
				},
			},
			expected: poa.Params{
				Admins: []string{f.govModAddr, f.govModAddr},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.request)
			require.NoError(err)

			r, err := f.queryServer.Params(f.ctx, &poa.QueryParamsRequest{})
			require.NoError(err)

			require.EqualValues(tc.expected, r.Params)
		})
	}

}
