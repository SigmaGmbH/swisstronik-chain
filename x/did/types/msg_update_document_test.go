package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "swisstronik/x/did/types"
)

var _ = Describe("Message for DID updating", func() {
	type TestCaseMsgUpdateDID struct {
		msg      *MsgUpdateDIDDocument
		isValid  bool
		errorMsg string
	}

	DescribeTable("Tests for message for DID updating", func(testCase TestCaseMsgUpdateDID) {
		err := testCase.msg.ValidateBasic()

		if testCase.isValid {
			Expect(err).To(BeNil())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testCase.errorMsg))
		}
	},

		Entry(
			"All fields are set properly",
			TestCaseMsgUpdateDID{
				msg: &MsgUpdateDIDDocument{
					Payload: &MsgUpdateDIDDocumentPayload{
						Id: "did:swtr:testnet:zABCDEFG123456789abcd",
						VerificationMethod: []*VerificationMethod{
							{
								Id:                     "did:swtr:testnet:zABCDEFG123456789abcd#key1",
								VerificationMethodType: "Ed25519VerificationKey2020",
								Controller:             "did:swtr:testnet:zABCDEFG123456789abcd",
								VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
							},
						},
						Authentication: []string{"did:swtr:testnet:zABCDEFG123456789abcd#key1", "did:swtr:testnet:zABCDEFG123456789abcd#aaa"},
						VersionId:      "version1",
					},
					Signatures: nil,
				},
				isValid: true,
			}),

		Entry(
			"IDs are duplicated",
			TestCaseMsgUpdateDID{
				msg: &MsgUpdateDIDDocument{
					Payload: &MsgUpdateDIDDocumentPayload{
						Id: "did:swtr:testnet:zABCDEFG123456789abcd",
						VerificationMethod: []*VerificationMethod{
							{
								Id:                     "did:swtr:testnet:zABCDEFG123456789abcd#key1",
								VerificationMethodType: "Ed25519VerificationKey2020",
								Controller:             "did:swtr:testnet:zABCDEFG123456789abcd",
								VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
							},
						},
						Authentication: []string{"did:swtr:testnet:zABCDEFG123456789abcd#key1", "did:swtr:testnet:zABCDEFG123456789abcd#key1"},
						VersionId:      "version1",
					},
					Signatures: nil,
				},
				isValid:  false,
				errorMsg: "payload: (authentication: there should be no duplicates.).: basic validation failed",
			}),
		Entry(
			"VersionId is empty",
			TestCaseMsgUpdateDID{
				msg: &MsgUpdateDIDDocument{
					Payload: &MsgUpdateDIDDocumentPayload{
						Id: "did:swtr:testnet:zABCDEFG123456789abcd",
						VerificationMethod: []*VerificationMethod{
							{
								Id:                     "did:swtr:testnet:zABCDEFG123456789abcd#key1",
								VerificationMethodType: "Ed25519VerificationKey2020",
								Controller:             "did:swtr:testnet:zABCDEFG123456789abcd",
								VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
							},
						},
						Authentication: []string{"did:swtr:testnet:zABCDEFG123456789abcd#key1", "did:swtr:testnet:zABCDEFG123456789abcd#aaa"},
					},
					Signatures: nil,
				},
				isValid:  false,
				errorMsg: "payload: (version_id: cannot be blank.).: basic validation failed",
			}),
	)
})
