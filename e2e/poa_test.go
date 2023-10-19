package e2e

import (
	"fmt"
	"testing"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"

	poa "github.com/strangelove-ventures/poa"
)

// TODO: test 50/50 split, set one node too 0. Ensure network halt
// add in safe guards for this ^^

func poaEncoding() *moduletestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	poa.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

const (
	// cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr (test_node.sh)
	acc_mnemonic = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	userFunds    = 10_000_000_000
	numVals      = 2
	numNodes     = 0
)

func TestPOA(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	cfg := ibc.ChainConfig{
		Images: []ibc.DockerImage{
			{
				Repository: "poa",
				Version:    "local",
				UidGid:     "1025:1025",
			},
		},
		ModifyGenesis: cosmos.ModifyGenesis([]cosmos.GenesisKV{
			{
				Key: "app_state.poa.params.admins",
				Value: []string{
					"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // gov
					"cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr", // testing account
				},
			},
		}),
		// TODO: modify gentxs / genesis account amounts?
		EncodingConfig: poaEncoding(),
		Type:           "cosmos",
		Name:           "poa",
		ChainID:        "poa-1",
		Bin:            "poad",
		Bech32Prefix:   "cosmos",
		Denom:          "stake", // maybe poa?
		CoinType:       "118",
		GasPrices:      "0stake,0utest",
		TrustingPeriod: "330h",
	}

	chains := interchaintest.CreateChainWithConfig(t, numVals, numNodes, "poa", "", cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	enableBlockDB := false
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, enableBlockDB)

	acc0, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", acc_mnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	incorrectUser := users[0]

	// TODO: validate all staking commands are disabled.

	validators := []string{}
	for _, v := range helpers.GetValidators(t, ctx, chain).Validators {
		validators = append(validators, v.OperatorAddress)
		t.Log(v.Tokens)
	}
	t.Log(validators)
	require.Equal(t, len(validators), numVals)

	// === Setting Power Test ===
	powerOne := int64(9_000_000_000_000)
	powerTwo := int64(2500)

	helpers.POASetPower(t, ctx, chain, acc0, validators[0], powerOne)
	helpers.POASetPower(t, ctx, chain, acc0, validators[1], powerTwo)
	helpers.POASetPower(t, ctx, chain, incorrectUser, validators[1], 11111111) // err.
	testutil.WaitForBlocks(ctx, 2, chain)

	vals := helpers.GetValidators(t, ctx, chain).Validators
	require.Equal(t, vals[0].Tokens, fmt.Sprintf("%d", powerOne))
	require.Equal(t, vals[1].Tokens, fmt.Sprintf("%d", powerTwo))

	// === Remove Validator Test ===
	helpers.POARemove(t, ctx, chain, acc0, validators[1])
	testutil.WaitForBlocks(ctx, 2, chain)

	vals = helpers.GetValidators(t, ctx, chain).Validators
	require.Equal(t, vals[0].Tokens, fmt.Sprintf("%d", powerOne))
	require.Equal(t, vals[1].Tokens, "0")
	require.Equal(t, vals[1].Status, 1) // 1 = Unbonded status (stakingtypes.Unbonded)
	for _, v := range vals {
		t.Log(v.OperatorAddress, v.Tokens)
	}
}
