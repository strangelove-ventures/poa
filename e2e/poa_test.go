package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

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

	// setup base chain
	chains := interchaintest.CreateChainWithConfig(t, numVals, numNodes, "poa", "", POACfg)
	chain := chains[0].(*cosmos.CosmosChain)

	enableBlockDB := false
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, enableBlockDB)

	// setup accounts
	acc0, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", accMnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	incorrectUser := users[0]

	// get validator operator addresses
	validators := helpers.GetValidatorsOperatingAddresses(t, ctx, chain)
	require.Equal(t, len(validators), numVals)
	assertSignatures(t, ctx, chain, len(validators))

	// === Test Cases ===
	testStakingDisabled(t, ctx, chain, validators, acc0)
	testGovernance(t, ctx, chain, acc0, validators)
	testPowerErrors(t, ctx, chain, validators, incorrectUser, acc0)
	testRemoveValidator(t, ctx, chain, validators, acc0)
	testPowerSwing(t, ctx, chain, validators, acc0)

	// add a new node, create validator, add add them now at 50% with the validators[0]. This new validator is validator[2]
	// - create new node
	// - sync it
	// - create validator
	// - add validator
	// - set power to 1_000_000 (50%
	// - wait for blocks
	// - verify there are 2 signatures in blocks again.

	// Shut down 1 of those validators, ensure the network height halts.

}

func testRemoveValidator(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0 ibc.Wallet) {
	t.Log("===== TEST REMOVE VALIDATOR =====")
	powerOne := int64(9_000_000_000_000)
	powerTwo := int64(2_500_000)

	helpers.POASetPower(t, ctx, chain, acc0, validators[0], powerOne, "--unsafe")
	helpers.POASetPower(t, ctx, chain, acc0, validators[1], powerTwo)

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
	require.Equal(t, vals[1].Status, 1) // 1 = unbonded

	// only 1 validator is signing now
	assertSignatures(t, ctx, chain, 1)
}

func testPowerSwing(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0 ibc.Wallet) {
	t.Log("===== TEST POWER SWING =====")
	helpers.POASetPower(t, ctx, chain, acc0, validators[0], 1_000_000_000_000_000, "--unsafe")
	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
	helpers.POASetPower(t, ctx, chain, acc0, validators[0], 1_000_000, "--unsafe")
}

func testStakingDisabled(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0 ibc.Wallet) {
	t.Log("===== TEST STAKING DISABLED =====")
	txRes, _ := helpers.StakeTokens(t, ctx, chain, acc0, validators[0], "1stake")
	require.Contains(t, txRes.RawLog, poa.ErrStakingActionNotAllowed.Error())
}

func testGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet, validators []string) {
	t.Log("===== TEST GOVERNANCE =====")
	// ibc.ChainConfig key: app_state.poa.params.admins must contain the governance address.
	propID := helpers.SubmitGovernanceProposalForValidatorChanges(t, ctx, chain, acc0, validators[0], 1_234_567, true)
	helpers.ValidatorVote(t, ctx, chain, propID, 25)

	// validate the validator[0] was set to 1_234_567
	val := helpers.GetValidators(t, ctx, chain).Validators[0]
	require.Equal(t, val.Tokens, "1234567")
}

func testPowerErrors(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, incorrectUser ibc.Wallet, admin ibc.Wallet) {
	t.Log("===== TEST POWER ERRORS =====")
	var res helpers.TxResponse
	var err error

	// unauthorized user
	res, _ = helpers.POASetPower(t, ctx, chain, incorrectUser, validators[1], 1_000_000)
	txRes := helpers.GetTxHash(t, ctx, chain, res.Txhash)
	require.Contains(t, txRes.RawLog, poa.ErrNotAnAuthority.Error())

	// below minimum power requirement (self bond)
	res, err = helpers.POASetPower(t, ctx, chain, admin, validators[0], 1)
	require.Error(t, err) // cli validate error
	require.Contains(t, err.Error(), poa.ErrPowerBelowMinimum.Error())

	// above 30% without unsafe
	res, _ = helpers.POASetPower(t, ctx, chain, admin, validators[0], 9_000_000_000_000_000)
	res = helpers.GetTxHash(t, ctx, chain, res.Txhash)
	require.Contains(t, res.RawLog, poa.ErrUnsafePower.Error())
}

// assertSignatures asserts that the current block has the exact number of signatures as expected
func assertSignatures(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, expectedSigs int) {
	height, err := chain.GetNode().Height(ctx)
	require.NoError(t, err)
	block := helpers.GetBlockData(t, ctx, chain, height)
	require.Equal(t, len(block.LastCommit.Signatures), expectedSigs)
}
