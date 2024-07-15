package helpers

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func StakeTokens(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper, coinAmt string) (sdk.TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "staking", "delegate", valoper, coinAmt}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func ClaimStakingRewards(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (sdk.TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "distribution", "withdraw-rewards", valoper}, user.KeyName())
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

func GetStakingParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) stakingtypes.Params {
	var res StakingParams
	ExecuteQuery(ctx, chain, []string{"query", "staking", "params"}, &res)
	return res.Params
}
