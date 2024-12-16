package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

func TestInitGenesis(t *testing.T) {
	fixture := SetupTest(t, 2_000_000)
	require := require.New(t)

	t.Run("default state", func(t *testing.T) {
		data := &poa.GenesisState{}
		err := fixture.k.InitGenesis(fixture.ctx, data)
		require.NoError(err)
	})

	t.Run("duplicate validators found in state", func(t *testing.T) {
		data := &poa.GenesisState{
			Vals: []poa.Validator{{
				OperatorAddress: "cosmos1abc",
			}, {
				OperatorAddress: "cosmos1abc",
			}},
		}
		err := fixture.k.InitGenesis(fixture.ctx, data)
		require.Error(err)
	})

	t.Run("pending validator export", func(t *testing.T) {
		acc := GenAcc()
		valAddr := sdk.ValAddress(acc.addr)

		val, err := stakingtypes.NewValidator(valAddr.String(), acc.valKey.PubKey(), stakingtypes.Description{})
		require.NoError(err)

		val.Tokens = sdkmath.NewInt(1_234_567)

		state := &poa.GenesisState{
			Vals: []poa.Validator{poa.ConvertStakingToPOA(val)},
		}

		err = fixture.k.InitGenesis(fixture.ctx, state)
		require.NoError(err)

		exported := fixture.k.ExportGenesis(fixture.ctx)

		require.Len(state.Vals, len(exported.Vals))
		require.Equal(state.Vals[0].OperatorAddress, exported.Vals[0].OperatorAddress)
	})
}
