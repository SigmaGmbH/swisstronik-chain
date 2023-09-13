package types_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "swisstronik/x/did/types"
)

var _ = Describe("Message for DID updating", func() {
	type TestCaseMsgDeactivateDID struct {
		msg      *MsgDeactivateDIDDocument
		isValid  bool
		errorMsg string
	}

	DescribeTable("Tests for message for DID deactivating", func(testCase TestCaseMsgDeactivateDID) {
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
			TestCaseMsgDeactivateDID{
				msg: &MsgDeactivateDIDDocument{
					Payload: &MsgDeactivateDIDDocumentPayload{
						Id:        "did:swtr:zABCDEFG123456789abcd",
						VersionId: uuid.NewString(),
					},
					Signatures: nil,
				},
				isValid: true,
			}),

		Entry(
			"Negative: Invalid DID Method",
			TestCaseMsgDeactivateDID{
				msg: &MsgDeactivateDIDDocument{
					Payload: &MsgDeactivateDIDDocumentPayload{
						Id:        "did:swtrttt:testnet:zABCDEFG123456789abcd",
						VersionId: uuid.NewString(),
					},
					Signatures: nil,
				},
				isValid:  false,
				errorMsg: "payload: (id: did method must be: swtr.).: basic validation failed",
			}),

		Entry(
			"Negative: Id is required",
			TestCaseMsgDeactivateDID{
				msg: &MsgDeactivateDIDDocument{
					Payload: &MsgDeactivateDIDDocumentPayload{
						VersionId: uuid.NewString(),
					},
					Signatures: nil,
				},
				isValid:  false,
				errorMsg: "payload: (id: cannot be blank.).: basic validation failed",
			}),
	)
})
