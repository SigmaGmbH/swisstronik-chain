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
	)

	return cmd
}

// CmdSetIssuerDetails command sets provided issuer details.
func CmdSetIssuerDetails() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-issuer-details [issuer-address] [name] [description] [url] [logo-url] [legalEntity]",
		Short: "Sets issuer details",
		Args:  cobra.ExactArgs(2),
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
				clientCtx.GetFromAddress(),
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
