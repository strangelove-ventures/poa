package helpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func GetBlockData(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, height uint64) BlockData {
	var res BlockData
	ExecuteQuery(ctx, chain, []string{"query", "block", "--type=height", fmt.Sprintf("%d", height)}, &res)
	return res
}
