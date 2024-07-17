package e2e

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	GovModuleAddress = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	RandAcc          = "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"
)

func TestPOAGovernance(t *testing.T) {
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
	assertConsensus(t, ctx, chain, len(validators))

	// === Test Cases ===
	testUpdatePOAParams(t, ctx, chain, acc0, incorrectUser)
	testGovernance(t, ctx, chain, acc0, validators)
}

func testGovernance(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0 ibc.Wallet, validators []string) {
	t.Log("\n===== TEST GOVERNANCE =====")

	t.Run("success: gov proposal update params", func(t *testing.T) {
		updatedParams := []cosmos.ProtoMessage{
			&poa.MsgUpdateParams{
				Sender: GovModuleAddress,
				Params: poa.Params{
					AllowValidatorSelfExit: false,
				},
			},
		}

		propId := helpers.SubmitParamChangeProp(t, ctx, chain, acc0, updatedParams, GovModuleAddress, 25)
		helpers.ValidatorVote(t, ctx, chain, propId, cosmos.ProposalVoteYes, 30)

		require.True(t, helpers.GetPOAParams(t, ctx, chain).AllowValidatorSelfExit, "AllowValidatorSelfExit should be true")
	})

	t.Run("success: gov proposal validator change", func(t *testing.T) {
		// ibc.ChainConfig key: app_state.poa.params.admins must contain the governance address.
		powerAmt := uint64(1_234_567)
		propID := helpers.SubmitGovernanceProposalForValidatorChanges(t, ctx, chain, acc0, validators[0], powerAmt, true, GovModuleAddress)
		helpers.ValidatorVote(t, ctx, chain, propID, cosmos.ProposalVoteYes, 25)

		// validate the validator[0] was set to powerAmt
		val := helpers.GetValidators(t, ctx, chain).Validators[0]
		require.Equal(t, val.Tokens, "1234567")
		p := helpers.GetPOAConsensusPower(t, ctx, chain, val.OperatorAddress)
		require.EqualValues(t, powerAmt/1_000_000, p, "Validator[0] should have 1 consensus power")
	})
}

func testUpdatePOAParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, acc0, incorrectUser ibc.Wallet) {
	var tx sdk.TxResponse
	var err error

	t.Log("\n===== TEST UPDATE POA PARAMS =====")

	t.Run("fail: update-params message from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateParams(t, ctx, chain, incorrectUser, true)
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.TxHash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.Contains(t, txRes.RawLog, poa.ErrNotAnAuthority.Error())
	})

	t.Run("fail: update staking params from a non authorized user", func(t *testing.T) {
		tx, err = helpers.POAUpdateStakingParams(t, ctx, chain, incorrectUser, stakingtypes.DefaultParams())
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.TxHash)
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

		txRes, err := chain.GetTransaction(tx.TxHash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.EqualValues(t, txRes.Code, 0)

		sp := helpers.GetStakingParams(t, ctx, chain)
		fmt.Printf("%+v", sp)
		require.EqualValues(t, sp.MaxValidators, 10)
	})

	t.Run("success: update-params message from an authorized user with cli.", func(t *testing.T) {
		tx, err = helpers.POAUpdateParams(t, ctx, chain, acc0, true)
		if err != nil {
			t.Fatal(err)
		}
		txRes, err := chain.GetTransaction(tx.TxHash)
		require.NoError(t, err)
		fmt.Printf("%+v", txRes)
		require.EqualValues(t, txRes.Code, 0)

		p := helpers.GetPOAParams(t, ctx, chain)
		require.False(t, p.AllowValidatorSelfExit) // TODO: check this
	})

}
