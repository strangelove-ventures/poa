package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const waitForBlocks = 2

func ExecuteQuery(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, i interface{}, extraFlags ...string) {
	flags := []string{
		"--node", chain.GetRPCAddress(),
		"--output=json",
	}
	flags = append(flags, extraFlags...)

	ExecuteExec(ctx, chain, cmd, i, flags...)
}
func ExecuteExec(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, i interface{}, extraFlags ...string) {
	command := []string{chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, extraFlags...)
	fmt.Println(command)

	stdout, _, err := chain.Exec(ctx, command, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(stdout))
	if err := json.Unmarshal(stdout, &i); err != nil {
		fmt.Println(err)
	}
}

// Executes a command from CommandBuilder
func ExecuteTransaction(ctx context.Context, chain *cosmos.CosmosChain, cmd []string) (TxResponse, error) {
	var err error
	var stdout []byte

	stdout, _, err = chain.Exec(ctx, cmd, nil)
	if err != nil {
		return TxResponse{}, err
	}

	if err := testutil.WaitForBlocks(ctx, waitForBlocks, chain); err != nil {
		return TxResponse{}, err
	}

	var res TxResponse
	if err := json.Unmarshal(stdout, &res); err != nil {
		return TxResponse{}, err
	}

	return res, err
}

func ExecuteTransactionNoError(ctx context.Context, chain *cosmos.CosmosChain, cmd []string) TxResponse {
	res, _ := ExecuteTransaction(ctx, chain, cmd)
	return res
}

func TxCommandBuilder(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, fromUser ibc.Wallet, extraFlags ...string) []string {
	return TxCommandBuilderNode(ctx, chain.GetNode(), cmd, fromUser, extraFlags...)
}

func TxCommandBuilderNode(ctx context.Context, node *cosmos.ChainNode, cmd []string, fromUser ibc.Wallet, extraFlags ...string) []string {
	command := []string{node.Chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, "--node", node.Chain.GetRPCAddress())
	command = append(command, "--home", node.HomeDir())
	command = append(command, "--chain-id", node.Chain.Config().ChainID)
	command = append(command, "--from", fromUser.KeyName())
	command = append(command, "--keyring-backend", keyring.BackendTest)
	command = append(command, "--output=json")
	command = append(command, "--yes")

	gasFlag := false
	for _, flag := range extraFlags {
		if flag == "--gas" {
			gasFlag = true
		}
	}

	if !gasFlag {
		command = append(command, "--gas", "500000")
	}

	command = append(command, extraFlags...)
	return command
}
