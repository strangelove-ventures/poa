package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
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

	// setup base chain
	chains := interchaintest.CreateChainWithConfig(t, 1, 0, "poa", "", POACfg)
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

	// === Test Cases ===
	testAddValidator(t, ctx, chain, admin, fakeVal)
	time.Sleep(1_000 * time.Minute)

}

func testAddValidator(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin, fakeVal ibc.Wallet) {
	// create a new validator and ensure it is added to the consensus set, but assert signatures should still not be

	// pubKeyJSON := `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg="}`
	pubKeyJSON := `pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg=`
	moniker := "mytestval"
	commissionRate := "0.1"
	commissionMaxRate := "0.2"
	commissionMaxChangeRate := "0.01"

	txRes, err := helpers.POACreatePendingValidator(t, ctx, chain, fakeVal, pubKeyJSON, moniker, commissionRate, commissionMaxRate, commissionMaxChangeRate)
	require.NoError(t, err)
	fmt.Println("txRes: ", txRes)

	pending := helpers.GetPOAPending(t, ctx, chain)

	// set power from Gov
	txRes, err = helpers.POASetPower(t, ctx, chain, admin, pending.Pending[0].OperatorAddress, 1_000_000)
	require.NoError(t, err)
	fmt.Println("set power: ", txRes)

	require.NoError(t, testutil.WaitForBlocks(ctx, 5, chain))

	vals, err := chain.StakingQueryValidators(ctx, stakingtypes.BondStatusBonded)
	require.NoError(t, err)
	fmt.Println("Validators: ", vals)

	require.Equal(t, 2, len(vals))

	assertSignatures(t, ctx, chain, 1) // They are not signing, should be 0
	assertConsensus(t, ctx, chain, 2)  // ensure they were added to CometBFT
}
