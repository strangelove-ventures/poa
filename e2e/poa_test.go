package e2e

import (
	"context"
	"fmt"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	testPending(t, ctx, chain, acc0)
	testGovernance(t, ctx, chain, acc0, validators)
	testUpdatePOAParams(t, ctx, chain, acc0, incorrectUser)
}

func testUpdatePOAParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0, incorrectUser ibc.Wallet) {
	var tx helpers.TxResponse
	var err error

	t.Run("fail: update-params message from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateParams(t, ctx, chain, incorrectUser, []string{incorrectUser.FormattedAddress()}, true)
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.Txhash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.Contains(t, txRes.RawLog, poa.ErrNotAnAuthority.Error())
	})

	t.Run("fail: update staking params from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateStakingParams(t, ctx, chain, incorrectUser, stakingtypes.DefaultParams())
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.Txhash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.EqualValues(t, txRes.Code, 3)

		sp := helpers.GetStakingParams(t, ctx, chain)
		fmt.Printf("%+v", sp)
	})

	t.Run("success: update staking params from an authorized user.", func(t *testing.T) {
		stakingparams := stakingtypes.DefaultParams()
		tx, err = helpers.POAUpdateStakingParams(t, ctx, chain, acc0, stakingtypes.Params{
			UnbondingTime:     stakingparams.UnbondingTime,
			MaxValidators:     10,
			MaxEntries:        stakingparams.MaxEntries,
			HistoricalEntries: stakingparams.HistoricalEntries,
			BondDenom:         stakingparams.BondDenom,
			MinCommissionRate: stakingparams.MinCommissionRate,
		})
		if err != nil {
			t.Fatal(err)
		}

		txRes, err := chain.GetTransaction(tx.Txhash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.EqualValues(t, txRes.Code, 0)

		sp := helpers.GetStakingParams(t, ctx, chain)
		fmt.Printf("%+v", sp)
		require.EqualValues(t, sp.MaxValidators, 10)
	})

	t.Run("success: update-params message from an authorized user.", func(t *testing.T) {
		govModule := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
		randAcc := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

		newAdmins := []string{acc0.FormattedAddress(), govModule, randAcc, incorrectUser.FormattedAddress()}
		tx, err = helpers.POAUpdateParams(t, ctx, chain, acc0, newAdmins, true)
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.Txhash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.EqualValues(t, txRes.Code, 0)

		p := helpers.GetPOAParams(t, ctx, chain)
		for _, admin := range newAdmins {
			require.Contains(t, p.Admins, admin)
		}
	})

	t.Run("success: gov proposal update", func(t *testing.T) {
		govModule := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
		randAcc := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

		updatedParams := []cosmos.ProtoMessage{
			&poa.MsgUpdateParams{
				Sender: govModule,
				Params: poa.Params{
					Admins: []string{acc0.FormattedAddress(), govModule, randAcc},
				},
			},
		}
		propId := helpers.SubmitParamChangeProp(t, ctx, chain, incorrectUser, updatedParams, govModule, 25)
		helpers.ValidatorVote(t, ctx, chain, propId, cosmos.ProposalVoteYes, 25)
		for _, admin := range helpers.GetPOAParams(t, ctx, chain).Admins {
			require.NotEqual(t, admin, incorrectUser.FormattedAddress())
		}
	})

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

func testPending(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet) {
	t.Log("\n===== TEST PENDING =====")

	_, err := helpers.POACreatePendingValidator(t, ctx, chain, acc0, "pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg=", "testval", "0.10", "0.25", "0.05")
	require.NoError(t, err)

	pv := helpers.GetPOAPending(t, ctx, chain)
	require.Equal(t, 1, len(pv.Pending))
	require.Equal(t, "0", pv.Pending[0].Tokens)
	require.Equal(t, "1", pv.Pending[0].MinSelfDelegation)

	_, err = helpers.POARemovePending(t, ctx, chain, acc0, pv.Pending[0].OperatorAddress)
	require.NoError(t, err)

	// validate it was removed
	pv = helpers.GetPOAPending(t, ctx, chain)
	require.Equal(t, 0, len(pv.Pending))
}

func testGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet, validators []string) {
	t.Log("\n===== TEST GOVERNANCE =====")

	authorityAddr := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // gov

	// ibc.ChainConfig key: app_state.poa.params.admins must contain the governance address.
	propID := helpers.SubmitGovernanceProposalForValidatorChanges(t, ctx, chain, acc0, validators[0], 1_234_567, true, authorityAddr)
	helpers.ValidatorVote(t, ctx, chain, propID, cosmos.ProposalVoteYes, 25)

	// validate the validator[0] was set to 1_234_567
	val := helpers.GetValidators(t, ctx, chain).Validators[0]
	require.Equal(t, val.Tokens, "1234567")
	p := helpers.GetPOAConsensusPower(t, ctx, chain, val.OperatorAddress)
	require.EqualValues(t, 1_234_567/1_000_000, p)
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
