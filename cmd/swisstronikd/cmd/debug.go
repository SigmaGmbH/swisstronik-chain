package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"encoding/hex"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

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
	cmd.AddCommand(ExtractPubkeyCmd())
	cmd.AddCommand(ConvertAddressCmd())

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
