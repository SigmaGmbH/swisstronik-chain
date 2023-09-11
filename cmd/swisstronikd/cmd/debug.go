package cmd

import (
	"fmt"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	didutil "swisstronik/testutil/did"
	didtypes "swisstronik/x/did/types"
)

type JsonSignInfo struct {
	VerificationMethodId string `json:"verificationMethodId"`
	PrivateKey ed25519.PrivateKey `json:"privKey"`
}

// Cmd creates a CLI main command
func DebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Commands for debug",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(RandomEd25519PrivateKeypair())
	cmd.AddCommand(SampleDIDDocument())

	return cmd
}

// RandomEd25519PrivateKeypair returns random-ed25519-keypair cobra Command.
func RandomEd25519PrivateKeypair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "random-ed25519-keypair",
		Short: "Generates random ed25519 keypair",
		Long: `Generates random ed25519 keypair and outputs it in JSON format with base64-encoded private and public keys. Do not use that keypair in production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			public, private, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return err
			}

			keyPair := struct {
				PrivateKeyBase64 string `json:"private_key_base_64"`
				PublicKeyBase64 string `json:"public_key_base_64"`
			} {
				PrivateKeyBase64: base64.StdEncoding.EncodeToString(private),
				PublicKeyBase64: base64.StdEncoding.EncodeToString(public),
			}

			jsonKeyPair, err := json.Marshal(keyPair)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(jsonKeyPair))
			return err
		},
	}

	return cmd
}

func SampleDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sample-did-document [base64-encoded-ed25519-private-key]",
		Short: "Generates sample DID document ready to be stored in DID registry",
		Long: `Generates sample self-controlled DID document, which is ready to be stored in DID registry. 
		If private key was not provided, this command will generate random ed25519 keypair`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				privateKey ed25519.PrivateKey
				publicKey ed25519.PublicKey
				err error
			)
			// Check if private key was provided. 
			if len(args) != 1 {
				publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return err
				}
			} else {
				privateKeyBytes, err := base64.StdEncoding.DecodeString(args[0])
				if err != nil {
					return err
				}
				privateKey = ed25519.PrivateKey(privateKeyBytes)
				publicKey = privateKey.Public().(ed25519.PublicKey)
			}

			// Generate random DID and key id
			did := didutil.GenerateDID(didutil.Base58_16bytes)
			keyId := did + "#key1" 

			// Construct DID document
			document := didtypes.DIDDocument{
				Context: []string{"https://www.w3.org/ns/did/v1"},
				Id: did,
				Authentication: []string{keyId},
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						Id: keyId,
						VerificationMethodType: didtypes.Ed25519VerificationKey2018Type,
						Controller: did,
						VerificationMaterial: didutil.GenerateEd25519VerificationKey2018VerificationMaterial(publicKey),
					},
				},
			}

			// Construct sign inputs
			signInputs := JsonSignInfo {
				VerificationMethodId: keyId,
				PrivateKey: privateKey,
			}

			result := struct {
				Document didtypes.DIDDocument `json:"payload"`
				SignInputs []JsonSignInfo `json:"signInputs"`		
			} {
				Document: document,
				SignInputs: []JsonSignInfo{signInputs},
			}

			encodedResult, err := json.Marshal(result)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(encodedResult))
			return err
		},
	}

	return cmd
}