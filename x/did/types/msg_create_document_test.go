package types_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "swisstronik/x/did/types"
)

type MsgCreateDIDSuite struct {
	suite.Suite
}

func (suite *MsgCreateDIDSuite) TestMessageForDIDCreation() {
	type TestCaseMsgCreateDID struct {
		msg      *MsgCreateDIDDocument
		isValid  bool
		errorMsg string
	}

	testCases := []TestCaseMsgCreateDID{
		{
			msg: &MsgCreateDIDDocument{
				Payload: &MsgCreateDIDDocumentPayload{
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
					VersionId:      uuid.NewString(),
				},
				Signatures: nil,
			},
			isValid: true,
		},
		{
			msg: &MsgCreateDIDDocument{
				Payload: &MsgCreateDIDDocumentPayload{
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
				},
				Signatures: nil,
			},
			isValid:  false,
			errorMsg: "payload: (authentication: there should be no duplicates.).: basic validation failed",
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

func TestMsgCreateDIDSuite(t *testing.T) {
	suite.Run(t, new(MsgCreateDIDSuite))
}
