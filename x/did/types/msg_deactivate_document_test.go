package types_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "swisstronik/x/did/types"
)

type MsgDeactivateDIDSuite struct {
	suite.Suite
}

func (suite *MsgDeactivateDIDSuite) TestMessageForDIDDeactivating() {
	type TestCaseMsgDeactivateDID struct {
		msg      *MsgDeactivateDIDDocument
		isValid  bool
		errorMsg string
	}

	testCases := []TestCaseMsgDeactivateDID{
		{
			msg: &MsgDeactivateDIDDocument{
				Payload: &MsgDeactivateDIDDocumentPayload{
					Id:        "did:swtr:zABCDEFG123456789abcd",
					VersionId: uuid.NewString(),
				},
				Signatures: nil,
			},
			isValid: true,
		},
		{
			msg: &MsgDeactivateDIDDocument{
				Payload: &MsgDeactivateDIDDocumentPayload{
					Id:        "did:swtrttt:testnet:zABCDEFG123456789abcd",
					VersionId: uuid.NewString(),
				},
				Signatures: nil,
			},
			isValid:  false,
			errorMsg: "payload: (id: did method must be: swtr.).: basic validation failed",
		},
		{
			msg: &MsgDeactivateDIDDocument{
				Payload: &MsgDeactivateDIDDocumentPayload{
					VersionId: uuid.NewString(),
				},
				Signatures: nil,
			},
			isValid:  false,
			errorMsg: "payload: (id: cannot be blank.).: basic validation failed",
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

func TestMsgDeactivateDIDSuite(t *testing.T) {
	suite.Run(t, new(MsgDeactivateDIDSuite))
}
