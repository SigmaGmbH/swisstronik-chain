package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/common"
	"swisstronik/x/compliance/types"
	"time"

	"github.com/spf13/cobra"

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

	cmd.AddCommand(CmdSetAddressInfo())

	return cmd
}

// CmdSetAddressInfo command sets sample verification data for specific address.
// This function is used only for debug purposes and will be removed before chain update.
func CmdSetAddressInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-address-info [user-address] [issuer-address] [issuer-alias] [origin-chain]",
		Short: "Sets sample verification data for provided address",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			userAddress := args[0]
			if !common.IsHexAddress(userAddress) {
				return fmt.Errorf("provided non-eth user address")
			}

			issuerAddress := args[1]
			if !common.IsHexAddress(issuerAddress) {
				return fmt.Errorf("provided non-eth user address")
			}

			issuerAlias := args[2]
			if len(issuerAlias) == 0 {
				issuerAlias = "sample_issuer"
			}

			originChain := args[3]
			if len(originChain) == 0 {
				originChain = "swisstronik"
			}

			timestamp := uint32(time.Now().Unix())

			msg := types.NewMsgSetAddressInfo(
				clientCtx.GetFromAddress().String(),
				userAddress,
				issuerAddress,
				issuerAlias,
				originChain,
				timestamp,
				types.VerificationType_VT_KYC,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
