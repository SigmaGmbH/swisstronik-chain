package types_test

import (
	"fmt"

	. "swisstronik/x/did/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type DIDDocTestCase struct {
	didDoc            *DIDDocument
	allowedNamespaces []string
	isValid           bool
	errorMsg          string
}

var _ = DescribeTable("DIDDoc Validation tests", func(testCase DIDDocTestCase) {
	err := testCase.didDoc.Validate(testCase.allowedNamespaces)

	if testCase.isValid {
		Expect(err).To(BeNil())
	} else {
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(testCase.errorMsg))
	}
},

	Entry(
		"DIDDoc is valid",
		DIDDocTestCase{
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
		}),

	Entry(
		"DIDDoc is invalid",
		DIDDocTestCase{
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
		}),

	Entry(
		"Verification method is Ed25519VerificationKey2020",
		DIDDocTestCase{
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
		}),

	Entry(
		"Verification method is JWK",
		DIDDocTestCase{
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
		}),

	Entry("Verification method has wrong ID",
		DIDDocTestCase{
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
		}),
	Entry(
		"Verification method has wrong controller",
		DIDDocTestCase{
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
		}),
	Entry(
		"List of DIDs in controller is allowed",
		DIDDocTestCase{
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
		}),
	Entry(
		"List of DIDs in controller is not allowed",
		DIDDocTestCase{
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
		}),
	Entry(
		"Controller is duplicated",
		DIDDocTestCase{
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
		}),
	Entry(
		"Verification method is duplicated",
		DIDDocTestCase{
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
		}),
)
