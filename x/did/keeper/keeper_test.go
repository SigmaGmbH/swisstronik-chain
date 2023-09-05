package keeper_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"swisstronik/simapp"
	didutil "swisstronik/testutil/did"
	"swisstronik/x/did/keeper"
	"swisstronik/x/did/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	goCtx  context.Context
	keeper keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	s.Setup(t)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) Setup(t *testing.T) {
	app, _ := simapp.Setup(t, false)
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{ChainID: "swisstronik_1291-1"})
	s.goCtx = sdk.WrapSDKContext(s.ctx)
	s.keeper = app.DIDKeeper
}

func (suite *KeeperTestSuite) TestCreateSimpleDIDEd25519VerificationKey2020() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  keypair.Private,
		},
	}

	// create DID document
	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *resp.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCreateSimpleDIDJsonWebKey2020() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.JSONWebKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateJSONWebKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  keypair.Private,
		},
	}

	// Create DID document
	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// Check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *resp.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCreateSimpleDIDEd25519VerificationKey2018() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2018Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2018VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  keypair.Private,
		},
	}

	// Create DID document
	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// Check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *resp.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCreateDIDWithExternalControllers() {
	// Controllers
	firstController, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	secondController, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	// DID
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{firstController.Did, secondController.Did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             secondController.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{firstController.SignInput, secondController.SignInput}

	_, err = didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	created, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCreateDIDWithAllProperties() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)

	keypair1 := didutil.GenerateKeyPair()
	keyID1 := did + "#key-1"

	keypair2 := didutil.GenerateKeyPair()
	keyID2 := did + "#key-2"

	keypair3 := didutil.GenerateKeyPair()
	keyID3 := did + "#key-3"

	keypair4 := didutil.GenerateKeyPair()
	keyID4 := did + "#key-4"

	payload := &types.MsgCreateDIDDocumentPayload{
		Context:    []string{"abc", "def"},
		Id:         did,
		Controller: []string{did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID1,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair1.Public),
			},
			{
				Id:                     keyID2,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair2.Public),
			},
			{
				Id:                     keyID3,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair3.Public),
			},
			{
				Id:                     keyID4,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair4.Public),
			},
		},
		Authentication:       []string{keyID1, keyID2},
		AssertionMethod:      []string{keyID3},
		CapabilityInvocation: []string{keyID4, keyID1},
		CapabilityDelegation: []string{keyID4, keyID2},
		KeyAgreement:         []string{keyID1, keyID2, keyID3, keyID4},
		Service: []*types.Service{
			{
				Id:              did + "#service-1",
				ServiceType:     "type-1",
				ServiceEndpoint: []string{"endpoint-1"},
			},
		},
		AlsoKnownAs: []string{"alias-1", "alias-2"},
		VersionId:   uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID1,
			Key:                  keypair1.Private,
		},
	}

	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	created, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestShouldFailWithoutSignatureOfSecondController() {
	// Controller
	controller, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	// DID
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{controller.Did, did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  keypair.Private,
		},
	}

	_, err = didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("signer: %s: signature is required but not found", controller.Did))
}

func (suite *KeeperTestSuite) TestShouldFailWithoutSignatures() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{}

	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("signer: %s: signature is required but not found", did))
}

func (suite *KeeperTestSuite) TestShouldFailIfControllerWasNotFound() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	nonExistingDid := didutil.GenerateDID(didutil.Base58_16bytes)

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{nonExistingDid},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  keypair.Private,
		},
	}

	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("%s: DID Document not found", nonExistingDid))
}

func (suite *KeeperTestSuite) TestShouldFailIfWrongSignature() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	invalidKey := didutil.GenerateKeyPair()

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: keyID,
			Key:                  invalidKey.Private,
		},
	}

	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("method id: %s: invalid signature detected", keyID))
}

func (suite *KeeperTestSuite) TestShouldFailIfSignedByWrongController() {
	controller, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{controller.SignInput}

	_, err = didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("signer: %s: signature is required but not found", did))
}

func (suite *KeeperTestSuite) TestShouldFailIfSignedByWrongMethod() {
	did := didutil.GenerateDID(didutil.Base58_16bytes)
	keypair := didutil.GenerateKeyPair()
	keyID := did + "#key-1"

	payload := &types.MsgCreateDIDDocumentPayload{
		Id:             did,
		Controller:     []string{did},
		Authentication: []string{keyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     keyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	invalidKeyID := did + "#key-2"

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: invalidKeyID,
			Key:                  keypair.Private,
		},
	}

	_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("%s: verification method not found", invalidKeyID))
}

func (suite *KeeperTestSuite) TestCannotCreateIfDIDAlreadyExists() {
	did, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	msg := &types.MsgCreateDIDDocumentPayload{
		Id:             did.Did,
		Authentication: []string{did.KeyID},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(did.KeyPair.Public),
			},
		},
	}

	signatures := []didutil.SignInput{did.SignInput}

	_, err = didutil.CreateDID(suite.ctx, suite.keeper, msg, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("%s: DID Document exists", did.Did))
}

func (suite *KeeperTestSuite) TestShouldCreateWithMixedCases() {
	didPrefix := "did:swtr:"

	testCases := []struct {
		name string
		input string
		result string 
	} {
		{
			name: "lowercase UUID",
			input: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name: "Uppercase UUID",
			input:  didPrefix + "A86F9CAE-0902-4A7C-A144-96B60CED2FC9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name: "Mixed case UUID",
			input:  didPrefix + "A86F9CAE-0902-4a7c-a144-96b60ced2FC9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name: "Indy-style ID",
			input:  didPrefix + "zABCDEFG123456789abcd",
			result: didPrefix + "zABCDEFG123456789abcd",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Setup(suite.T())
			did := tc.input
			keypair := didutil.GenerateKeyPair()
			keyID := did + "#key-1"

			payload := &types.MsgCreateDIDDocumentPayload{
				Id:             did,
				Authentication: []string{keyID},
				VerificationMethod: []*types.VerificationMethod{
					{
						Id:                     keyID,
						VerificationMethodType: types.Ed25519VerificationKey2020Type,
						Controller:             did,
						VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
					},
				},
				VersionId: uuid.NewString(),
			}
	
			signatures := []didutil.SignInput{
				{
					VerificationMethodID: keyID,
					Key:                  keypair.Private,
				},
			}

			_, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
			suite.Require().NoError(err)

			resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
			suite.Require().NoError(err)
			suite.Require().Equal(resp.Value.DidDoc.Id, tc.result)
		})
	}
}
