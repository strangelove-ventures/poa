package e2e

import (
	"context"
	"testing"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

var (
	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"
	Denom            = "stake"

	POAImage = ibc.DockerImage{
		Repository: "poa",
		Version:    "local",
		UidGid:     "1025:1025",
	}

	defaultGenesis = []cosmos.GenesisKV{
		{
			Key: "app_state.poa.params.admins",
			Value: []string{
				"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // gov
				"cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr", // testing account
			},
		},
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: Denom,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.amount",
			Value: "1",
		},
	}

	POACfg = ibc.ChainConfig{
		Images: []ibc.DockerImage{
			POAImage,
		},
		GasAdjustment:  1.5,
		ModifyGenesis:  cosmos.ModifyGenesis(defaultGenesis),
		EncodingConfig: poaEncoding(),
		Type:           "cosmos",
		Name:           "poa",
		ChainID:        "poa-1",
		Bin:            "poad",
		Bech32Prefix:   "cosmos",
		Denom:          Denom,
		CoinType:       "118",
		GasPrices:      "0" + Denom,
		TrustingPeriod: "330h",
	}
)

func poaEncoding() *moduletestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	poa.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

// assertSignatures asserts that the current block has the exact number of signatures as expected
func assertSignatures(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, expectedSigs int) {
	height, err := chain.GetNode().Height(ctx)
	require.NoError(t, err)
	block := helpers.GetBlockData(t, ctx, chain, height)
	require.Equal(t, expectedSigs, len(block.LastCommit.Signatures))

}

func assertConsensus(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, expected int) {
	cbft := helpers.GetCometBFTConsensus(t, ctx, chain)
	amt := len(cbft.Validators)
	require.EqualValues(t, amt, expected, "expected %d in consensus, got %d", expected, amt)
}
