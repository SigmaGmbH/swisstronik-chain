package types_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "swisstronik/x/did/types"
)

var _ = Describe("Message for DID creation", func() {
	type TestCaseMsgCreateDID struct {
		msg      *MsgCreateDIDDocument
		isValid  bool
		errorMsg string
	}

	DescribeTable("Tests for message for DID creation", func(testCase TestCaseMsgCreateDID) {
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
			TestCaseMsgCreateDID{
				msg: &MsgCreateDIDDocument{
					Payload: &MsgCreateDIDDocumentPayload{
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
						VersionId:      uuid.NewString(),
					},
					Signatures: nil,
				},
				isValid: true,
			},
		),
		Entry(
			"IDs are duplicated",
			TestCaseMsgCreateDID{
				msg: &MsgCreateDIDDocument{
					Payload: &MsgCreateDIDDocumentPayload{
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
					},
					Signatures: nil,
				},
				isValid:  false,
				errorMsg: "payload: (authentication: there should be no duplicates.).: basic validation failed",
			},
		),
	)
})
