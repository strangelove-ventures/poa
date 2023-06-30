package keeper_test

import (
	"testing"

	"github.com/cosmosregistry/example"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	resp, err := f.queryServer.Params(f.ctx, &example.QueryParamsRequest{})
	require.NoError(err)
	require.Equal(example.Params{}, resp.Params)
}

func TestQueryCounter(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	resp, err := f.queryServer.Counter(f.ctx, &example.QueryCounterRequest{Address: f.addrs[0].String()})
	require.NoError(err)
	require.Equal(uint64(0), resp.Counter)

	_, err = f.msgServer.IncrementCounter(f.ctx, &example.MsgIncrementCounter{Sender: f.addrs[0].String()})
	require.NoError(err)

	resp, err = f.queryServer.Counter(f.ctx, &example.QueryCounterRequest{Address: f.addrs[0].String()})
	require.NoError(err)
	require.Equal(uint64(1), resp.Counter)
}
