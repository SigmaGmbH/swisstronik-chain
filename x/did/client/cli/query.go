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
		CmdGetCollectionResources(),
		CmdGetResource(),
		CmdGetResourceMetadata(),
		CmdGetControlledDocuments(),
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
		Use:   "document-metadata [id]",
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

func CmdGetCollectionResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection-metadata [collection-id]",
		Short: "Query metadata for an entire Collection",
		Long: `Query metadata for an entire Collection by Collection ID. This will return the metadata for all Resources in the Collection.
		
		Collection ID is the UNIQUE IDENTIFIER part of the DID the resource is linked to.
		Example: c82f2b02-bdab-4dd7-b833-3e143745d612, wGHEXrZvJxR8vw5P3UWH1j, etc.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			collectionID := args[0]

			params := &types.QueryCollectionResourcesRequest{
				CollectionId: collectionID,
			}

			resp, err := queryClient.CollectionResources(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "specific-resource [collection-id] [resource-id]",
		Short: "Query a specific resource",
		Long: `Query a specific resource by Collection ID and Resource ID.
		
		Collection ID is the UNIQUE IDENTIFIER part of the DID the resource is linked to.
		Example: c82f2b02-bdab-4dd7-b833-3e143745d612, wGHEXrZvJxR8vw5P3UWH1j, etc.

		Resource ID is the UUID of the specific resource.
		Example: 6e8bc430-9c3a-11d9-9669-0800200c9a66, 6e8bc430-9c3a-11d9-9669-0800200c9a67, etc.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			collectionID := args[0]
			id := args[1]

			params := &types.QueryResourceRequest{
				CollectionId: collectionID,
				Id:           id,
			}

			resp, err := queryClient.Resource(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetResourceMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-metadata [collection-id] [resource-id]",
		Short: "Query metadata for a specific resource",
		Long: `Query metadata for a specific resource by Collection ID and Resource ID.
		
		Collection ID is the UNIQUE IDENTIFIER part of the DID the resource is linked to.
		Example: c82f2b02-bdab-4dd7-b833-3e143745d612, wGHEXrZvJxR8vw5P3UWH1j, etc.

		Resource ID is the UUID of the specific resource.
		Example: 6e8bc430-9c3a-11d9-9669-0800200c9a66, 6e8bc430-9c3a-11d9-9669-0800200c9a67, etc.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			collectionID := args[0]
			id := args[1]

			params := &types.QueryResourceMetadataRequest{
				CollectionId: collectionID,
				Id:           id,
			}

			resp, err := queryClient.ResourceMetadata(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetControlledDocuments() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controlled-documents [verification-material]",
		Short: "Query all DID Documents controlled by given verification material",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			verificationMaterial := args[0]

			params := &types.QueryAllControlledDIDDocumentsRequest{
				VerificationMaterial: verificationMaterial,
			}

			resp, err := queryClient.AllControlledDIDDocuments(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}