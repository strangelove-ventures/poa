package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

const (
	FlagNodeID  = "node-id"
	FlagIP      = "ip"
	FlagP2PPort = "p2p-port"
)

// NewTxCmd returns a root CLI command handler for all x/POA transaction commands.
func NewTxCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        poa.ModuleName,
		Short:                      poa.ModuleName + "transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewCreateValidatorCmd(ac),
		NewSetPowerCmd(ac),
		NewRemovePendingCmd(),
		NewRemoveValidatorCmd(),
		NewUpdateParamsCmd(),
		NewUpdateStakingParamsCmd(),
	)

	return txCmd
}

func NewSetPowerCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-power [validator] [power] [--unsafe]",
		Short: "set the consensus power of a validator in the active set",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("GetClientTxContext failed: %w", err)
			}

			validator := args[0]
			_, err = sdk.ValAddressFromBech32(validator)
			if err != nil {
				return fmt.Errorf("ValAddressFromBech32 failed: %w", err)
			}

			power, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("strconv.ParseUint failed: %w", err)
			}

			unsafeAction, err := cmd.Flags().GetBool("unsafe")
			if err != nil {
				return fmt.Errorf("get unsafe flag failed: %w", err)
			}

			msg := &poa.MsgSetPower{
				Sender:           clientCtx.GetFromAddress().String(),
				ValidatorAddress: validator,
				Power:            power,
				Unsafe:           unsafeAction,
			}

			if err := msg.Validate(ac); err != nil {
				return fmt.Errorf("msg.Validate failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool("unsafe", false, "set power without checking if validator is in the validator set")

	return cmd
}

func NewRemoveValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [validator]",
		Short: "remove a validator from the active set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validator := args[0]
			_, err = sdk.ValAddressFromBech32(validator)
			if err != nil {
				return fmt.Errorf("ValAddressFromBech32 failed: %w", err)
			}

			msg := &poa.MsgRemoveValidator{
				Sender:           clientCtx.GetFromAddress().String(),
				ValidatorAddress: validator,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewRemovePendingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-pending [validator]",
		Short: "remove a validator from the pending set queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validator := args[0]
			_, err = sdk.ValAddressFromBech32(validator)
			if err != nil {
				return fmt.Errorf("ValAddressFromBech32 failed: %w", err)
			}

			msg := &poa.MsgRemovePending{
				Sender:           clientCtx.GetFromAddress().String(),
				ValidatorAddress: validator,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [allow-validator-self-exit-bool]",
		Short: "update the PoA module params",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			allowGracefulExit, err := strconv.ParseBool(args[0])
			if err != nil {
				return fmt.Errorf("strconv.ParseBool failed: %w", err)
			}

			p, err := poa.NewParams(allowGracefulExit)
			if err != nil {
				return fmt.Errorf("NewParams failed: %w", err)
			}

			msg := &poa.MsgUpdateParams{
				Sender: clientCtx.GetFromAddress().String(),
				Params: p,
			}

			if err := msg.Params.Validate(); err != nil {
				return fmt.Errorf("msg.Params.Validate failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUpdateStakingParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-staking-params [unbondingTime] [maxVals] [maxEntries] [historicalEntries] [bondDenom] [minCommissionRate]",
		Short: "update the staking module params",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			unbondingTime, err := time.ParseDuration(args[0])
			if err != nil {
				return fmt.Errorf("unbondingTime time.ParseDuration failed: %w", err)
			}

			maxVals, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("maxVals strconv.ParseUint failed: %w", err)
			}

			maxEntries, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("maxEntries strconv.ParseUint failed: %w", err)
			}

			historicalEntries, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("historicalEntries strconv.ParseUint failed: %w", err)
			}

			bondDenom := args[4]
			if bondDenom == "" {
				return fmt.Errorf("bondDenom is empty")
			}

			minCommission, err := math.LegacyNewDecFromStr(args[5])
			if err != nil {
				return fmt.Errorf("minCommission math.LegacyNewDecFromStr failed: %w", err)
			}

			msg := &poa.MsgUpdateStakingParams{
				Sender: clientCtx.GetFromAddress().String(),
				Params: poa.StakingParams{
					UnbondingTime:     unbondingTime,
					MaxValidators:     uint32(maxVals),
					MaxEntries:        uint32(maxEntries),
					HistoricalEntries: uint32(historicalEntries),
					BondDenom:         bondDenom,
					MinCommissionRate: minCommission,
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCreateValidatorCmd returns a CLI command handler for creating a MsgCreateValidator transaction.
func NewCreateValidatorCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [path/to/validator.json]",
		Short: "create new validator for POA (anyone)",
		Args:  cobra.ExactArgs(1),
		Long:  `Create a new validator creates the new validator for the POA module.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx poa create-validator path/to/validator.json --from keyname

Where validator.json contains:

{
	"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
	"amount": "1stake", # ignored
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1" # ignored
}

where we can get the pubkey using "%s tendermint show-validator"
`, version.AppName, version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			validator, err := parseAndValidateValidatorJSON(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildCreateValidatorMsg(clientCtx, txf, cmd.Flags(), validator, ac)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}
func newBuildCreateValidatorMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet, val validator, valAc address.Codec) (tx.Factory, *poa.MsgCreateValidator, error) {
	valAddr := clientCtx.GetFromAddress()

	description := poa.NewDescription(
		val.Moniker,
		val.Identity,
		val.Website,
		val.Security,
		val.Details,
	)

	commissionRates := poa.NewCommissionRates(
		val.CommissionRates.Rate,
		val.CommissionRates.MaxRate,
		val.CommissionRates.MaxChangeRate,
	)

	valStr, err := valAc.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txf, nil, err
	}
	msg, err := poa.NewMsgCreateValidator(
		valStr, val.PubKey, description, commissionRates, val.MinSelfDelegation,
	)
	if err != nil {
		return txf, nil, err
	}
	if err := msg.Validate(valAc); err != nil {
		return txf, nil, err
	}

	genOnly, _ := fs.GetBool(flags.FlagGenerateOnly)
	if genOnly {
		ip, _ := fs.GetString(FlagIP)
		p2pPort, _ := fs.GetUint(FlagP2PPort)
		nodeID, _ := fs.GetString(FlagNodeID)

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txf, msg, nil
}
