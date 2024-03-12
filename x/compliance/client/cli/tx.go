package cli

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/common"
	"swisstronik/x/compliance/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"time"
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

	cmd.AddCommand(CmdSetVerificationData())

	return cmd
}

// CmdSetVerificationData command sets verification data for specific address.
// This function is used only for debug purposes and will be removed before chain update.
func CmdSetVerificationData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-verification-data [userAddress] [issuerAddress]",
		Short: "Sets verification data for provided address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			userAddress := common.HexToAddress(args[0])
			issuerAddress := common.HexToAddress(args[1])

			adapterData := types.IssuerAdapterContractDetail{
				IssuerAlias:     issuerAddress.String(),
				ContractAddress: issuerAddress.Bytes(),
			}

			entry := types.VerificationEntry{
				AdapterData:         &adapterData,
				OriginChain:         "swisstronik",
				IssuanceTimestamp:   uint32(time.Now().Unix()),
				ExpirationTimestamp: 0,
				OriginalData:        nil,
			}
			verificationData := types.VerificationData{
				VerificationType: 0,
				Entries:          []*types.VerificationEntry{&entry},
			}

			msg := types.MsgSetVerificationData{
				UserAddress: userAddress.String(),
				Data:        &verificationData,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
