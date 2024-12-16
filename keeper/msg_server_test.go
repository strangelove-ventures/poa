package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

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
				Sender: f.authorityAddr,
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

// Verifies that if the minimum commission rate is brought up, all validators with too low of a commission are updated.
// i.e.:
// - MinCommission is 0, change is now 1%, all validators below 1% are updated to 1%.
// - MinCommission is 1%, but now changed back down to 0%. No changes.
func TestUpdateStakingParamChangeValidatorMinCommission(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	// current staking params
	sp, err := f.k.GetStakingKeeper().GetParams(f.ctx)
	require.NoError(err)
	require.True(sp.MinCommissionRate.Equal(sdkmath.LegacyZeroDec()))

	// verify the validator is at the current minimum commission rate (0%)
	vals, err := f.stakingKeeper.GetValidators(f.ctx, 1)
	require.NoError(err)
	require.True(vals[0].Commission.Rate.Equal(sp.MinCommissionRate))

	// increase the min commission rate to 1%
	p := poa.DefaultStakingParams()
	p.MinCommissionRate = sdkmath.LegacyMustNewDecFromStr("0.01")
	_, err = f.msgServer.UpdateStakingParams(f.ctx, &poa.MsgUpdateStakingParams{
		Sender: f.authorityAddr,
		Params: p,
	})
	require.NoError(err)

	// valiadte the validator is now at 1% (the new updated mincommission)
	vals, err = f.stakingKeeper.GetValidators(f.ctx, 1)
	require.NoError(err)
	require.True(vals[0].Commission.Rate.Equal(p.MinCommissionRate))

	// set commission rate back to 0
	p.MinCommissionRate = sdkmath.LegacyZeroDec()
	_, err = f.msgServer.UpdateStakingParams(f.ctx, &poa.MsgUpdateStakingParams{
		Sender: f.authorityAddr,
		Params: p,
	})
	require.NoError(err)

	// no updates made to the validator. It is still at 1%.
	postChangeVals, err := f.stakingKeeper.GetValidators(f.ctx, 1)
	require.NoError(err)
	require.NotEqual(postChangeVals[0].Commission.Rate, p.MinCommissionRate)
}

func TestSetPowerAndCreateValidator(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	totalBonded := sdkmath.ZeroInt()
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
				require.Len(pendingVals.Validators, 1)
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

				require.Len(postVals, len(preVals)+1)
			} else {
				require.EqualValues(len(vals), len(preVals))
			}
		})
	}
}

func TestRemovePending(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	valAddr := f.CreatePendingValidator("val-1", 1_000_000)
	pendingVals, err := f.k.GetPendingValidators(f.ctx)
	require.NoError(err)
	require.Len(pendingVals.Validators, 1)

	testCases := []struct {
		name               string
		request            *poa.MsgRemovePending
		expectErrMsg       string
		expectedPendingLen int
	}{
		{
			name: "fail; not an admin",
			request: &poa.MsgRemovePending{
				Sender:           "foo",
				ValidatorAddress: valAddr.String(),
			},
			expectedPendingLen: 1,
			expectErrMsg:       "not an authority",
		},
		{
			name: "success; removed admin",
			request: &poa.MsgRemovePending{
				Sender:           f.addrs[0].String(),
				ValidatorAddress: valAddr.String(),
			},
			expectedPendingLen: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.RemovePending(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}

			pendingVals, err := f.k.GetPendingValidators(f.ctx)
			require.NoError(err)
			require.Len(pendingVals.Validators, tc.expectedPendingLen)
		})
	}
}

func MustValAddressFromBech32(address string) sdk.ValAddress {
	bz, _ := sdk.ValAddressFromBech32(address)
	return bz
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

	firstVal := vals[0].OperatorAddress

	testCases := []struct {
		name                 string
		request              *poa.MsgRemoveValidator
		isSelfRemovalAllowed bool
		expectErrMsg         string
	}{
		{
			name: "fail; set invalid authority (not an address)",
			request: &poa.MsgRemoveValidator{
				Sender:           "foo",
				ValidatorAddress: firstVal,
			},
			expectErrMsg: "invalid address",
		},
		{
			name: "fail; not from admin or validator",
			request: &poa.MsgRemoveValidator{
				Sender:           f.addrs[1].String(),
				ValidatorAddress: vals[1].OperatorAddress,
			},
			expectErrMsg: poa.ErrNotAnAuthority.Error(),
		},
		{
			name: "success; remove validator as admin",
			request: &poa.MsgRemoveValidator{
				Sender:           f.addrs[0].String(),
				ValidatorAddress: firstVal,
			},
		},
		{
			name: "fail; re-remove same validator as admin",
			request: &poa.MsgRemoveValidator{
				Sender:           f.addrs[0].String(),
				ValidatorAddress: firstVal,
			},
			expectErrMsg: "is not bonded",
		},
		{
			name: "success; remove validator as itself",
			request: &poa.MsgRemoveValidator{
				Sender:           sdk.AccAddress(MustValAddressFromBech32(vals[1].OperatorAddress)).String(),
				ValidatorAddress: vals[1].OperatorAddress,
			},
			isSelfRemovalAllowed: true,
		},
		{
			name: "fail; try again (no longer exist)",
			request: &poa.MsgRemoveValidator{
				Sender:           sdk.AccAddress(MustValAddressFromBech32(vals[1].OperatorAddress)).String(),
				ValidatorAddress: vals[1].OperatorAddress,
			},
			expectErrMsg:         "is not bonded",
			isSelfRemovalAllowed: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err = f.msgServer.RemoveValidator(f.ctx, tc.request)

			if tc.expectErrMsg != "" {
				require.Error(err, tc.expectErrMsg)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)

				_, err := f.IncreaseBlock(3, true)
				require.NoError(err)
			}

			amt, err := f.stakingKeeper.TotalBondedTokens(f.ctx)
			require.NoError(err)
			require.True(amt.IsPositive())

			notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)
			bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
			require.NoError(err)
			bal := f.bankkeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom)
			require.EqualValues(0, bal.Amount.Uint64())

			// BondedRatio
			bondRatio, err := f.stakingKeeper.BondedRatio(f.ctx)
			require.NoError(err)
			require.EqualValues(sdkmath.LegacyOneDec(), bondRatio)
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
