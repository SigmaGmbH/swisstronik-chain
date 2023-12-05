package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/hex"
	didutil "swisstronik/testutil/did"
	didcli "swisstronik/x/did/client/cli"
	didtypes "swisstronik/x/did/types"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
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

type KeyPair struct {
	PrivateKeyBase64 string `json:"private_key_base_64"`
	PublicKeyBase64  string `json:"public_key_base_64"`
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
	cmd.AddCommand(ExtractPubkeyCmd())
	cmd.AddCommand(ConvertAddressCmd())
	cmd.AddCommand(SampleDIDResource())
	cmd.AddCommand(SignDIDDocument())

	return cmd
}

// RandomEd25519PrivateKeypair returns random-ed25519-keypair cobra Command.
func RandomEd25519PrivateKeypair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "random-ed25519-keypair",
		Short: "Generates random ed25519 keypair",
		Long:  `Generates random ed25519 keypair and outputs it in JSON format with base64-encoded private and public keys. Do not use that keypair in production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			public, private, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return err
			}

			keyPair := struct {
				PrivateKeyBase64 string `json:"private_key_base_64"`
				PublicKeyBase64  string `json:"public_key_base_64"`
			}{
				PrivateKeyBase64: base64.StdEncoding.EncodeToString(private),
				PublicKeyBase64:  base64.StdEncoding.EncodeToString(public),
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
				publicKey  ed25519.PublicKey
				err        error
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
				Context:            []string{"https://www.w3.org/ns/did/v1"},
				ID:                 did,
				Authentication:     []string{keyId},
				VerificationMethod: []VerificationMethod{verificationMethod},
			}

			encodedDocument, err := json.Marshal(document)
			if err != nil {
				return err
			}

			// Construct sign inputs
			signInputs := SignInput{
				VerificationMethodID: keyId,
				PrivKey:              privateKey,
			}

			result := PayloadWithSignInputs{
				Payload:    encodedDocument,
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

func ReadKeyPairFromFile(file string) (KeyPair, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return KeyPair{}, err
	}

	keyPair := KeyPair{}
	err = json.Unmarshal(bytes, &keyPair)
	if err != nil {
		return KeyPair{}, err
	}

	return keyPair, nil
}

func SignDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-did-document [did.json] [key.json]",
		Short: "Generates signed DID document ready to be stored in DID registry",
		Long:  "Generates signed self-controlled DID document from the payload and key information provided, which is ready to be stored in DID registry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("invalid input parameters")
			}

			signInputs := make([]didcli.SignInput, 0)

			// Decode did.json to have payload
			payloadJSON, err := didcli.ReadPayloadFromFile(args[0])
			if err != nil {
				// Decode did.json to have payload & sign inputs
				payloadJSON, signInputs, err = didcli.ReadPayloadWithSignInputsFromFile(args[0])
			}

			if err != nil {
				return errors.New("invalid payload")
			}

			// Decode key.json to have private and public key pair
			keyPairFromFile, err := ReadKeyPairFromFile(args[1])
			if err != nil {
				return err
			}

			// Decode base64 based private key string to have byte[]
			privateKeyBytesFromFile, err := base64.StdEncoding.DecodeString(keyPairFromFile.PrivateKeyBase64)
			if err != nil {
				return err
			}

			// Encode ed25519 based private key
			privateKeyFromFile := ed25519.PrivateKey(privateKeyBytesFromFile)
			// Encode ed25519 based public key
			publicKeyFromFile := privateKeyFromFile.Public().(ed25519.PublicKey)
			// Multibase public key
			publicKeyMultibaseFromFile := didutil.GenerateEd25519VerificationKey2020VerificationMaterial(publicKeyFromFile)

			// Unmarshal spec-compliant payload
			var specPayload didcli.DIDDocument
			err = json.Unmarshal([]byte(payloadJSON), &specPayload)
			if err != nil {
				return err
			}

			if len(specPayload.VerificationMethod) < 1 {
				return errors.New("publicKeyMultibase is not specified")
			}

			validKey := false
			keyId := ""
			for _, v := range specPayload.VerificationMethod {
				// Check if public key is addressed in verfication method
				_, ok := v["publicKeyMultibase"]
				if !ok {
					continue
				}

				_, ok = v["id"]
				if !ok {
					continue
				}

				// Get multibase public key address
				publicKeyMultibase := v["publicKeyMultibase"].(string)

				// if there is matching verification method,
				if publicKeyMultibase == publicKeyMultibaseFromFile {
					keyId = v["id"].(string)
					validKey = true
					break
				}
			}

			// if there is no matching verification method
			if !validKey {
				return errors.New("invalid key information")
			}

			// Construct sign inputs
			if len(signInputs) < 1 {
				signInputs = append(signInputs, didcli.SignInput{
					VerificationMethodID: keyId,
					PrivKey:              privateKeyFromFile,
				})
			}

			// Construct payload with sign inputs
			result := didcli.PayloadWithSignInputs{
				Payload:    payloadJSON,
				SignInputs: signInputs,
			}

			// Encode the structured result
			encodedResult, err := json.Marshal(result)
			if err != nil {
				return err
			}

			// Print output.
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(encodedResult))
			return err
		},
	}

	return cmd
}

func SampleDIDResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sample-did-resource [existing-did] [base64-encoded-ed25519-private-key]",
		Short: "Generates sample DID resource ready to be stored in DID resource registry",
		Long: `Generates sample DID resource, which is ready to be stored in DID resource registry. 
		If private key was not provided, this command will generate random ed25519 keypair`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				privateKey ed25519.PrivateKey
				err        error
			)
			// Check if private key was provided.
			if len(args) == 1 {
				_, privateKey, err = ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return err
				}
			} else {
				privateKeyBytes, err := base64.StdEncoding.DecodeString(args[1])
				if err != nil {
					return err
				}
				privateKey = ed25519.PrivateKey(privateKeyBytes)
			}

			did := args[0]
			if !didtypes.IsValidDID(did, didtypes.DIDMethod) {
				return fmt.Errorf("provided DID is invalid")
			}

			// Derive collection id from provided DID
			_, collectionId, err := didtypes.TrySplitDID(did)
			if err != nil {
				return err
			}

			resource := didtypes.MsgCreateResourcePayload{
				CollectionId: collectionId,
				Id:           uuid.NewString(),
				Name:         "sample-resource",
				Version:      "sample-version",
				ResourceType: "SampleResourceType",
				AlsoKnownAs: []*didtypes.AlternativeUri{
					{
						Uri:         "http://example.com/example-did",
						Description: "http-uri",
					},
				},
			}

			encodedResource, err := json.Marshal(resource)
			if err != nil {
				return err
			}

			// Construct sign inputs
			keyId := did + "#key1"
			signInputs := SignInput{
				VerificationMethodID: keyId,
				PrivKey:              privateKey,
			}

			result := PayloadWithSignInputs{
				Payload:    encodedResource,
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

func ConvertAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "convert-address [address]",
		Short: "Convert an address between hex and bech32",
		Long:  "Convert an address between hex encoding and bech32.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrString := args[0]
			cfg := sdk.GetConfig()

			var addr []byte
			switch {
			case common.IsHexAddress(addrString):
				addr = common.HexToAddress(addrString).Bytes()
			case strings.HasPrefix(addrString, cfg.GetBech32ValidatorAddrPrefix()):
				addr, _ = sdk.ValAddressFromBech32(addrString)
			case strings.HasPrefix(addrString, cfg.GetBech32AccountAddrPrefix()):
				addr, _ = sdk.AccAddressFromBech32(addrString)
			default:
				return fmt.Errorf("expected a valid hex or bech32 address (acc prefix %s), got '%s'", cfg.GetBech32AccountAddrPrefix(), addrString)
			}

			cmd.Println("Address bytes:", addr)
			cmd.Printf("Address (hex): %s\n", bytes.HexBytes(addr))
			cmd.Printf("Address (EIP-55): %s\n", common.BytesToAddress(addr))
			cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr))
			cmd.Printf("Bech32 Val: %s\n", sdk.ValAddress(addr))
			return nil
		},
	}
}

// getPubKeyFromString decodes SDK PubKey using JSON marshaler.
func getPubKeyFromString(ctx client.Context, pkstr string) (cryptotypes.PubKey, error) {
	var pk cryptotypes.PubKey
	err := ctx.Codec.UnmarshalInterfaceJSON([]byte(pkstr), &pk)
	return pk, err
}

func ExtractPubkeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extract-pubkey [pubkey]",
		Short: "Decode a pubkey from proto JSON",
		Long:  "Decode a pubkey from proto JSON and display it's address",
		Example: fmt.Sprintf(
			`"$ %s debug pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"}'`,
			version.AppName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pk, err := getPubKeyFromString(clientCtx, args[0])
			if err != nil {
				return err
			}

			addr := pk.Address()
			cmd.Printf("Address (EIP-55): %s\n", common.BytesToAddress(addr))
			cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr))
			cmd.Println("PubKey Hex:", hex.EncodeToString(pk.Bytes()))
			return nil
		},
	}
}
