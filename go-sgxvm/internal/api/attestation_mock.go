//go:build !nosgx && !attestationServer
// +build !nosgx,!attestationServer

package api

import (
	"fmt"
)

// StartAttestationServer starts attestation server
func StartAttestationServer(dcapAddress string) error {
	fmt.Println("[Attestation Server] Not enabled")
	return nil
}
