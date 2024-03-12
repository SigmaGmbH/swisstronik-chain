package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"strings"
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
			cfg := sdk.GetConfig()

			var address sdk.AccAddress
			var err error
			if !strings.HasPrefix(args[0], cfg.GetBech32AccountAddrPrefix()) {
				// Assume that was provided eth address
				ethAddress := common.HexToAddress(args[0])
				address = ethAddress.Bytes()
			} else {
				// Assume that was provided bech32 address
				address, err = sdk.AccAddressFromBech32(args[0])
				if err != nil {
					return err
				}
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
