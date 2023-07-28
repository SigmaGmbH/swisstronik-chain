package cmd

import (
	"net"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
	"github.com/SigmaGmbH/librustgo"
)

const flagShouldReset = "reset"

// Cmd creates a CLI main command
func EnclaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enclave",
		Short: "Commands for interaction with Swisstronik SGX Enclave",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(RequestMasterKeyCmd())
	cmd.AddCommand(CreateMasterKey())

	return cmd
}

// RequestMasterKeyCmd returns request-master-key cobra Command.
func RequestMasterKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-master-key [bootstrap-node-address]",
		Short: "Requests master key from bootstrap node",
		Long: `Initializes SGX enclave by passing process of Remote Attestation agains bootstrap node. If remote attestation was successful, bootstrap node shares encrypted master key with this node. Process of Remote Attestation is performed over pure TCP protocol.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			host, portString, err := net.SplitHostPort(args[0])
			if err != nil {
				return err
			}

			port, err := strconv.Atoi(portString)
			if err != nil {
				return err
			}

			if err := librustgo.RequestSeed(host, port); err != nil {
				return err
			}

			fmt.Println("Remote Attestation passed. Node is ready for work")
			return nil
		},
	}

	return cmd
}

// CreateMasterKey returns create-master-key cobra Command.
func CreateMasterKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-master-key",
		Short: "Creates new master key",
		Long: `Initializes SGX enclave by creating new master key. Use this function for first validator in network`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldReset, err := cmd.Flags().GetBool(flagShouldReset)
			if err != nil {
				return err
			}

			if err := librustgo.InitializeMasterKey(shouldReset); err != nil {
				return err
			}

			fmt.Println("Node is ready for work")

			return nil
		},
	}

	cmd.Flags().Bool(flagShouldReset, false, "reset already existing master key. Default: false")

	return cmd
}
