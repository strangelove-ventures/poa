package e2e

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingttypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

const (
	chainName   = "poa"
	upgradeName = "v2-remove-poa"

	haltHeightDelta    = int64(9) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = int64(3)
)

func TestPoAToPoSUpgrade(t *testing.T) {
	// repo, version := GetDockerImageInfo()
	CosmosChainUpgradeTest(t, chainName, POAImage.Version, POAImage.Repository)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, upgradeBranchVersion, upgradeRepo string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(chainName, upgradeBranchVersion, upgradeRepo, upgradeName)

	numVals, numNodes := 4, 0
	cfg := POACfg

	chains := interchaintest.CreateChainWithConfig(t, numVals, numNodes, "poa", "", cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ctx, ic, client, _ := interchaintest.BuildInitialChain(t, chains, false)
	t.Cleanup(func() {
		_ = ic.Close()
	})

	userFunds := sdkmath.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// upgrade
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta

	propIdStr := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName, haltHeight)

	propId, err := strconv.ParseUint(propIdStr, 10, 64)
	require.NoError(t, err, "failed to convert proposal ID to uint64")

	err = chain.VoteOnProposalAllValidators(ctx, propIdStr, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, propId, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	t.Logf("height before upgrade: %d", height)

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*10)
	defer timeoutCtxCancel()

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, chain)

	// // bring down nodes to prepare for upgrade
	err = chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	chain.UpgradeVersion(ctx, client, "poa-removed", "local") // TODO: push to GHCR?

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*30)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	vals, err := chain.StakingQueryValidators(ctx, stakingttypes.BondStatusBonded)
	require.NoError(t, err, "error querying validators")

	validator := vals[0]

	val, err := chain.StakingQueryValidator(ctx, validator.OperatorAddress)
	require.NoError(t, err, "error querying validator")
	t.Logf("validator: %s", val)

	before := val.Tokens

	// TODO: confirm staking works here by staking to a validator
	err = chain.GetNode().StakingDelegate(ctx, "validator", validator.OperatorAddress, "7000000"+chain.Config().Denom)
	require.NoError(t, err, "error delegating to validator")

	val, err = chain.StakingQueryValidator(ctx, validator.OperatorAddress)
	require.NoError(t, err, "error querying validator")
	t.Logf("validator: %s", val)

	after := val.Tokens

	require.True(t, after.GT(before), "after tokens is not greater than before")
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight int64, upgradeRepo, upgradeBranchVersion string) {
	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	t.Log("starting node(s)")
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")
}

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight int64) string {
	upgradeMsg := []cosmos.ProtoMessage{
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
				Info:   "",
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg, "Chain Upgrade 1", "Summary desc", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), user.FormattedAddress(), false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}
