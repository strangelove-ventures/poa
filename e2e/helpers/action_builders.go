package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
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

	stdout, stderr, err := chain.Exec(ctx, command, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("ExecuteExec", "stdout", string(stdout), "stderr", string(stderr))

	err2 := json.Unmarshal(stdout, &i)
	if err2 != nil {
		fmt.Println("json.Unmarshal", err2)
		return // guard return as to not show the next error
	}

	// If the codec can not properly unmarshal, then it may be a standard json Unmarshal request.
	// This is required since we are ExecuteExec'ing an interface{} instead of some concrete type.
	cdc := chain.GetCodec()
	if cdc != nil {
		err = cdc.UnmarshalInterface(stdout, &i)
		if err != nil && !strings.Contains(err.Error(), "illegal wireType") {
			fmt.Println("chain.GetCodec.UnmarshalInterface", err)
		}
	}
}

// Executes a command from CommandBuilder
func ExecuteTransaction(ctx context.Context, chain *cosmos.CosmosChain, cmd []string) (sdk.TxResponse, error) {
	var err error
	var stdout []byte

	stdout, _, err = chain.Exec(ctx, cmd, nil)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	if err := testutil.WaitForBlocks(ctx, waitForBlocks, chain); err != nil {
		return sdk.TxResponse{}, err
	}

	var res sdk.TxResponse

	cdc := chain.GetCodec()

	if cdc != nil {
		if err := cdc.UnmarshalJSON(stdout, &res); err != nil {
			// return sdk.TxResponse{}, fmt.Errorf("failed to cdc unmarshal tx response: %s", err)
			fmt.Println("cdc.UnmarshalJSON err", err, "using json Unmarshal instead")
			json.Unmarshal(stdout, &res)
		}
	} else {
		json.Unmarshal(stdout, &res)
	}

	return res, err
}

func ExecuteTransactionNoError(ctx context.Context, chain *cosmos.CosmosChain, cmd []string) sdk.TxResponse {
	res, _ := ExecuteTransaction(ctx, chain, cmd)
	return res
}

func TxCommandBuilder(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, fromUser string, extraFlags ...string) []string {
	return TxCommandBuilderNode(ctx, chain.GetNode(), cmd, fromUser, extraFlags...)
}

func TxCommandBuilderNode(ctx context.Context, node *cosmos.ChainNode, cmd []string, fromUser string, extraFlags ...string) []string {
	command := []string{node.Chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, "--node", node.Chain.GetRPCAddress())
	command = append(command, "--home", node.HomeDir())
	command = append(command, "--chain-id", node.Chain.Config().ChainID)
	command = append(command, "--from", fromUser)
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
