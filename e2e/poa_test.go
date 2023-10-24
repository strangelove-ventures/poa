package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

// TODO: test 50/50 split, set one node too 0. Ensure network halt
// add in safe guards for this ^^

const (
	// cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr (test_node.sh)
	accMnemonic = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	userFunds   = 10_000_000_000
	numVals     = 2
	numNodes    = 0
)

func TestPOA(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	chains := interchaintest.CreateChainWithConfig(t, numVals, numNodes, "poa", "", POACfg)
	chain := chains[0].(*cosmos.CosmosChain)

	enableBlockDB := false
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, enableBlockDB)

	acc0, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", accMnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	incorrectUser := users[0]

	validators := []string{}
	for _, v := range helpers.GetValidators(t, ctx, chain).Validators {
		validators = append(validators, v.OperatorAddress)
		t.Log(v.Tokens)
	}
	t.Log(validators)
	require.Equal(t, len(validators), numVals)

	assertSignatures(t, ctx, chain, 2)

	// === Staking Commands Disabled (ante) ===
	txRes := helpers.StakeTokens(t, ctx, chain, acc0, validators[0], "1stake")
	require.Contains(t, txRes.RawLog, poa.ErrStakingActionNotAllowed.Error())

	// === Setting Power Test ===
	powerOne := int64(9_000_000_000_000)
	powerTwo := int64(2_500_000)

	helpers.POASetPower(t, ctx, chain, acc0, validators[0], powerOne, true)
	helpers.POASetPower(t, ctx, chain, acc0, validators[1], powerTwo, false)
	helpers.POASetPower(t, ctx, chain, incorrectUser, validators[1], 11111111, false) // err.

	vals := helpers.GetValidators(t, ctx, chain).Validators
	require.Equal(t, vals[0].Tokens, fmt.Sprintf("%d", powerOne))
	require.Equal(t, vals[1].Tokens, fmt.Sprintf("%d", powerTwo))

	// === Remove Validator Test ===
	helpers.POARemove(t, ctx, chain, acc0, validators[1])

	// allow the poa.BeginBlocker to update new status
	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}

	vals = helpers.GetValidators(t, ctx, chain).Validators
	require.Equal(t, vals[0].Tokens, fmt.Sprintf("%d", powerOne))
	require.Equal(t, vals[1].Tokens, "0")
	require.Equal(t, vals[1].Status, 1)
	for _, v := range vals {
		t.Log(v.OperatorAddress, v.Tokens)
	}

	assertSignatures(t, ctx, chain, 1)
}

// assertSignatures asserts that the current block has the exact number of signatures as expected
func assertSignatures(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, expectedSigs int) {
	height, err := chain.GetNode().Height(ctx)
	require.NoError(t, err)
	block := helpers.GetBlockData(t, ctx, chain, height)
	require.Equal(t, len(block.LastCommit.Signatures), expectedSigs)
}
