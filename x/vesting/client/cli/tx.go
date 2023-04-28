package cli

import (
	"strconv"
	"swisstronik/x/vesting/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Monthly vesting transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateMonthlyVestingAccount())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdCreateMonthlyVestingAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-monthly-vesting-account [to-address] [start-time] [amount] [month]",
		Short: "Broadcast message create-monthly-vesting-account",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			argToAddress := args[0]
			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			argStartTime, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			argAmount, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			argMonth, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateMonthlyVestingAccount(
				clientCtx.GetFromAddress().String(),
				argToAddress,
				argStartTime,
				argAmount,
				argMonth,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
