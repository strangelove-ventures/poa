package e2e

import (
	"fmt"
	"testing"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

const (
	numNodes = 0
	numVals  = 1
)

func TestPOAJailing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	updatedSlashingCfg := POACfg.Clone()

	updatedSlashingCfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesis, []cosmos.GenesisKV{
		{
			Key:   "app_state.slashing.params.signed_blocks_window",
			Value: "10",
		},
		{
			Key:   "app_state.slashing.params.min_signed_per_window",
			Value: "1.000000000000000000",
		},
		{
			Key:   "app_state.slashing.params.downtime_jail_duration",
			Value: "600s",
		},
		{
			Key:   "app_state.slashing.params.slash_fraction_double_sign",
			Value: "0.000000000000000000",
		},
		{
			Key:   "app_state.slashing.params.slash_fraction_downtime",
			Value: "0.000000000000000000",
		},
	}...))

	// setup base chain
	chainVals := 5
	chains := interchaintest.CreateChainWithConfig(t, chainVals, numNodes, "poa", "", updatedSlashingCfg)
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, false)

	chain := chains[0].(*cosmos.CosmosChain)

	// Verify all validators are signing as expected
	validators := helpers.GetValidatorsOperatingAddresses(t, ctx, chain)
	require.Equal(t, len(validators), chainVals)
	assertSignatures(t, ctx, chain, len(validators))
	assertConsensus(t, ctx, chain, len(validators))

	// Stop validator 1 from signing
	if err := chain.Validators[1].StopContainer(ctx); err != nil {
		t.Fatal(err)
	}

	// Wait for the stopped node to be jailed & persist
	t.Log("Waiting for validator to become jailed")
	require.NoError(t, testutil.WaitForBlocks(ctx, 15, chain.Validators[0]))

	// Validate 1 validator is jailed (status 1)
	vals := helpers.GetValidators(t, ctx, chain)
	jailedValAddr := ""
	require.True(t, func() bool {
		for _, v := range vals.Validators {
			if v.Status == int(stakingtypes.Unbonded) || v.Status == int(stakingtypes.Unbonding) {
				fmt.Println("Validator", v.OperatorAddress, "is jailed", v.Status)
				jailedValAddr = v.OperatorAddress
				return true
			}
		}
		return false
	}())

	// Validate the unjail time is in the future
	now := time.Now()
	si := helpers.GetSigningInformation(t, ctx, chain)
	for _, i := range si.Info {
		if i.Address == jailedValAddr {
			t.Log(jailedValAddr, "jailed_until", i.JailedUntil)

			unjailTime, err := time.Parse(time.RFC3339Nano, i.JailedUntil)
			require.NoError(t, err)
			require.True(t, unjailTime.After(now))
		}
	}

	// wait to ensure the chain is not halted
	t.Log("Waiting for chain to progress")
	require.NoError(t, testutil.WaitForBlocks(ctx, 5, chain))
}
