package helpers

import (
	"context"
	"fmt"
	"testing"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// Modified from ictest
func VoteOnProposalAllValidators(ctx context.Context, c *cosmos.CosmosChain, proposalID string, vote string) error {
	var eg errgroup.Group
	valKey := "validator"
	for _, n := range c.Nodes() {
		if n.Validator {
			n := n
			eg.Go(func() error {
				// gas-adjustment was using 1.3 default instead of the setup's 2.0+ for some reason.
				// return n.VoteOnProposal(ctx, valKey, proposalID, vote)

				_, err := n.ExecTx(ctx, valKey,
					"gov", "vote",
					proposalID, vote, "--gas", "auto", "--gas-adjustment", "2.0",
				)
				return err
			})
		}
	}
	return eg.Wait()
}

func ValidatorVote(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, voteOp string, searchHeightDelta uint64) {
	err := VoteOnProposalAllValidators(ctx, chain, proposalID, voteOp)
	require.NoError(t, err, "failed to vote on proposal")

	height, err := chain.Height(ctx)
	require.NoError(t, err, "failed to get height")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+searchHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, updatedParams []cosmosproto.Message, sender string, waitForBlocks uint64) string {
	expedited := false
	proposal, err := chain.BuildProposal(updatedParams, "UpdateParams", "params", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), sender, expedited)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	height, _ := chain.Height(ctx)

	err = chain.VoteOnProposalAllValidators(ctx, txProp.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+waitForBlocks, txProp.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	return txProp.ProposalID
}
