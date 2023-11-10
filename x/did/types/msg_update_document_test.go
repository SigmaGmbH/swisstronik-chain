package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "swisstronik/x/did/types"
)

type MsgUpdateDIDSuite struct {
	suite.Suite
}

func (suite *MsgUpdateDIDSuite) TestMessageForDIDUpdating() {
	type TestCaseMsgUpdateDID struct {
		msg      *MsgUpdateDIDDocument
		isValid  bool
		errorMsg string
	}

	testCases := []TestCaseMsgUpdateDID{
		{
			msg: &MsgUpdateDIDDocument{
				Payload: &MsgUpdateDIDDocumentPayload{
					Id: "did:swtr:zABCDEFG123456789abcd",
					VerificationMethod: []*VerificationMethod{
						{
							Id:                     "did:swtr:zABCDEFG123456789abcd#key1",
							VerificationMethodType: "Ed25519VerificationKey2020",
							Controller:             "did:swtr:zABCDEFG123456789abcd",
							VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
						},
					},
					Authentication: []string{"did:swtr:zABCDEFG123456789abcd#key1", "did:swtr:zABCDEFG123456789abcd#aaa"},
					VersionId:      "version1",
				},
				Signatures: nil,
			},
			isValid: true,
		},
		{
			msg: &MsgUpdateDIDDocument{
				Payload: &MsgUpdateDIDDocumentPayload{
					Id: "did:swtr:zABCDEFG123456789abcd",
					VerificationMethod: []*VerificationMethod{
						{
							Id:                     "did:swtr:zABCDEFG123456789abcd#key1",
							VerificationMethodType: "Ed25519VerificationKey2020",
							Controller:             "did:swtr:zABCDEFG123456789abcd",
							VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
						},
					},
					Authentication: []string{"did:swtr:zABCDEFG123456789abcd#key1", "did:swtr:zABCDEFG123456789abcd#key1"},
					VersionId:      "version1",
				},
			},
			isValid:  false,
			errorMsg: "payload: (authentication: there should be no duplicates.).: basic validation failed",
		},
		{
			msg: &MsgUpdateDIDDocument{
				Payload: &MsgUpdateDIDDocumentPayload{
					Id: "did:swtr:zABCDEFG123456789abcd",
					VerificationMethod: []*VerificationMethod{
						{
							Id:                     "did:swtr:zABCDEFG123456789abcd#key1",
							VerificationMethodType: "Ed25519VerificationKey2020",
							Controller:             "did:swtr:zABCDEFG123456789abcd",
							VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
						},
					},
					Authentication: []string{"did:swtr:zABCDEFG123456789abcd#key1", "did:swtr:zABCDEFG123456789abcd#aaa"},
				},
			},
			isValid:  false,
			errorMsg: "payload: (version_id: cannot be blank.).: basic validation failed",
		},
	}

	for _, testCase := range testCases {
		err := testCase.msg.ValidateBasic()

		if testCase.isValid {
			assert.Nil(suite.T(), err)
		} else {
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), testCase.errorMsg)
		}
	}
}

func TestMsgUpdateDIDSuite(t *testing.T) {
	suite.Run(t, new(MsgUpdateDIDSuite))
}
