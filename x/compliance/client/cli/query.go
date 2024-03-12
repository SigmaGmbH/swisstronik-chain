package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"swisstronik/x/compliance/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdGetAddressInfo(),
		CmdGetIssuerDetails(),
	)

	return cmd
}

func CmdGetAddressInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-address-info [bech32-or-hex-address]",
		Short: "Returns AddressInfo associated with provided address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			address, err := types.ParseAddress(args[0])
			if err != nil {
				return err
			}

			req := &types.QueryVerificationDataRequest{
				Address: address.String(),
			}

			resp, err := queryClient.VerificationData(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetIssuerDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-issuer-details [bech32-or-hex-address]",
		Short: "Returns details of provided address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			address, err := types.ParseAddress(args[0])
			if err != nil {
				return err
			}

			req := &types.QueryIssuerDetailsRequest{
				IssuerAddress: address.String(),
			}

			resp, err := queryClient.IssuerDetails(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
