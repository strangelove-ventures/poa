package helpers

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func POASetPower(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string, power int64, flags ...string) (sdk.TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "set-power", valoper, fmt.Sprintf("%d", power)}, user.KeyName(), flags...)
	return ExecuteTransaction(ctx, chain, cmd)
}

func POARemove(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (sdk.TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "remove", valoper}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func POARemovePending(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (sdk.TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "remove-pending", valoper}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func POACreatePendingValidator(
	t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet,
	ed25519PubKey string, moniker string, commissionRate string, commissionMaxRate string, commissionMaxChangeRate string,
) (sdk.TxResponse, error) {
	file := "validator_file.json"

	content := fmt.Sprintf(`{
		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"%s"},
		"amount": "0%s",
		"moniker": "%s",
		"identity": "",
		"website": "https://website.com",
		"security": "security@cosmos.xyz",
		"details": "description",
		"commission-rate": "%s",
		"commission-max-rate": "%s",
		"commission-max-change-rate": "%s",
		"min-self-delegation": "1"

	}`, ed25519PubKey, chain.Config().Denom, moniker, commissionRate, commissionMaxRate, commissionMaxChangeRate)

	err := chain.GetNode().WriteFile(ctx, []byte(content), file)
	require.NoError(t, err)

	filePath := fmt.Sprintf("%s/%s", chain.GetNode().HomeDir(), file)
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "create-validator", filePath}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func SubmitGovernanceProposalForValidatorChanges(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, validator string, power uint64, unsafe bool, authority string) string {
	powerMsg := []cosmos.ProtoMessage{
		&poa.MsgSetPower{
			Sender:           authority,
			ValidatorAddress: validator,
			Power:            power,
			Unsafe:           unsafe,
		},
	}

	title := fmt.Sprintf("Update" + validator + "Power")
	desc := fmt.Sprintf("Updating power for validator %s to %d", validator, power)

	proposal, err := chain.BuildProposal(powerMsg, title, desc, desc, fmt.Sprintf(`50%s`, chain.Config().Denom), user.FormattedAddress(), false)
	require.NoError(t, err, "error building proposal")

	fmt.Printf("proposal: %+v\n", proposal)

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}

func GetPOAConsensusPower(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, valoperAddr string) int64 {
	var res POAConsensusPower
	ExecuteQuery(ctx, chain, []string{"query", "poa", "consensus-power", valoperAddr}, &res)

	power := int64(0)
	_, err := fmt.Sscanf(res.Power, "%d", &power)
	fmt.Println("error", err) // does not panic incase omitempty

	return power
}

func GetPOAPending(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) POAPending {
	var res POAPending
	ExecuteQuery(ctx, chain, []string{"query", "poa", "pending-validators"}, &res)
	return res
}

func POAUpdateStakingParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, sp stakingtypes.Params) (sdk.TxResponse, error) {
	command := []string{"tx", "poa", "update-staking-params",
		sp.UnbondingTime.String(),
		fmt.Sprintf("%d", sp.MaxValidators),
		fmt.Sprintf("%d", sp.MaxEntries),
		fmt.Sprintf("%d", sp.HistoricalEntries),
		sp.BondDenom,
		fmt.Sprintf("%d", sp.MinCommissionRate),
	}

	cmd := TxCommandBuilder(ctx, chain, command, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}
