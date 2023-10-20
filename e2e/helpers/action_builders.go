package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func QueryBuilder(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, i interface{}, extraFlags ...string) {
	command := []string{chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, "--node", chain.GetRPCAddress())
	command = append(command, "--output=json")
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
