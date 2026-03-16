package e2e

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

func TestPOARemoveAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		delegation        int64 = 1_000_000_000
		initialValidators       = 3
	)

	rmCfg := POACfg.Clone()
	rmCfg.ModifyGenesisAmounts = func(i int) (sdk.Coin, sdk.Coin) {
		var denom string = rmCfg.Denom
		delCoin := sdk.NewCoin(denom, sdkmath.NewInt(delegation))

		if i == 0 {
			delCoin = sdk.NewCoin(denom, sdkmath.NewInt(delegation*5))
		}
		return sdk.NewCoin(denom, sdkmath.NewInt(userFunds.Int64())), delCoin
	}

	// setup base chain
	chains := interchaintest.CreateChainWithConfig(t, initialValidators, numNodes, "poa", "", rmCfg)
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, false)

	chain := chains[0].(*cosmos.CosmosChain)

	acc0, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", accMnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all validators are signing as expected
	vals := helpers.GetValidators(t, ctx, chain).Validators
	consensus := helpers.GetCometBFTConsensus(t, ctx, chain)
	assertSignatures(t, ctx, chain, initialValidators)
	require.Equal(t, initialValidators, len(consensus.Validators), "BFT consensus should have the same number of validators", initialValidators, consensus.Validators)
	require.Equal(t, initialValidators, len(vals), "Validators should have the same number of validators", initialValidators, len(vals))

	////////////////////////////
	// REMOVING A VALIDATOR
	////////////////////////////

	// Gets the first validator that has said delegation amount
	valToRemove := getValToRemove(t, vals, delegation)

	// Remove a validator from consensus (keep it singing)
	txRes, err := helpers.POARemove(t, ctx, chain, acc0, valToRemove)
	require.NoError(t, err)
	require.EqualValues(t, 0, txRes.Code, "txRes.Code should be 0")
	fmt.Println("txRes", txRes)

	testutil.WaitForBlocks(ctx, 5, chain)

	// query validators now (we should have less, also check consensus)
	vals = helpers.GetValidators(t, ctx, chain).Validators
	fmt.Println("validators", len(vals), vals)

	consensus = helpers.GetCometBFTConsensus(t, ctx, chain)
	fmt.Printf("consensus: %+v", consensus)

	require.EqualValues(t, 0, helpers.GetPOAConsensusPower(t, ctx, chain, valToRemove))

	assertSignatures(t, ctx, chain, initialValidators-1)
	require.Equal(t, initialValidators-1, len(consensus.Validators), "BFT consensus should have one less validator")
	require.Equal(t, initialValidators-1, getBondedValidators(t, ctx, chain), "Bonded validators should have one less active validator")

	/////////////////////////////
	// RESTART CHAIN
	////////////////////////////

	err = chain.StopAllNodes(ctx)
	require.NoError(t, err)

	t.Log("Waiting for chain to stop...")

	time.Sleep(30 * time.Second)

	err = chain.StartAllNodes(ctx)
	require.NoError(t, err)

	////////////////////////////
	// INDUCTING A NEW VALIDATOR
	////////////////////////////

	pubKeyJSON := `pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg=`
	moniker := "mytestval"
	commissionRate := "0.1"
	commissionMaxRate := "0.2"
	commissionMaxChangeRate := "0.01"

	require.EqualValues(t, 0, len(helpers.GetPOAPending(t, ctx, chain).Pending))

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	fakeVal := users[0]

	txRes, err = helpers.POACreatePendingValidator(t, ctx, chain, fakeVal, pubKeyJSON, moniker, commissionRate, commissionMaxRate, commissionMaxChangeRate)
	require.NoError(t, err)
	log.Debug().Msgf("Create pending validator: %v", txRes)

	pending := helpers.GetPOAPending(t, ctx, chain).Pending
	require.EqualValues(t, 1, len(pending))
	require.Equal(t, "0", pending[0].Tokens)
	require.Equal(t, "1", pending[0].MinSelfDelegation)

	// set power from Gov
	txRes, err = helpers.POASetPower(t, ctx, chain, acc0, pending[0].OperatorAddress, 1_000_000)
	require.NoError(t, err)
	log.Debug().Msgf("Set pending's power: %v", txRes)

	require.NoError(t, err)
	require.NoError(t, testutil.WaitForBlocks(ctx, 4, chain))
}
