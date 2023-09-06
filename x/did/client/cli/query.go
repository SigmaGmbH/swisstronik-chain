package cli

import (
	"fmt"
	"context"

	"swisstronik/x/did/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdGetDIDDocument(),
		CmdGetDIDDocumentVersion(),
		CmdGetAllDidDocVersionsMetadata(),
	)

	return cmd
}

func CmdGetDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document [id]",
		Short: "Query a DID Document by DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			did := args[0]
			params := &types.QueryDIDDocumentRequest{
				Id: did,
			}

			resp, err := queryClient.DIDDocument(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetDIDDocumentVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document-version [id] [version]",
		Short: "Query specific version of a DID Document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			did := args[0]
			versionID := args[1]
			params := &types.QueryDIDDocumentVersionRequest{
				Id:      did,
				Version: versionID,
			}

			resp, err := queryClient.DIDDocumentVersion(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetAllDidDocVersionsMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata [id]",
		Short: "Query all versions metadata for a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			did := args[0]
			params := &types.QueryAllDIDDocumentVersionsMetadataRequest{
				Id: did,
			}

			resp, err := queryClient.AllDIDDocumentVersionsMetadata(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}