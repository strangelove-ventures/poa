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

func ValidatorVote(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, voteOp string, searchHeightDelta int64) {
	if err := chain.VoteOnProposalAllValidators(ctx, proposalID, voteOp); err != nil {
		t.Fatal(err)
	}

	height, err := chain.Height(ctx)
	require.NoError(t, err, "failed to get height")

	propID, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposalID")

	resp, err := cosmos.PollForProposalStatusV1(ctx, chain, height, (height-2)+searchHeightDelta, propID, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
	require.NoError(t, err, "failed to poll for proposal status")

	t.Log("PollForProposalStatusV1 resp", resp)
	require.NotNil(t, resp, "ValidatorVote proposal not found:", fmt.Sprintf("proposalID: %s", proposalID))

	require.EqualValues(t, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, resp.Status, "proposal status did not change to passed in expected number of blocks")
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, updatedParams []cosmos.ProtoMessage, sender string, waitForBlocks int64) string {
	expedited := false
	proposal, err := chain.BuildProposal(updatedParams, "UpdateParams", "params", "ipfs://CID", fmt.Sprintf(`50%s`, chain.Config().Denom), sender, expedited)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}
