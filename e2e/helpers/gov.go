package helpers

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func ValidatorVote(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, voteOp string, searchHeightDelta uint64) {
	chain.VoteOnProposalAllValidators(ctx, proposalID, voteOp)

	height, err := chain.Height(ctx)
	require.NoError(t, err, "failed to get height")

	propID, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposalID")

	resp, _ := cosmos.PollForProposalStatusV1(ctx, chain, height, height+searchHeightDelta, propID, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
	t.Log("PollForProposalStatusV8 resp", resp)
	require.EqualValues(t, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, resp.Status, "proposal status did not change to passed in expected number of blocks")
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, updatedParams []cosmos.ProtoMessage, sender string, waitForBlocks uint64) string {
	expedited := false
	proposal, err := chain.BuildProposal(updatedParams, "UpdateParams", "params", "ipfs://CID", fmt.Sprintf(`50%s`, chain.Config().Denom), sender, expedited)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	ValidatorVote(t, ctx, chain, txProp.ProposalID, cosmos.ProposalVoteYes, waitForBlocks)

	return txProp.ProposalID
}
