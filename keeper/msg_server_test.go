package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
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

			// add a new validator if the test case requires it
			if tc.createNewValidator {
				val := GenAcc()
				valAddr := sdk.ValAddress(val.addr).String()

				v := poa.ConvertPOAToStaking(CreateNewValidator(
					fmt.Sprintf("val-%s", tc.name),
					valAddr,
					val.valKey.PubKey(),
					int64(tc.request.Power),
				))

				if err := f.k.AddPendingValidator(f.ctx, v, val.valKey.PubKey()); err != nil {
					panic(err)
				}

				f.increaseBlock(1)

				// update the request to include the newly created valAddr
				tc.request.ValidatorAddress = valAddr

				// check the pending validators includes the new validator
				pendingVals, err := f.k.GetPendingValidators(f.ctx)
				require.NoError(err)
				require.EqualValues(1, len(pendingVals.Validators))
			}

			preVals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
			require.NoError(err)

			// sets the power
			_, err = f.msgServer.SetPower(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}

			// check number of vals in the set
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
	f := SetupTest(t)
	require := require.New(t)

	vals, err := f.stakingKeeper.GetValidators(f.ctx, 100)
	require.NoError(err)

	for idx, v := range vals {
		power := 10_000_000
		if idx == 0 {
			power = 1_000_000
		}

		_, err = f.msgServer.SetPower(f.ctx, &poa.MsgSetPower{
			Sender:           f.addrs[0].String(),
			ValidatorAddress: v.OperatorAddress,
			Power:            uint64(power),
			Unsafe:           true,
		})
		require.NoError(err)
	}

	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 5)
	updates, err := f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.ctx)
	require.NoError(err)
	fmt.Printf("%+v", updates)
	require.EqualValues(3, len(updates))

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
			name: "removal",
			request: &poa.MsgRemoveValidator{
				Sender: f.addrs[0].String(),
				// The validator with 1/10th the power
				ValidatorAddress: vals[0].OperatorAddress,
			},
			expectErrMsg: "",
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

				// TODO: move this to SetPower / RemoveValidator checks
				// will require the bankKeeper again
				f.mintTokensToBondedPool(t)

				u, err := f.increaseBlock(1)
				require.NoError(err)
				require.EqualValues(3, len(u))
				fmt.Println(u, err)

				u, err = f.increaseBlock(1)
				require.NoError(err)
				fmt.Println(u)
				require.EqualValues(2, len(u))
			}
		})

	}
}

// TODO: move this to SetPower
func (f *testFixture) mintTokensToBondedPool(t *testing.T) error {
	require := require.New(t)

	// TODO: when RemoveValidator, mint tokens to the bonded pool? (seems we need to, so we can apply and return the set update early)
	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	require.NoError(err)

	validators, err := f.stakingKeeper.GetAllValidators(f.ctx)
	require.NoError(err)

	amt := int64(0)
	for _, v := range validators {
		amt += v.GetBondedTokens().Int64()
	}

	// simplify this to the diff between them
	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(amt)))

	if err := f.bankkeeper.MintCoins(f.ctx, poa.ModuleName, coins); err != nil {
		return err
	}

	if err := f.bankkeeper.SendCoinsFromModuleToModule(f.ctx, poa.ModuleName, types.BondedPoolName, coins); err != nil {
		return err
	}

	return nil
}

func (f *testFixture) increaseBlock(amt int64) ([]abci.ValidatorUpdate, error) {
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + amt)
	updates, err := f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.ctx)
	return updates, err
}
