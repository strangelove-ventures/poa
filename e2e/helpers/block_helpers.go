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

// TODO: replace with `GetTransaction` https://github.com/strangelove-ventures/interchaintest/pull/836
func GetTxHash(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, txHash string) TxResponse {
	var res TxResponse
	ExecuteQuery(ctx, chain, []string{"query", "tx", "--type=hash", txHash}, &res)
	return res
}
