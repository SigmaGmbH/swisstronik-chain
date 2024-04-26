package cmd

import (
	"github.com/SigmaGmbH/librustgo/internal/api"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attestation-server",
		Short: "Commands for interaction with Swisstronik Attestation Server",
	}

	cmd.AddCommand(
		StartAttestationServer(),
	)

	return cmd
}

// StartAttestationServer returns start-attestation-server cobra Command.
func StartAttestationServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-server [epid-address-with-port] [dcap-address-with-port]",
		Short: "Starts attestation server",
		Long:  "Start server for Intel SGX Remote Attestation to share encryption keys with new nodes",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := api.StartAttestationServer(args[0], args[1]); err != nil {
				return err
			}
			return WaitForQuitSignals()
		},
	}

	return cmd
}
