package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

func TestUpdateParams(t *testing.T) {
	f := SetupTest(t, 2_000_000)
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
			name: "set invalid admins",
			request: &poa.MsgUpdateParams{
				Sender: f.govModAddr,
				Params: poa.Params{},
			},
			expectErrMsg: poa.ErrMustProvideAtLeastOneAddress.Error(),
		},
		{
			name: "set valid params",
			request: &poa.MsgUpdateParams{
				Sender: f.govModAddr,
				Params: poa.Params{
					Admins: []string{f.addrs[0].String()},
				},
			},
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

func TestUpdateStakingParams(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	testCases := []struct {
		name         string
		request      *poa.MsgUpdateStakingParams
		expectErrMsg string
	}{
		{
			name: "set invalid authority (not an address)",
			request: &poa.MsgUpdateStakingParams{
				Sender: "foo",
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "set invalid authority (not defined authority)",
			request: &poa.MsgUpdateStakingParams{
				Sender: f.addrs[1].String(),
				Params: poa.DefaultStakingParams(),
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "set valid params",
			request: &poa.MsgUpdateStakingParams{
				Sender: f.govModAddr,
				Params: poa.DefaultStakingParams(),
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateStakingParams(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestSetPowerAndCreateValidator(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	totalBonded := math.ZeroInt()
	for _, val := range vals {
		totalBonded = totalBonded.Add(val.GetBondedTokens())
	}

	testCases := []struct {
		name               string
		createNewValidator bool
		request            *poa.MsgSetPower
		expectErrMsg       string
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
		{
			name:               "new validator",
			createNewValidator: true,
			request: &poa.MsgSetPower{
				Sender: f.addrs[0].String(),
				Power:  1_000_000,
				Unsafe: true,
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.createNewValidator {
				valAddr := f.CreatePendingValidator(fmt.Sprintf("val-%s", tc.name), tc.request.Power)
				tc.request.ValidatorAddress = valAddr.String()

				// check the pending validators includes the new validator
				pendingVals, err := f.k.GetPendingValidators(f.ctx)
				require.NoError(err)
				require.EqualValues(1, len(pendingVals.Validators))
			}

			require.NotEmpty(tc.request.ValidatorAddress)

			preVals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
			require.NoError(err)

			_, err = f.msgServer.SetPower(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}

			// check number of vals changed the expected amount
			if tc.createNewValidator {
				postVals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
				require.NoError(err)

				require.EqualValues(len(preVals)+1, len(postVals))
			} else {
				require.EqualValues(len(vals), len(preVals))
			}
		})
	}
}

func TestRemoveValidator(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	for _, v := range vals {
		power := 10_000_000

		_, err = f.msgServer.SetPower(f.ctx, &poa.MsgSetPower{
			Sender:           f.addrs[0].String(),
			ValidatorAddress: v.OperatorAddress,
			Power:            uint64(power),
			Unsafe:           true,
		})
		require.NoError(err)
	}

	_, err = f.IncreaseBlock(2, true)
	require.NoError(err)

	testCases := []struct {
		name         string
		request      *poa.MsgRemoveValidator
		expectErrMsg string
	}{
		{
			name: "set invalid authority (not an address)",
			request: &poa.MsgRemoveValidator{
				Sender:           "foo",
				ValidatorAddress: vals[0].OperatorAddress,
			},
			expectErrMsg: "not an authority",
		},
		{
			name: "remove validator",
			request: &poa.MsgRemoveValidator{
				Sender:           f.addrs[0].String(),
				ValidatorAddress: vals[0].OperatorAddress,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.RemoveValidator(f.ctx, tc.request)

			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)

				// This is only required in testing as we do not have a 'real' validator set
				// signing blocks.
				if err := f.mintTokensToBondedPool(t); err != nil {
					panic(err)
				}

				_, err := f.IncreaseBlock(5, true)
				require.NoError(err)
			}
		})
	}
}

func TestMultipleUpdatesInASingleBlock(t *testing.T) {
	f := SetupTest(t, 3_000_000)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	if _, err := f.IncreaseBlock(5, true); err != nil {
		panic(err)
	}

	testCases := []struct {
		name               string
		createNewValidator bool
		request            []*poa.MsgSetPower
		expectedErrIdx     int
		expectErrMsg       string
	}{
		{
			name:               "multiple validator updates",
			createNewValidator: true,
			request: []*poa.MsgSetPower{
				// 11.11%
				{
					Sender:           f.addrs[0].String(),
					ValidatorAddress: vals[0].OperatorAddress,
					Power:            4_000_000,
					Unsafe:           false,
				},
				// 22.22%
				{
					Sender:           f.addrs[0].String(),
					ValidatorAddress: vals[1].OperatorAddress,
					Power:            4_000_000,
					Unsafe:           false,
				},
				// 33.33% modified (>30%, fails if not unsafe)
				{
					Sender:           f.addrs[0].String(),
					ValidatorAddress: vals[2].OperatorAddress,
					Power:            4_000_000,
					Unsafe:           false,
				},
			},
			expectedErrIdx: 2,
			expectErrMsg:   poa.ErrUnsafePower.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, err := f.IncreaseBlock(5); err != nil {
				panic(err)
			}

			for idx, req := range tc.request {
				_, err = f.msgServer.SetPower(f.ctx, req)

				if idx == tc.expectedErrIdx {
					require.Error(err)
					require.ErrorContains(err, tc.expectErrMsg)
				} else {
					require.NoError(err)
				}
			}
		})
	}
}

// mintTokensToBondedPool mints tokens to the bonded pool so the validator set
// in testing can be removed.
// In the future, this same logic would be run during the migration from POA->POS.
func (f *testFixture) mintTokensToBondedPool(t *testing.T) error {
	t.Helper()
	require := require.New(t)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	require.NoError(err)

	validators, err := f.stakingKeeper.GetAllValidators(f.ctx)
	require.NoError(err)

	amt := int64(0)
	for _, v := range validators {
		amt += v.GetBondedTokens().Int64()
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(amt)))

	if err := f.bankkeeper.MintCoins(f.ctx, minttypes.ModuleName, coins); err != nil {
		return err
	}

	if err := f.bankkeeper.SendCoinsFromModuleToModule(f.ctx, minttypes.ModuleName, types.BondedPoolName, coins); err != nil {
		return err
	}

	return nil
}
