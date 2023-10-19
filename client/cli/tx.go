package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

// NewTxCmd returns a root CLI command handler for all x/POA transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        poa.ModuleName,
		Short:                      poa.ModuleName + "transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSetPowerCmd(),
		NewRemoveValidatorCmd(),
	)
	return txCmd
}

func NewSetPowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "set-power [validator] [power]",
		Args: cobra.ExactArgs(2),
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

			msg := &poa.MsgSetPower{
				FromAddress:      clientCtx.GetFromAddress().String(),
				ValidatorAddress: validator,
				Power:            power,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewRemoveValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "remove [validator]",
		Args: cobra.ExactArgs(1),
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
				FromAddress:      clientCtx.GetFromAddress().String(),
				ValidatorAddress: validator,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
