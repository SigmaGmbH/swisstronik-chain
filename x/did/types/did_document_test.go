package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "swisstronik/x/did/types"
)

type DIDDocValidationSuite struct {
	suite.Suite
}

func (suite *DIDDocValidationSuite) TestDIDDocValidation() {
	type DIDDocTestCase struct {
		didDoc            *DIDDocument
		allowedNamespaces []string
		isValid           bool
		errorMsg          string
	}

	testCases := []DIDDocTestCase{
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  true,
			errorMsg: "",
		},
		{
			didDoc: &DIDDocument{
				Id: InvalidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "id: unable to split did into method, namespace and id; verification_method: (0: (id: must have prefix: badDid.).).",
		},
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  true,
			errorMsg: "",
		},
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "JsonWebKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidJWK2020VerificationMaterial,
					},
				},
			},
			isValid:  true,
			errorMsg: "",
		},
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     InvalidTestDID,
						VerificationMethodType: "JsonWebKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidJWK2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "verification_method: (0: (id: unable to split did into method, namespace and id.).).",
		},
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "JsonWebKey2020",
						Controller:             InvalidTestDID,
						VerificationMaterial:   ValidJWK2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "verification_method: (0: (controller: unable to split did into method, namespace and id.).).",
		},
		{
			didDoc: &DIDDocument{
				Id:         ValidTestDID,
				Controller: []string{ValidTestDID, ValidTestDID2},
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  true,
			errorMsg: "",
		},
		{
			didDoc: &DIDDocument{
				Context:    nil,
				Id:         ValidTestDID,
				Controller: []string{ValidTestDID, InvalidTestDID},
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "controller: (1: unable to split did into method, namespace and id.).",
		},
		{
			didDoc: &DIDDocument{
				Id:         ValidTestDID,
				Controller: []string{ValidTestDID, ValidTestDID},
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "controller: there should be no duplicates.",
		},
		{
			didDoc: &DIDDocument{
				Id: ValidTestDID,
				VerificationMethod: []*VerificationMethod{
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
					{
						Id:                     fmt.Sprintf("%s#fragment", ValidTestDID),
						VerificationMethodType: "Ed25519VerificationKey2020",
						Controller:             ValidTestDID,
						VerificationMaterial:   ValidEd25519VerificationKey2020VerificationMaterial,
					},
				},
			},
			isValid:  false,
			errorMsg: "verification_method: there are verification method duplicates.",
		},
	}

	for _, testCase := range testCases {
		err := testCase.didDoc.Validate(testCase.allowedNamespaces)

		if testCase.isValid {
			assert.Nil(suite.T(), err)
		} else {
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), testCase.errorMsg)
		}
	}
}

func TestDIDDocValidationSuite(t *testing.T) {
	suite.Run(t, new(DIDDocValidationSuite))
}
