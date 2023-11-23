package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "swisstronik/x/did/types"
)

type SignInfoSuite struct {
	suite.Suite
}

func (suite *SignInfoSuite) TestSignInfoValidation() {
	type TestCaseSignInfoStruct struct {
		si                SignInfo
		allowedNamespaces []string
		isValid           bool
		errorMsg          string
	}

	testCases := []TestCaseSignInfoStruct{
		{
			si: SignInfo{
				VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
				Signature:            []byte("aaa="),
			},
			isValid:  true,
			errorMsg: "",
		},
	}

	for _, testCase := range testCases {
		err := testCase.si.Validate(testCase.allowedNamespaces)

		if testCase.isValid {
			assert.Nil(suite.T(), err)
		} else {
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), testCase.errorMsg)
		}
	}
}

func (suite *SignInfoSuite) TestSignInfoDuplicates() {
	type TestCaseSignInfosStruct struct {
		signInfos []*SignInfo
		isValid   bool
	}

	testCases := []TestCaseSignInfosStruct{
		{
			signInfos: []*SignInfo{
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("bbb="),
				},
			},
			isValid: true,
		},
		{
			signInfos: []*SignInfo{
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("bbb="),
				},
			},
			isValid: true,
		},
		{
			signInfos: []*SignInfo{
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
			},
			isValid: false,
		},
		{
			signInfos: []*SignInfo{
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
				{
					VerificationMethodId: "did:swtr:zABCDEFG123456789abcd#method1",
					Signature:            []byte("aaa="),
				},
			},
			isValid: false,
		},
	}

	for _, testCase := range testCases {
		res := IsUniqueSignInfoList(testCase.signInfos)
		assert.Equal(suite.T(), testCase.isValid, res)
	}
}

func TestSignInfoSuite(t *testing.T) {
	suite.Run(t, new(SignInfoSuite))
}
