package helpers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func ExecuteAuthzGrantMsg(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, granter ibc.Wallet, grantee ibc.Wallet, msgType string) (sdk.TxResponse, error) {
	if !strings.HasPrefix(msgType, "/") {
		msgType = "/" + msgType
	}

	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "authz", "grant", grantee.FormattedAddress(), "generic", "--msg-type", msgType}, granter.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func ExecuteAuthzExecMsg(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, grantee ibc.Wallet, nestedMsgCmd []string) (sdk.TxResponse, error) {
	// generate the message to JSON, then exec the message
	fileName := "authz.json"
	node := chain.GetNode()
	if err := createAuthzJSON(ctx, node, fileName, nestedMsgCmd); err != nil {
		t.Fatal(err)
	}

	// execute the nested message
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "authz", "exec", node.HomeDir() + "/" + fileName}, grantee.FormattedAddress())
	return ExecuteTransaction(ctx, chain, cmd)
}

func createAuthzJSON(ctx context.Context, node *cosmos.ChainNode, filePath string, genMsgCmd []string) error {
	if !strings.Contains(strings.Join(genMsgCmd, " "), "--generate-only") {
		genMsgCmd = append(genMsgCmd, "--generate-only")
	}

	// Generate msg output
	res, resErr, err := node.Exec(ctx, genMsgCmd, nil)
	if resErr != nil {
		return fmt.Errorf("failed to generate msg: %s", resErr)
	}
	if err != nil {
		return err
	}

	// Write output to file
	return node.WriteFile(ctx, res, filePath)
}
