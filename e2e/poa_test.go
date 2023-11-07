package e2e

import (
	"context"
	"fmt"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	// This should be the last test since param changes happen
	testUpdatePOAParams(t, ctx, chain, validators, acc0, incorrectUser)
}

func testUpdatePOAParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0, incorrectUser ibc.Wallet) {
	var tx helpers.TxResponse
	var err error
	var txRes helpers.TxResponse

	t.Run("fail: update-params message from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateParams(t, ctx, chain, incorrectUser, []string{incorrectUser.FormattedAddress()})
		if err != nil {
			t.Fatal(err)
		}
		txRes = helpers.GetTxHash(t, ctx, chain, tx.Txhash)
		fmt.Printf("%+v", txRes)
		require.Contains(t, txRes.RawLog, poa.ErrNotAnAuthority.Error())
	})

	t.Run("fail: update staking params from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateStakingParams(t, ctx, chain, incorrectUser, stakingtypes.DefaultParams())
		if err != nil {
			t.Fatal(err)
		}
		txRes = helpers.GetTxHash(t, ctx, chain, tx.Txhash)
		fmt.Printf("%+v", txRes)
		require.Equal(t, txRes.Code, 3)

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

		txRes = helpers.GetTxHash(t, ctx, chain, tx.Txhash)
		fmt.Printf("%+v", txRes)
		require.Equal(t, txRes.Code, 0)

		sp := helpers.GetStakingParams(t, ctx, chain)
		fmt.Printf("%+v", sp)
		require.EqualValues(t, sp.MaxValidators, 10)
	})

	t.Run("success: update-params message from an authorized user.", func(t *testing.T) {
		govModule := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
		randAcc := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

		newAdmins := []string{acc0.FormattedAddress(), govModule, randAcc, incorrectUser.FormattedAddress()}
		tx, err = helpers.POAUpdateParams(t, ctx, chain, acc0, newAdmins)
		if err != nil {
			t.Fatal(err)
		}
		txRes = helpers.GetTxHash(t, ctx, chain, tx.Txhash)
		fmt.Printf("%+v", txRes)
		require.Equal(t, txRes.Code, 0)

		p := helpers.GetPOAParams(t, ctx, chain)
		for _, admin := range newAdmins {
			require.Contains(t, p.Admins, admin)
		}
	})

	// TODO: SDK v50 does not seem to unmarshal this with the CLI.
	// cannot marshal response proposals:{id:1 messages:{[/strangelove_ventures.poa.v1.MsgUpdateParams]:{sender:"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" params:{admins:"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" admins:"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"}}} status:PROPOSAL_STATUS_VOTING_PERIOD final_tally_result:{yes_count:"0" abstain_count:"0" no_count:"0" no_with_veto_count:"0"} submit_time:{seconds:1698682973 nanos:181594516} deposit_end_time:{seconds:1698855773 nanos:181594516} total_deposit:{denom:"stake" amount:"1000000"} voting_start_time:{seconds:1698682973 nanos:181594516} voting_end_time:{seconds:1698855773 nanos:181594516} metadata:"ipfs://CID" title:"title" summary:"summary" proposer:"cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr"} pagination:{total:1}: unexpected end of JSON input
	// t.Run("success: gov proposal update", func(t *testing.T) {
	// 	govModule := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	// 	randAcc := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

	// 	updatedParams := []cosmosproto.Message{
	// 		&poa.MsgUpdateParams{
	// 			Sender: govModule,
	// 			Params: poa.Params{
	// 				Admins: []string{acc0.FormattedAddress(), govModule, randAcc},
	// 			},
	// 		},
	// 	}
	// 	propId := helpers.SubmitParamChangeProp(t, ctx, chain, incorrectUser, updatedParams, govModule, 25)
	// 	helpers.ValidatorVote(t, ctx, chain, propId, cosmos.ProposalVoteYes, uint64(25))
	// 	p := helpers.GetPOAParams(t, ctx, chain)
	// 	require.NotContains(t, p.Admins, incorrectUser.FormattedAddress())
	// })

}

func testRemoveValidator(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0 ibc.Wallet) {
	t.Log("\n===== TEST REMOVE VALIDATOR =====")
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
	t.Log("\n===== TEST POWER SWING =====")
	helpers.POASetPower(t, ctx, chain, acc0, validators[0], 1_000_000_000_000_000, "--unsafe")
	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
	helpers.POASetPower(t, ctx, chain, acc0, validators[0], 1_000_000, "--unsafe")
}

func testStakingDisabled(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, acc0 ibc.Wallet) {
	t.Log("\n===== TEST STAKING DISABLED =====")
	txRes, _ := helpers.StakeTokens(t, ctx, chain, acc0, validators[0], "1stake")
	require.Contains(t, txRes.RawLog, poa.ErrStakingActionNotAllowed.Error())
}

func testGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet, validators []string) {
	t.Log("\n===== TEST GOVERNANCE =====")
	// ibc.ChainConfig key: app_state.poa.params.admins must contain the governance address.
	propID := helpers.SubmitGovernanceProposalForValidatorChanges(t, ctx, chain, acc0, validators[0], 1_234_567, true)
	helpers.ValidatorVote(t, ctx, chain, propID, cosmos.ProposalVoteYes, 25)

	// validate the validator[0] was set to 1_234_567
	val := helpers.GetValidators(t, ctx, chain).Validators[0]
	require.Equal(t, val.Tokens, "1234567")
}

func testPowerErrors(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, validators []string, incorrectUser ibc.Wallet, admin ibc.Wallet) {
	t.Log("\n===== TEST POWER ERRORS =====")
	var res helpers.TxResponse
	var err error

	t.Run("fail: set-power message from a non authorized user", func(t *testing.T) {
		res, _ = helpers.POASetPower(t, ctx, chain, incorrectUser, validators[1], 1_000_000)
		txRes := helpers.GetTxHash(t, ctx, chain, res.Txhash)
		require.Contains(t, txRes.RawLog, poa.ErrNotAnAuthority.Error())
	})

	t.Run("fail: set-power message below minimum power requirement (self bond)", func(t *testing.T) {
		res, err = helpers.POASetPower(t, ctx, chain, admin, validators[0], 1)
		require.Error(t, err) // cli validate error
		require.Contains(t, err.Error(), poa.ErrPowerBelowMinimum.Error())
	})

	t.Run("fail: set-power message above 30%% without unsafe flag", func(t *testing.T) {
		res, _ = helpers.POASetPower(t, ctx, chain, admin, validators[0], 9_000_000_000_000_000)
		res = helpers.GetTxHash(t, ctx, chain, res.Txhash)
		require.Contains(t, res.RawLog, poa.ErrUnsafePower.Error())
	})
}

// assertSignatures asserts that the current block has the exact number of signatures as expected
func assertSignatures(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, expectedSigs int) {
	height, err := chain.GetNode().Height(ctx)
	require.NoError(t, err)
	block := helpers.GetBlockData(t, ctx, chain, height)
	require.Equal(t, len(block.LastCommit.Signatures), expectedSigs)
}
