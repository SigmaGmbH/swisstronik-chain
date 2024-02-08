package cmd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/SigmaGmbH/librustgo"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(StartAttestationServer())
	cmd.AddCommand(Status())

	return cmd
}

// RequestMasterKeyCmd returns request-master-key cobra Command.
func RequestMasterKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-master-key [bootstrap-node-address]",
		Short: "Requests master key from bootstrap node",
		Long:  "Initializes SGX enclave by passing process of Remote Attestation agains bootstrap node. If remote attestation was successful, bootstrap node shares encrypted master key with this node. Process of Remote Attestation is performed over pure TCP protocol.",
		Args:  cobra.ExactArgs(1),
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

// StartAttestationServer returns start-attestation-server cobra Command.
func StartAttestationServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-attestation-server [address-with-port]",
		Short: "Starts attestation server",
		Long:  "Start server for Intel SGX Remote Attestation to share master key with new nodes",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := librustgo.StartSeedServer(args[0]); err != nil {
				return err
			}
			return server.WaitForQuitSignals()
		},
	}

	return cmd
}

// Healthcheck checks if Intel SGX Enclave is accessible and if Intel SGX was properly configured
func Status() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Checks status of Intel SGX Enclave",
		Long:  "Checks if Intel SGX Enclave is accessible and if Intel SGX was properly configured",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return librustgo.CheckNodeStatus()
		},
	}

	return cmd
}
