package e2e

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/strangelove-ventures/poa/e2e/helpers"
	"github.com/stretchr/testify/require"
)

const (
	// cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr (test_node.sh)
	accMnemonic  = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	acc1Mnemonic = "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"
)

var (
	userFunds = sdkmath.NewInt(10_000_000_000)
)

func TestPOAAddValidator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := POACfg.Clone()
	cfg.Env = []string{
		fmt.Sprintf("POA_ADMIN_ADDRESS=%s", "cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr"), // acc0 / admin
	}
	cfg.ModifyGenesisAmounts = func(i int) (sdk.Coin, sdk.Coin) {
		var delegation int64 = 1_000_000000
		return sdk.NewCoin(cfg.Denom, sdkmath.NewInt(userFunds.Int64())), sdk.NewCoin(cfg.Denom, sdkmath.NewInt(delegation))
	}

	// setup base chain
	chains := interchaintest.CreateChainWithConfig(t, 1, 0, "poa", "", cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	enableBlockDB := false
	ctx, _, _, _ := interchaintest.BuildInitialChain(t, chains, enableBlockDB)

	// setup accounts
	admin, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "acc0", accMnemonic, userFunds, chain)
	if err != nil {
		t.Fatal(err)
	}

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	fakeVal := users[0]

	// get validator operator addresses
	validators := helpers.GetValidatorsOperatingAddresses(t, ctx, chain)
	require.Equal(t, len(validators), 1)
	assertSignatures(t, ctx, chain, 1)
	assertConsensus(t, ctx, chain, 1)

	// create a new validator and ensure it is added to the consensus set, but assert signatures should still not be

	// pubKeyJSON := `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg="}`
	pubKeyJSON := `pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg=`
	moniker := "mytestval"
	commissionRate := "0.1"
	commissionMaxRate := "0.2"
	commissionMaxChangeRate := "0.01"

	require.EqualValues(t, 0, len(helpers.GetPOAPending(t, ctx, chain).Pending))

	txRes, err := helpers.POACreatePendingValidator(t, ctx, chain, fakeVal, pubKeyJSON, moniker, commissionRate, commissionMaxRate, commissionMaxChangeRate)
	require.NoError(t, err)
	log.Debug().Msgf("Create pending validator: %v", txRes)

	pending := helpers.GetPOAPending(t, ctx, chain).Pending
	require.EqualValues(t, 1, len(pending))
	require.Equal(t, "0", pending[0].Tokens)
	require.Equal(t, "1", pending[0].MinSelfDelegation)

	// set power from Gov
	txRes, err = helpers.POASetPower(t, ctx, chain, admin, pending[0].OperatorAddress, 1_000_000)
	require.NoError(t, err)
	log.Debug().Msgf("Set pending's power: %v", txRes)

	vals, err := chain.StakingQueryValidators(ctx, stakingtypes.BondStatusBonded)
	require.NoError(t, err)
	log.Debug().Msgf("Validators: %v", vals)

	require.Equal(t, 2, len(vals))

	assertSignatures(t, ctx, chain, 1) // They are not signing, should be 0
	assertConsensus(t, ctx, chain, 2)  // ensure they were added to CometBFT

	// validate that the ABCI events are not "stuck" (where the event is not cleared by POA or x/staking)
	_, err = helpers.POASetPower(t, ctx, chain, admin, pending[0].OperatorAddress, 2_000_000)
	require.NoError(t, err)

	require.NoError(t, testutil.WaitForBlocks(ctx, 4, chain))

}
