package e2e

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

func TestPOABase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

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
	acc1, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc1", acc1Mnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	incorrectUser := users[0]

	// get validator operator addresses
	validators := helpers.GetValidatorsOperatingAddresses(t, ctx, chain)
	require.Equal(t, len(validators), numVals)
	assertSignatures(t, ctx, chain, len(validators))
	assertConsensus(t, ctx, chain, len(validators))

	// === Test Cases ===
	testStakingDisabled(t, ctx, chain, validators, acc0, acc1)
	testPowerErrors(t, ctx, chain, validators, incorrectUser, acc0)
	testRemovePending(t, ctx, chain, acc0)
}

func testStakingDisabled(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0, acc1 ibc.Wallet) {
	t.Log("\n===== TEST STAKING DISABLED =====")
	// Normal delegation execution fails
	txRes, _ := helpers.StakeTokens(t, ctx, chain, acc0, validators[0], "1stake")
	require.Contains(t, txRes.RawLog, poa.ErrStakingActionNotAllowed.Error())

	granter := acc1
	grantee := acc0

	// Grant grantee (acc0) the ability to delegate from granter (acc1)
	res, err := helpers.ExecuteAuthzGrantMsg(t, ctx, chain, granter, grantee, "/cosmos.staking.v1beta1.MsgDelegate")
	require.NoError(t, err)
	require.EqualValues(t, res.Code, 0)

	// Generate nested message
	nested := []string{"tx", "staking", "delegate", validators[0], "1stake"}
	nestedCmd := helpers.TxCommandBuilder(ctx, chain, nested, granter.FormattedAddress())

	// Execute nested message via a wrapped Exec
	res, err = helpers.ExecuteAuthzExecMsg(t, ctx, chain, grantee, nestedCmd)
	require.NoError(t, err)
	require.Contains(t, res.RawLog, poa.ErrStakingActionNotAllowed.Error())
}

func testRemovePending(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet) {
	t.Log("\n===== TEST PENDING =====")

	_, err := helpers.POACreatePendingValidator(t, ctx, chain, acc0, "pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg=", "testval", "0.10", "0.25", "0.05")
	require.NoError(t, err)

	pv := helpers.GetPOAPending(t, ctx, chain).Pending
	require.Equal(t, 1, len(pv))

	_, err = helpers.POARemovePending(t, ctx, chain, acc0, pv[0].OperatorAddress)
	require.NoError(t, err)

	// validate it was removed
	require.Equal(t, 0, len(helpers.GetPOAPending(t, ctx, chain).Pending))
}

func testPowerErrors(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, incorrectUser ibc.Wallet, admin ibc.Wallet) {
	t.Log("\n===== TEST POWER ERRORS =====")
	var res helpers.TxResponse
	var err error

	t.Run("fail: set-power message from a non authorized user", func(t *testing.T) {
		// runtime error: index out of range [1] with length 1 [recovered]
		res, _ = helpers.POASetPower(t, ctx, chain, incorrectUser, validators[0], 1_000_000)
		res, err := chain.GetTransaction(res.Txhash)
		require.NoError(t, err)
		require.Contains(t, res.RawLog, poa.ErrNotAnAuthority.Error())
	})

	t.Run("fail: set-power message below minimum power requirement (self bond)", func(t *testing.T) {
		res, err = helpers.POASetPower(t, ctx, chain, admin, validators[0], 1)
		require.Error(t, err) // cli validate error
		require.Contains(t, err.Error(), poa.ErrPowerBelowMinimum.Error())
	})

	t.Run("fail: set-power message above 30%% without unsafe flag", func(t *testing.T) {
		res, _ = helpers.POASetPower(t, ctx, chain, admin, validators[0], 9_000_000_000_000_000)
		res, err := chain.GetTransaction(res.Txhash)
		require.NoError(t, err)
		require.Contains(t, res.RawLog, poa.ErrUnsafePower.Error())
	})
}
