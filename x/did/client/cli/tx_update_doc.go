package cli

import (
	"encoding/json"

	"swisstronik/x/did/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func CmdUpdateDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [payload-file] --version-id [version-id]",
		Short: "Updates a DID and its associated DID Document.",
		Long: `Update DID Document associated with a given DID. 
[payload-file] is JSON encoded DID Document alongside with sign inputs. 
Version ID is optional and is determined by the '--version-id' flag. 
If not provided, a random UUID will be used as version-id.

NOTES:
2. DID update operations require the FULL new DID Document to be provided. Specifying just the changes/diff is not supported.
3. Payload file should be a JSON file containing properties specified in the DID Core Specification. Rules from DID Core spec are followed on which properties are mandatory and which ones are optional.
4. Private key provided in sign inputs is ONLY used locally to generate signature(s) and not sent to the ledger.

Example payload file:
{
    "payload": {
        "context": [ "https://www.w3.org/ns/did/v1" ],
        "id": "did:swtr:<unique-identifier>",
        "controller": [
            "did:swtr:<unique-identifier>"
        ],
        "authentication": [
            "did:swtr:<unique-identifier>#<key-id>"
        ],
        "assertionMethod": [],
        "capabilityInvocation": [],
        "capabilityDelegation": [],
        "keyAgreement": [],
        "alsoKnownAs": [],
        "verificationMethod": [
            {
                "id": "did:swtr:<unique-identifier>#<key-id>",
                "type": "<verification-method-type>",
                "controller": "did:swtr:<unique-identifier>",
                "publicKeyMultibase": "<public-key>"
            }
        ],
        "service": [
			{
                "id": "did:swtr:<unique-identifier>#<service-id>",
                "type": "<service-type>",
                "serviceEndpoint": [
                    "<service-endpoint>"
                ]
            }
		]
    },
	"signInputs": [
        {
            "verificationMethodId": "did:swtr:<unique-identifier>#<key-id>",
            "privKey": "<private-key-bytes-encoded-to-base64>"
        }
    ]
}`,
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

			// Unmarshal spec-compliant payload
			var specPayload DIDDocument
			err = json.Unmarshal([]byte(payloadJSON), &specPayload)
			if err != nil {
				return err
			}

			// Validate spec-compliant payload & get verification methods
			verificationMethod, service, err := GetFromSpecCompliantPayload(specPayload)
			if err != nil {
				return err
			}

			// Construct MsgUpdateDIDDocumentPayload
			payload := types.MsgUpdateDIDDocumentPayload{
				Context:              specPayload.Context,
				Id:                   specPayload.ID,
				Controller:           specPayload.Controller,
				VerificationMethod:   verificationMethod,
				Authentication:       specPayload.Authentication,
				AssertionMethod:      specPayload.AssertionMethod,
				CapabilityInvocation: specPayload.CapabilityInvocation,
				CapabilityDelegation: specPayload.CapabilityDelegation,
				KeyAgreement:         specPayload.KeyAgreement,
				Service:              service,
				AlsoKnownAs:          specPayload.AlsoKnownAs,
				VersionId:            versionID, // Set version id, from flag or random
			}

			// Build identity message
			signBytes := payload.GetSignBytes()
			identitySignatures := SignWithSignInputs(signBytes, signInputs)

			msg := types.MsgUpdateDIDDocument{
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