package e2e

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa/e2e/helpers"
)

func TestPOARemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		delegation        int64 = 1_000_000000
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

	// consensus needs to be len of initialValidators - 1
	// TODO: validator is not being removed. Instead is stuck in unbonding (this is technically fine, but I don't like).
	// - So signatures are still committed to the block but not really contributing
	assertSignatures(t, ctx, chain, initialValidators-1)
	require.Equal(t, initialValidators-1, len(consensus.Validators), "BFT consensus should have one less validator")
	require.Equal(t, initialValidators-1, len(vals), "Validators should have one less validator")

	// TODO: validate it is in unbonded status with 0 tokens
}

func getValToRemove(t *testing.T, vals helpers.Validators, delegationAmt int64) string {
	valToRemove := ""
	for _, v := range vals {
		if v.Tokens == fmt.Sprintf("%d", delegationAmt) {
			valToRemove = v.OperatorAddress
			break
		}
	}
	require.NotEmpty(t, valToRemove)

	return valToRemove
}
