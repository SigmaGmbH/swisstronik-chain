package cli

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"swisstronik/x/compliance/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Compliance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSetIssuerDetails(),
		CmdUpdateIssuerDetails(),
		CmdRemoveIssuer(),
	)

	return cmd
}

// CmdSetIssuerDetails command sets provided issuer details.
func CmdSetIssuerDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-issuer-details [issuer-address] [name] [description] [url] [logo-url] [legalEntity]",
		Short: "Sets issuer details",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuerAddress, err := types.ParseAddress(args[0])
			if err != nil {
				return err
			}

			issuerName := args[1]
			issuerDescription := args[2]
			issuerURL := args[3]
			issuerLogo := args[4]
			issuerLegalEntity := args[5]

			msg := types.NewSetIssuerDetailsMsg(
				clientCtx.GetFromAddress().String(),
				issuerAddress.String(),
				issuerName,
				issuerDescription,
				issuerURL,
				issuerLogo,
				issuerLegalEntity,
			)

			_ = clientCtx.PrintProto(&msg)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateIssuerDetails command updates existing issuer details.
func CmdUpdateIssuerDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-issuer-details [issuer-address] [new-operator] [name] [description] [url] [logo-url] [legalEntity]",
		Short: "Update issuer details",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuerAddress, err := types.ParseAddress(args[0])
			if err != nil {
				return err
			}

			newOperator := args[1]
			issuerName := args[2]
			issuerDescription := args[3]
			issuerURL := args[4]
			issuerLogo := args[5]
			issuerLegalEntity := args[6]

			msg := types.NewUpdateIssuerDetailsMsg(
				clientCtx.GetFromAddress().String(),
				newOperator,
				issuerAddress.String(),
				issuerName,
				issuerDescription,
				issuerURL,
				issuerLogo,
				issuerLegalEntity,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveIssuer command removes existing issuer.
func CmdRemoveIssuer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-issuer [issuer-address]",
		Short: "Removes existing issuer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuerAddress, err := types.ParseAddress(args[0])
			if err != nil {
				return err
			}

			msg := types.NewRemoveIssuerMsg(
				clientCtx.GetFromAddress().String(),
				issuerAddress.String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
