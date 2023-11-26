package helpers

import (
	"context"
	"fmt"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func POASetPower(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string, power int64, flags ...string) (TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "set-power", valoper, fmt.Sprintf("%d", power)}, user.KeyName(), flags...)
	return ExecuteTransaction(ctx, chain, cmd)
}

func POARemove(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (TxResponse, error) {
	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "remove", valoper}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func SubmitGovernanceProposalForValidatorChanges(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, validator string, power uint64, unsafe bool) string {
	govAddr := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	powerMsg := []cosmosproto.Message{
		&poa.MsgSetPower{
			Sender:           govAddr,
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

func GetPOAParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) poa.Params {
	var res poa.ParamsResponse
	ExecuteQuery(ctx, chain, []string{"query", "poa", "params"}, &res)
	return res.Params
}

func GetPOAConsensusPower(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, valoperAddr string) int64 {
	var res POAConsensusPower
	ExecuteQuery(ctx, chain, []string{"query", "poa", "power", valoperAddr}, &res)

	var power int64
	_, err := fmt.Sscanf(res.Power, "%d", &power)
	if err != nil {
		return 0
	}

	return power
}

func POAUpdateParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, admins []string) (TxResponse, error) {
	// admin1,admin2,admin3
	adminList := ""
	for _, admin := range admins {
		adminList += admin + ","
	}
	adminList = adminList[:len(adminList)-1]

	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "poa", "update-params", adminList}, user.KeyName())
	return ExecuteTransaction(ctx, chain, cmd)
}

func POAUpdateStakingParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, sp stakingtypes.Params) (TxResponse, error) {
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
