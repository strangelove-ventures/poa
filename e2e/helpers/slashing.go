package helpers

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func GetSigningInformation(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) SingingInformation {
	var res SingingInformation
	ExecuteQuery(ctx, chain, []string{"query", "slashing", "signing-infos"}, &res)
	return res
}
