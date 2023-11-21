package keeper_test

import (
	"testing"

	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	fixture := SetupTest(t, 2_000_000)

	t.Run("default params", func(t *testing.T) {
		data := &poa.GenesisState{
			PendingValidators: []poa.Validator{},
			Params:            poa.DefaultParams(),
		}
		err := fixture.k.InitGenesis(fixture.ctx, data)
		require.NoError(t, err)

		params, err := fixture.k.GetParams(fixture.ctx)
		require.NoError(t, err)
		require.Equal(t, poa.DefaultParams(), params)
	})

	// check custom
	t.Run("custom params", func(t *testing.T) {
		p, err := poa.NewParams([]string{fixture.addrs[0].String(), fixture.addrs[1].String()})
		require.NoError(t, err)

		data := &poa.GenesisState{
			PendingValidators: []poa.Validator{},
			Params:            p,
		}
		err = fixture.k.InitGenesis(fixture.ctx, data)
		require.NoError(t, err)

		params, err := fixture.k.GetParams(fixture.ctx)
		require.NoError(t, err)
		require.Equal(t, p, params)
	})

	// TODO: PendingValidators
}
