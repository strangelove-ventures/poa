package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	// "github.com/cosmos/cosmos-sdk/x/staking/types"
	sdkmath "cosmossdk.io/math"

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

func TestRemovePending(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	valAddr := f.CreatePendingValidator("val-1", 1_000_000)
	pendingVals, err := f.k.GetPendingValidators(f.ctx)
	require.NoError(err)
	require.EqualValues(1, len(pendingVals.Validators))

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
			require.EqualValues(tc.expectedPendingLen, len(pendingVals.Validators))
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
			name: "fail; try to remove validator as itself with self removal disabled",
			request: &poa.MsgRemoveValidator{
				Sender:           sdk.AccAddress(MustValAddressFromBech32(vals[1].OperatorAddress)).String(),
				ValidatorAddress: vals[1].OperatorAddress,
			},
			expectErrMsg:         poa.ErrValidatorSelfRemoval.Error(),
			isSelfRemovalAllowed: false,
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
			// Update the params for self approval
			currParams, _ := f.k.GetParams(f.ctx)
			currParams.AllowValidatorSelfExit = tc.isSelfRemovalAllowed

			err = f.k.SetParams(f.ctx, currParams)
			require.NoError(err)

			_, err = f.msgServer.RemoveValidator(f.ctx, tc.request)

			other := f.bankkeeper.GetBalance(f.ctx, authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName), "stake").Amount
			fmt.Println("other1", other)

			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)

				_, err := f.IncreaseBlock(3, true)
				require.NoError(err)
			}

			amt, err := f.stakingKeeper.TotalBondedTokens(f.ctx)
			require.NoError(err)
			fmt.Println("total bonded tokens", amt)

			notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)
			bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
			require.NoError(err)
			bal := f.bankkeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom)
			fmt.Println("notBondedPool", bal.Amount)

			// BondedRatio
			bondRatio, err := f.stakingKeeper.BondedRatio(f.ctx)
			require.NoError(err)
			fmt.Println("bonded ratio", bondRatio)
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
