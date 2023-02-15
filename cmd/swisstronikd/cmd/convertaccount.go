package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

// ConvertAddressCmd returns add-genesis-account cobra Command.
func ConvertAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-address [cosmos_address]",
		Short: "Converts address from cosmos format to ethereum format",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cosmosAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			ethAddress := common.BytesToAddress(cosmosAddress.Bytes())
			println(ethAddress.String())

			return nil
		},
	}

	return cmd
}
