package helpers

import (
	"context"
	"fmt"
	"testing"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func ValidatorVote(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, voteOp string, searchHeightDelta uint64) {
	chain.VoteOnProposalAllValidators(ctx, proposalID, voteOp)

	height, err := chain.Height(ctx)
	require.NoError(t, err, "failed to get height")

	resp, _ := cosmos.PollForProposalStatusV8(ctx, chain, height, height+searchHeightDelta, proposalID, cosmos.ProposalStatusPassedV8)
	t.Log("PollForProposalStatusV8 resp", resp)
	require.EqualValues(t, resp.Proposal.Status, cosmos.ProposalStatusPassedV8, "proposal status did not change to passed in expected number of blocks")
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, updatedParams []cosmosproto.Message, sender string, waitForBlocks uint64) string {
	expedited := false
	proposal, err := chain.BuildProposal(updatedParams, "UpdateParams", "params", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), sender, expedited)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	ValidatorVote(t, ctx, chain, txProp.ProposalID, cosmos.ProposalVoteYes, waitForBlocks)

	return txProp.ProposalID
}
