package cli

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"

	"swisstronik/x/compliance/types"
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

func CmdVerifyIssuerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify-issuer [issuer-address]",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Submit a proposal to verify issuer",
		Long:    "Submit a proposal to verify issuer along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal verify-issuer <issuer address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription) //nolint:staticcheck
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			issuerAddress := args[0]
			from := clientCtx.GetFromAddress()

			// Verified issuer can't create proposal
			queryClient := types.NewQueryClient(clientCtx)
			addressDetails, err := queryClient.AddressDetails(cmd.Context(), &types.QueryAddressDetailsRequest{
				Address: issuerAddress,
			})
			if err != nil {
				return err
			}
			if addressDetails.Data.IsVerified {
				return errors.New("issuer was already verified")
			}

			content := types.NewVerifyIssuerProposal(title, description, issuerAddress)

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	// TODO, should be renamed to `aswtr` when merge into main
	cmd.Flags().String(cli.FlagDeposit, "1uswtr", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}
