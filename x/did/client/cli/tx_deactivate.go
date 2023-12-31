package cli

import (
	"swisstronik/x/did/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func CmdDeactivateDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-document [payload-file] --version-id [version-id]",
		Short: "Deactivate a DID.",
		Long: `Deactivates a DID and its associated DID Document. 
[payload-file] is JSON encoded MsgDeactivateDidDocPayload alongside with sign inputs. 

NOTES:
1. A new DID Document version is created when deactivating a DID Document so that the operation timestamp can be recorded. Version ID is optional and is determined by the '--version-id' flag. If not provided, a random UUID will be used as version-id.
2. Payload file should be a JSON file containing the properties given in example below.
3. Private key provided in sign inputs is ONLY used locally to generate signature(s) and not sent to the ledger.

Example payload file:
{
    "payload": {
        "id": "did:swtr:<unique-identifier>"
    },
    "signInputs": [
        {
            "verificationMethodId": "did:swtr:<unique-identifier>#<key-id>",
            "privKey": "<private-key-bytes-encoded-to-base64>"
        }
    ]
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read payload file arg
			payloadFile := args[0]

			// Read version-id flag
			versionID, err := cmd.Flags().GetString(FlagVersionID)
			if err != nil {
				return err
			}

			if versionID != "" {
				err = types.ValidateUUID(versionID)
				if err != nil {
					return err
				}
			} else {
				versionID = uuid.NewString()
			}

			payloadJSON, signInputs, err := ReadPayloadWithSignInputsFromFile(payloadFile)
			if err != nil {
				return err
			}

			// Build payload
			payload := types.MsgDeactivateDIDDocumentPayload{}
			err = clientCtx.Codec.UnmarshalJSON([]byte(payloadJSON), &payload)
			if err != nil {
				return err
			}

			// Set version id from flag or random
			payload.VersionId = versionID

			// Build identity message
			signBytes := payload.GetSignBytes()
			identitySignatures := SignWithSignInputs(signBytes, signInputs)

			msg := types.MsgDeactivateDIDDocument{
				Payload:    &payload,
				Signatures: identitySignatures,
			}

			// Set fee-payer if not set
			err = SetFeePayerFromSigner(&clientCtx)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	// add custom / override flags
	cmd.Flags().String(FlagVersionID, "", "Version ID of the DID Document")

	return cmd
}