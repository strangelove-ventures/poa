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

	t.Run("default params", func(t *testing.T) {
		data := &poa.GenesisState{
			Params: poa.DefaultParams(),
		}
		err := fixture.k.InitGenesis(fixture.ctx, data)
		require.NoError(err)

		params, err := fixture.k.GetParams(fixture.ctx)
		require.NoError(err)
		require.Equal(poa.DefaultParams(), params)
	})

	// check custom
	t.Run("custom params", func(t *testing.T) {
		p, err := poa.NewParams([]string{fixture.addrs[0].String(), fixture.addrs[1].String()})
		require.NoError(err)

		data := &poa.GenesisState{
			Params: p,
		}
		err = fixture.k.InitGenesis(fixture.ctx, data)
		require.NoError(err)

		params, err := fixture.k.GetParams(fixture.ctx)
		require.NoError(err)
		require.Equal(p, params)
	})

	t.Run("pending validator export", func(t *testing.T) {
		p, err := poa.NewParams([]string{fixture.addrs[0].String(), fixture.addrs[1].String()})
		require.NoError(err)

		acc := GenAcc()
		valAddr := sdk.ValAddress(acc.addr)

		val, err := stakingtypes.NewValidator(valAddr.String(), acc.valKey.PubKey(), stakingtypes.Description{})
		require.NoError(err)

		val.Tokens = sdkmath.NewInt(1_234_567)

		err = fixture.k.InitGenesis(fixture.ctx, &poa.GenesisState{
			Params: p,
		})
		require.NoError(err)

		// load params
		params, err := fixture.k.GetParams(fixture.ctx)
		require.NoError(err)

		require.Equal(p, params)
	})
}
