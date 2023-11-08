package cli

import (
	"os"

	"swisstronik/x/did/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func CmdCreateResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-resource [payload-file] [resource-data-file]",
		Short: "Create a new Resource.",
		Long: `Create a new Resource within a DID Resource Collection. 
[payload-file] is JSON encoded MsgCreateResourcePayload alongside with sign inputs. 
[resource-data-file] is a path to the Resource data file.

NOTES:
1. Payload file should contain the properties given in example below.
2. Private key provided in sign inputs is ONLY used locally to generate signature(s) and not sent to the ledger.

Example payload file:
{
    "payload": {
        "collectionId": "<did-unique-identifier>",
        "id": "<uuid>",
        "name": "<human-readable resource name>",
        "version": "<human-readable version number>",
        "resourceType": "<resource-type>",
        "alsoKnownAs": [
            {
                "uri": "did:swtr:<unique-identifier>/resource/<uuid>",
                "description": "did-url"
            },
            {
                "uri": "https://example.com/alternative-uri",
                "description": "http-url"
            }
        ]
    },
    "signInputs": [
        {
            "verificationMethodId": "did:swtr:<unique-identifier>#<key-id>",
            "privKey": "<private-key-bytes-encoded-to-base64>"
        }
    ]
}
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read payload file arg
			payloadFile := args[0]

			// Read data file arg
			dataFile := args[1]

			payloadJSON, signInputs, err := ReadPayloadWithSignInputsFromFile(payloadFile)
			if err != nil {
				return err
			}

			// Unmarshal payload
			var payload types.MsgCreateResourcePayload
			err = clientCtx.Codec.UnmarshalJSON(payloadJSON, &payload)
			if err != nil {
				return err
			}

			// Read data file
			data, err := os.ReadFile(dataFile)
			if err != nil {
				return err
			}

			// Prepare payload
			payload.Data = data

			// Populate resource id if not set
			if payload.Id == "" {
				payload.Id = uuid.NewString()
			}

			// Build identity message
			signBytes := payload.GetSignBytes()
			identitySignatures := SignWithSignInputs(signBytes, signInputs)

			msg := types.MsgCreateResource{
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

	return cmd
}