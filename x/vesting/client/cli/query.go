package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"swisstronik/x/vesting/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group vesting queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdGetBalances())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdGetBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [vesting-account]",
		Short: "Gets locked, unvested and vested tokens for a vesting account",
		Long:  "Gets locked, unvested and vested tokens for a vesting account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			if _, err = types.ParseAddress(args[0]); err != nil {
				return err
			}
			req := &types.QueryBalancesRequest{
				Address: args[0],
			}

			resp, err := queryClient.Balances(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
