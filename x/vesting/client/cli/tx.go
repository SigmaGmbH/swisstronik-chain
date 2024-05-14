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

const (
	FlagMonthlyVesting = "monthly-vesting"
	FlagCliffDays      = "cliff-days"
	FlagVestingPeriods = "vesting-months"
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
		Use:   "create-monthly-vesting-account [to-address] [cliff-days] [months] [amount]",
		Short: "Create a new vesting account funded with an allocation of tokens with linear monthly vesting and cliff feature.",
		Long: `Create a new vesting account funded with an allocation of tokens. The token allowed will start
vested, tokens will be released only after the cliff days linearly relative start time for number of months.
All vesting accounts created will have their start time set by committed block's time.
e.g. User has cliff for 30 days and 12-month vesting period. After 30 days, linear monthly vesting starts.
So after another 30 days, 1/12 of vesting amount will be released. 
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			to := args[0]
			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliffDays, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			months, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateMonthlyVestingAccount(
				clientCtx.GetFromAddress().String(),
				to,
				cliffDays,
				months,
				amount,
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
