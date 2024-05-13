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

// EnclaveCmd creates a CLI main command
func EnclaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enclave",
		Short: "Commands for interaction with Swisstronik SGX Enclave",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(
		EPIDRemoteAttestationCmd(),
		DCAPRemoteAttestationCmd(),
		Status(),
	)
	return cmd
}

// EPIDRemoteAttestationCmd returns request-epoch-keys-epid cobra Command.
func EPIDRemoteAttestationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-epoch-keys-epid [attestation-server-address]",
		Short: "Requests epoch keys from attestation server using EPID",
		Long: `Initializes SGX enclave by passing process of EPID Remote Attestation agains attestation server. 
		If remote attestation was successful, attestation server shares encrypted epoch keys with this node. 
		Process of Remote Attestation is performed over pure TCP protocol.`,
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

			if err := librustgo.RequestEpochKeys(host, port, false); err != nil {
				return err
			}

			fmt.Println("EPID Remote Attestation passed. Node is ready for work")
			return nil
		},
	}

	return cmd
}

// DCAPRemoteAttestationCmd returns request-epochs-keys-dcap cobra Command.
func DCAPRemoteAttestationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-epoch-keys-dcap [attestation-server-address]",
		Short: "Requests epoch keys from attestation server using DCAP",
		Long: `Initializes SGX enclave by passing process of DCAP Remote Attestation agains attestation server. 
		If remote attestation was successful, attestation server node shares encrypted epoch keys with this node. 
		Process of Remote Attestation is performed over pure TCP protocol.`,
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

			if err := librustgo.RequestEpochKeys(host, port, true); err != nil {
				return err
			}

			fmt.Println("DCAP Remote Attestation passed. Node is ready for work")
			return nil
		},
	}

	return cmd
}

// Status checks if Intel SGX Enclave is accessible and if Intel SGX was properly configured
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
