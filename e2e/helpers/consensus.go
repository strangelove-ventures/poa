package helpers

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func GetCometBFTConsensus(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) CometBFTConsensus {
	var res CometBFTConsensus
	ExecuteQuery(ctx, chain, []string{"query", "consensus", "comet", "validator-set"}, &res)
	return res
}
