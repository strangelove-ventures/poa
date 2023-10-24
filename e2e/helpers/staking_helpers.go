package helpers

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func StakeTokens(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper, coinAmt string) (TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "staking", "delegate", valoper, coinAmt}, user)
	return ExecuteTransaction(ctx, chain, cmd)
}

func ClaimStakingRewards(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "distribution", "withdraw-rewards", valoper}, user)
	return ExecuteTransaction(ctx, chain, cmd)
}

func GetValidators(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) Vals {
	var res Vals
	ExecuteQuery(ctx, chain, []string{"query", "staking", "validators"}, &res)
	return res
}

func GetValidatorsOperatingAddresses(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) []string {
	validators := []string{}
	for _, v := range GetValidators(t, ctx, chain).Validators {
		validators = append(validators, v.OperatorAddress)
	}

	return validators
}
