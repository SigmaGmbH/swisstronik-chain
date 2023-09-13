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

type DIDDocument struct {
	Context              []string             `json:"context"`
	ID                   string               `json:"id"`
	Controller           []string             `json:"controller,omitempty"`
	VerificationMethod   []VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication       []string             `json:"authentication,omitempty"`
	AssertionMethod      []string             `json:"assertionMethod,omitempty"`
	CapabilityInvocation []string             `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string             `json:"capabilityDelegation,omitempty"`
	KeyAgreement         []string             `json:"keyAgreement,omitempty"`
	Service              []Service            `json:"service,omitempty"`
	AlsoKnownAs          []string             `json:"alsoKnownAs,omitempty"`
}

type VerificationMethod map[string]any

type PayloadWithSignInputs struct {
	Payload    json.RawMessage
	SignInputs []SignInput
}

type SignInput struct {
	VerificationMethodID string
	PrivKey              ed25519.PrivateKey
}

type Service struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	ServiceEndpoint []string `json:"serviceEndpoint"`
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

			// Construct verification method
			verificationMethod := make(map[string]any)
			verificationMethod["id"] = keyId
			verificationMethod["type"] = didtypes.Ed25519VerificationKey2020Type
			verificationMethod["controller"] = did
			verificationMethod["publicKeyMultibase"] = didutil.GenerateEd25519VerificationKey2020VerificationMaterial(publicKey)

			// Construct DID document
			document := DIDDocument{
				Context: []string{"https://www.w3.org/ns/did/v1"},
				ID: did,
				Authentication: []string{keyId},
				VerificationMethod: []VerificationMethod{verificationMethod},
			}

			encodedDocument, err := json.Marshal(document)
			if err != nil {
				return err
			}

			// Construct sign inputs
			signInputs := SignInput {
				VerificationMethodID: keyId,
				PrivKey: privateKey,
			}

			result := PayloadWithSignInputs{
				Payload: encodedDocument,
				SignInputs: []SignInput{signInputs},
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