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

	"crypto/sha256"
	"encoding/hex"
	"swisstronik/app"
	didutil "swisstronik/testutil/did"
	"swisstronik/utils"
	"swisstronik/x/did/keeper"
	"swisstronik/x/did/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	goCtx  context.Context
	keeper keeper.Keeper
	app    *app.App
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
	checkTx := false
	chainID := utils.TestnetChainID + "-1"

	app, _ := app.SetupSwissApp(checkTx, nil, chainID)
	s.ctx = app.BaseApp.NewContext(false)
	s.goCtx = s.ctx
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

	// Check if verification methods were indexed
	for _, vm := range payload.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		suite.Require().Contains(controlledDocuments, payload.Id)
		suite.Require().Equal(1, len(controlledDocuments))
	}
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
		name   string
		input  string
		result string
	}{
		{
			name:   "lowercase UUID",
			input:  didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name:   "Uppercase UUID",
			input:  didPrefix + "A86F9CAE-0902-4A7C-A144-96B60CED2FC9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name:   "Mixed case UUID",
			input:  didPrefix + "A86F9CAE-0902-4a7c-a144-96b60ced2FC9",
			result: didPrefix + "a86f9cae-0902-4a7c-a144-96b60ced2fc9",
		},
		{
			name:   "Indy-style ID",
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

func (suite *KeeperTestSuite) TestShouldDeactivateDID() {
	did, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	// Check if verification methods were indexed
	existingDID, err := suite.keeper.GetLatestDIDDocument(suite.ctx, did.Did)
	suite.Require().NoError(err)
	for _, vm := range existingDID.DidDoc.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		suite.Require().Contains(controlledDocuments, did.Did)
		suite.Require().Equal(1, len(controlledDocuments))
	}

	payload := &types.MsgDeactivateDIDDocumentPayload{
		Id:        did.Did,
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{did.DIDDocumentInfo.SignInput}

	res, err := didutil.DeactivateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)
	suite.Require().True(res.Value.Metadata.Deactivated)

	// Check that all versions are deactivated
	versions, err := suite.keeper.AllDIDDocumentVersionsMetadata(suite.goCtx, &types.QueryAllDIDDocumentVersionsMetadataRequest{Id: did.Did})
	suite.Require().NoError(err)

	for _, version := range versions.Versions {
		suite.Require().True(version.Deactivated)
	}

	// Check that deactivated document was removed from controlled list
	for _, vm := range existingDID.DidDoc.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		suite.Require().NotContains(controlledDocuments, did.Did)
		suite.Require().Equal(0, len(controlledDocuments))
	}
}

func (suite *KeeperTestSuite) TestUpdateCreatesNewVersion() {
	suite.Setup(suite.T())

	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	second, err := didutil.CreateDIDDocumentWithExternalControllers(
		suite.ctx,
		suite.keeper,
		[]string{first.Did}, []didutil.SignInput{first.SignInput},
	)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         second.Did,
		Controller: []string{first.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     second.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             second.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(second.KeyPair.Public),
			},
		},
		Authentication:  []string{second.KeyID},
		AssertionMethod: []string{second.KeyID},
		VersionId:       uuid.NewString(),
	}

	signatures := []didutil.SignInput{first.SignInput}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// check latest version
	created, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, second.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)

	// query the first version
	v1, err := didutil.GetDIDDocumentVersion(suite.ctx, suite.keeper, second.Did, second.VersionID)
	suite.Require().NoError(err)
	suite.Require().Equal(second.Msg.ToDidDoc(), *v1.Value.DidDoc)
	suite.Require().Equal(second.VersionID, v1.Value.Metadata.VersionId)
	suite.Require().Equal(payload.VersionId, v1.Value.Metadata.NextVersionId)

	// query the second version
	v2, err := didutil.GetDIDDocumentVersion(suite.ctx, suite.keeper, second.Did, payload.VersionId)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *v2.Value.DidDoc)
	suite.Require().Equal(payload.VersionId, v2.Value.Metadata.VersionId)
	suite.Require().Equal(second.VersionID, v2.Value.Metadata.PreviousVersionId)

	// query all versions
	versions, err := didutil.GetAllDIDVersionsMetadata(suite.ctx, suite.keeper, second.Did)
	suite.Require().NoError(err)
	suite.Require().Len(versions.Versions, 2)
	suite.Require().Contains(versions.Versions, v1.Value.Metadata)
	suite.Require().Contains(versions.Versions, v2.Value.Metadata)
}

func (suite *KeeperTestSuite) TestCannotUpdateDocumentWithoutControllerSignature() {
	suite.Setup(suite.T())

	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	second, err := didutil.CreateDIDDocumentWithExternalControllers(
		suite.ctx,
		suite.keeper,
		[]string{first.Did}, []didutil.SignInput{first.SignInput},
	)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         second.Did,
		Controller: []string{first.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     second.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             second.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(second.KeyPair.Public),
			},
		},
		Authentication:  []string{second.KeyID},
		AssertionMethod: []string{second.KeyID},
		VersionId:       uuid.NewString(),
	}

	signatures := []didutil.SignInput{}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s: signature is required but not found", first.Did))
}

func (suite *KeeperTestSuite) TestReplaceControllerWithBothSignatures() {
	suite.Setup(suite.T())

	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	second, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         first.Did,
		Controller: []string{second.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     first.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             first.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(first.KeyPair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		first.SignInput,
		second.SignInput,
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	updated, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, first.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *updated.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCannotReplaceControllerWithOnlyOneSignature() {
	suite.Setup(suite.T())

	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	second, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         first.Did,
		Controller: []string{second.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     first.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             first.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(first.KeyPair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	onlyNewControllerSignatures := []didutil.SignInput{
		second.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyNewControllerSignatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s (previous version): signature is required but not found", first.Did))

	onlyPreviousControllerSignatures := []didutil.SignInput{
		first.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyPreviousControllerSignatures)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s: signature is required but not found", second.Did))
}

func (suite *KeeperTestSuite) TestAddControllerWithBothSignatures() {
	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	second, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         first.Did,
		Controller: []string{first.Did, second.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     first.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             first.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(first.KeyPair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		first.SignInput,
		second.SignInput,
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	updated, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, first.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *updated.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCannotAddControllerWithOnlyOneSignature() {
	first, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	second, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         first.Did,
		Controller: []string{first.Did, second.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     first.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             first.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(first.KeyPair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	onlyExistingControllerSignature := []didutil.SignInput{
		first.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyExistingControllerSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s: signature is required but not found", second.Did))

	onlyNewControllerSignature := []didutil.SignInput{
		second.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyNewControllerSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s (previous version): signature is required but not found", first.Did))
}

func (suite *KeeperTestSuite) TestUpdateWithSameVerificationMethod() {
	controller, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	subject, err := didutil.CreateDIDDocumentWithExternalControllers(suite.ctx, suite.keeper, []string{controller.Did}, []didutil.SignInput{controller.SignInput})
	suite.Require().NoError(err)

	// Check if verification methods were indexed
	subjectDocument, err := suite.keeper.GetLatestDIDDocument(suite.ctx, subject.Did)
	suite.Require().NoError(err)
	for _, vm := range subjectDocument.DidDoc.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		suite.Require().Contains(controlledDocuments, subject.Did)
		suite.Require().Equal(1, len(controlledDocuments))
	}

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id:         subject.Did,
		Controller: []string{controller.Did},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
		},
		Authentication:  []string{subject.KeyID},
		AssertionMethod: []string{subject.KeyID},
		VersionId:       uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		controller.SignInput,
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// Check if verification methods still the same
	suite.Require().NoError(err)
	for _, vm := range subjectDocument.DidDoc.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		suite.Require().Contains(controlledDocuments, subject.Did)
		suite.Require().Equal(1, len(controlledDocuments))
	}

	created, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, subject.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestUpdateKeyForVerificationMethod() {
	did, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	newKeyPair := didutil.GenerateKeyPair()
	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: did.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             did.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(newKeyPair.Public),
			},
		},
		VersionId: uuid.NewString(),
	}

	onlyNewSignature := []didutil.SignInput{did.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyNewSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (new version): invalid signature detected", did.Did))

	onlyPreviousSignature := []didutil.SignInput{{
		VerificationMethodID: did.KeyID,
		Key:                  newKeyPair.Private,
	}}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyPreviousSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (previous version): invalid signature detected", did.Did))

	correctSignatures := []didutil.SignInput{
		did.SignInput, // Old signature
		{
			VerificationMethodID: did.KeyID, // New signature
			Key:                  newKeyPair.Private,
		},
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, correctSignatures)
	suite.Require().NoError(err)

	created, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, did.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(*created.Value.DidDoc, payload.ToDidDoc())
}

func (suite *KeeperTestSuite) TestUpdateController() {
	subject, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	controller, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             controller.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	onlyPreviousControllerSignature := []didutil.SignInput{controller.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyPreviousControllerSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s (previous version): signature is required but not found", subject.Did))

	onlyNewControllerSignature := []didutil.SignInput{subject.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, onlyNewControllerSignature)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one signature by %s: signature is required but not found", controller.Did))

	correctSignatures := []didutil.SignInput{subject.SignInput, controller.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, correctSignatures)
	suite.Require().NoError(err)

	updated, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, subject.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *updated.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestUpdateVerificationMethodID() {
	subject, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	newKeyID := subject.Did + "#key-2"

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     newKeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	signatureWithoutNewVerificationMethod := []didutil.SignInput{subject.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatureWithoutNewVerificationMethod)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (new version): invalid signature detected", subject.Did))

	signatureWithoutPreviousVerificationMethod := []didutil.SignInput{
		{
			VerificationMethodID: newKeyID,
			Key:                  subject.KeyPair.Private,
		},
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatureWithoutPreviousVerificationMethod)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (previous version): invalid signature detected", subject.Did))

	correctSignatures := []didutil.SignInput{
		{
			VerificationMethodID: newKeyID,
			Key:                  subject.KeyPair.Private,
		},
		subject.SignInput,
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, correctSignatures)
	suite.Require().NoError(err)

	updated, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, subject.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *updated.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestUpdateAddNewVerificationMethod() {
	subject, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	newKeyID := subject.Did + "#key-2"
	newKey := didutil.GenerateKeyPair()

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
			{
				Id:                     newKeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(newKey.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	signatureWithOnlyNewVerificationMethod := []didutil.SignInput{
		{
			VerificationMethodID: newKeyID,
			Key:                  newKey.Private,
		},
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatureWithOnlyNewVerificationMethod)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (previous version): invalid signature detected", subject.Did))

	correctSignature := []didutil.SignInput{
		subject.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, correctSignature)
	suite.Require().NoError(err)

	created, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, subject.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestRemoveVerificationMethod() {
	subject, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	secondKeyID := subject.Did + "#key-2"
	secondKey := didutil.GenerateKeyPair()
	secondSignInput := didutil.SignInput{
		VerificationMethodID: secondKeyID,
		Key:                  secondKey.Private,
	}

	addSecondKeyPayload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
			{
				Id:                     secondKeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(secondKey.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, addSecondKeyPayload, []didutil.SignInput{subject.SignInput})
	suite.Require().NoError(err)

	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.KeyPair.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	signatureWithRemovedMethod := []didutil.SignInput{secondSignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatureWithRemovedMethod)
	suite.Require().ErrorContains(err, fmt.Sprintf("there should be at least one valid signature by %s (new version): invalid signature detected", subject.Did))

	correctSignature := []didutil.SignInput{
		subject.SignInput,
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, correctSignature)
	suite.Require().NoError(err)

	created, err := didutil.GetDIDDocument(suite.ctx, suite.keeper, subject.Did)
	suite.Require().NoError(err)
	suite.Require().Equal(payload.ToDidDoc(), *created.Value.DidDoc)
}

func (suite *KeeperTestSuite) TestCannotUpdateDeactivatedDID() {
	subject, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	updatePayload := &types.MsgUpdateDIDDocumentPayload{
		Id: subject.Did,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     subject.DIDDocumentInfo.KeyID,
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             subject.DIDDocumentInfo.Did,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(subject.DIDDocumentInfo.KeyPair.Public),
			},
		},
		Authentication: []string{subject.KeyID},
		VersionId:      uuid.NewString(),
	}

	// Deactivate DID
	deactivatePayload := &types.MsgDeactivateDIDDocumentPayload{
		Id:        subject.Did,
		VersionId: uuid.NewString(),
	}
	signatures := []didutil.SignInput{subject.DIDDocumentInfo.SignInput}
	res, err := didutil.DeactivateDIDDocument(suite.ctx, suite.keeper, deactivatePayload, signatures)
	suite.Require().NoError(err)
	suite.Require().True(res.Value.Metadata.Deactivated)

	// Update deactivated DID
	signatures = []didutil.SignInput{subject.SignInput}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, updatePayload, signatures)
	suite.Require().ErrorContains(err, subject.DIDDocumentInfo.Did+": DID Document already deactivated")
}

func (suite *KeeperTestSuite) ExpectPayloadToMatchResource(payload *types.MsgCreateResourcePayload, resource *types.ResourceWithMetadata) {
	// Provided header
	suite.Require().Equal(resource.Metadata.Id, payload.Id)
	suite.Require().Equal(resource.Metadata.CollectionId, payload.CollectionId)
	suite.Require().Equal(resource.Metadata.Name, payload.Name)
	suite.Require().Equal(resource.Metadata.ResourceType, payload.ResourceType)

	defaultAlternativeURL := types.AlternativeUri{
		Uri:         "did:swtr:" + payload.CollectionId + "/resources/" + payload.Id,
		Description: "did-url",
	}

	suite.Require().Equal(resource.Metadata.AlsoKnownAs, append(payload.AlsoKnownAs, &defaultAlternativeURL))

	// Generated header
	hash := sha256.Sum256(payload.Data)
	hex := hex.EncodeToString(hash[:])
	suite.Require().Equal(hex, resource.Metadata.Checksum)

	// Provided data
	suite.Require().Equal(resource.Resource.Data, payload.Data)
}

func (suite *KeeperTestSuite) TestCreateWithControllerSignature() {
	did, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	payload := &types.MsgCreateResourcePayload{
		CollectionId: did.CollectionID,
		Id:           uuid.NewString(),
		Name:         "Test Resource Name",
		ResourceType: didutil.CLSchemaType,
		Data:         []byte(didutil.SchemaData),
		AlsoKnownAs: []*types.AlternativeUri{
			{
				Uri: "https://example.com/alternative-uri",
			},
			{
				Uri:         "https://example.com/alternative-uri",
				Description: "Alternative URI description",
			},
		},
	}

	_, err = didutil.CreateResource(suite.ctx, suite.keeper, payload, []didutil.SignInput{did.SignInput})
	suite.Require().NoError(err)

	// check
	created, err := suite.keeper.Resource(sdk.WrapSDKContext(suite.ctx), &types.QueryResourceRequest{
		CollectionId: did.CollectionID,
		Id:           payload.Id,
	})
	suite.Require().NoError(err)

	suite.ExpectPayloadToMatchResource(payload, created.Resource)
}

func (suite *KeeperTestSuite) TestIndexVerificationMethodAfterAddingDocument() {
	did, err := didutil.CreateDefaultDID(suite.ctx, suite.keeper)
	suite.Require().NoError(err)

	createdDocument, err := suite.keeper.GetLatestDIDDocument(suite.ctx, did.Did)
	suite.Require().NoError(err)

	// Check if verification method was indexed correctly
	for _, vm := range createdDocument.DidDoc.VerificationMethod {
		controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, vm.VerificationMaterial)
		suite.Require().NoError(err)
		// There should be only one associated document
		suite.Require().Equal(1, len(controlledDocuments))
		// Should contain DID URL of created document
		suite.Require().Contains(controlledDocuments, did.Did)
	}
}

func (suite *KeeperTestSuite) TestIndexMultipleDocuments() {
	documentsAmount := 3
	keypair := didutil.GenerateKeyPair()

	// Generate 3 did documents with same verification method
	var createdDocuments []*types.DIDDocument
	for i := 0; i < documentsAmount; i++ {
		did := didutil.GenerateDID(didutil.Base58_16bytes)
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
		createdDoc, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
		suite.Require().NoError(err)

		createdDocuments = append(createdDocuments, createdDoc.Value.DidDoc)
	}

	// Get verification method and obtain DID URLs
	verificationMaterial := createdDocuments[0].VerificationMethod[0].VerificationMaterial
	controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, verificationMaterial)
	suite.Require().NoError(err)

	suite.Require().Equal(documentsAmount, len(controlledDocuments))
	for _, doc := range createdDocuments {
		suite.Require().Contains(controlledDocuments, doc.Id)
	}
}

func (suite *KeeperTestSuite) TestDeactivateOneDocument() {
	documentsAmount := 2
	keypair := didutil.GenerateKeyPair()

	// Generate 2 did documents with same verification method
	var createdDocuments []*types.DIDDocument
	for i := 0; i < documentsAmount; i++ {
		did := didutil.GenerateDID(didutil.Base58_16bytes)
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
		createdDoc, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
		suite.Require().NoError(err)

		createdDocuments = append(createdDocuments, createdDoc.Value.DidDoc)
	}

	// Get verification method and obtain DID URLs
	verificationMaterial := createdDocuments[0].VerificationMethod[0].VerificationMaterial
	controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, verificationMaterial)
	suite.Require().NoError(err)

	suite.Require().Equal(documentsAmount, len(controlledDocuments))
	for _, doc := range createdDocuments {
		suite.Require().Contains(controlledDocuments, doc.Id)
	}

	// Deactivate one of documents
	documentToDeactivate := createdDocuments[0]
	remainingDocument := createdDocuments[1]
	payload := &types.MsgDeactivateDIDDocumentPayload{
		Id:        documentToDeactivate.Id,
		VersionId: uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: documentToDeactivate.Id + "#key-1",
			Key:                  keypair.Private,
		},
	}

	_, err = didutil.DeactivateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// Check if document was removed from index
	updatedControlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, verificationMaterial)
	suite.Require().NoError(err)
	suite.Require().Equal(documentsAmount-1, len(updatedControlledDocuments))
	suite.Require().NotContains(updatedControlledDocuments, documentToDeactivate.Id)
	suite.Require().Contains(updatedControlledDocuments, remainingDocument.Id)
}

func (suite *KeeperTestSuite) TestShouldUpdateIndexOnDocumentUpdate() {
	documentsAmount := 2
	keypair := didutil.GenerateKeyPair()

	// Generate 2 did documents with same verification method
	var createdDocuments []*types.DIDDocument
	for i := 0; i < documentsAmount; i++ {
		did := didutil.GenerateDID(didutil.Base58_16bytes)
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
		createdDoc, err := didutil.CreateDID(suite.ctx, suite.keeper, payload, signatures)
		suite.Require().NoError(err)

		createdDocuments = append(createdDocuments, createdDoc.Value.DidDoc)
	}

	// Get verification method and obtain DID URLs
	verificationMaterial := createdDocuments[0].VerificationMethod[0].VerificationMaterial
	controlledDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, verificationMaterial)
	suite.Require().NoError(err)

	suite.Require().Equal(documentsAmount, len(controlledDocuments))
	for _, doc := range createdDocuments {
		suite.Require().Contains(controlledDocuments, doc.Id)
	}

	// Update verification method in one document
	documentToUpdate := createdDocuments[0]
	newKeypair := didutil.GenerateKeyPair()
	payload := &types.MsgUpdateDIDDocumentPayload{
		Id: documentToUpdate.Id,
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     documentToUpdate.Id + "#key-2",
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             documentToUpdate.Id,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(newKeypair.Public),
			},
			{
				Id:                     documentToUpdate.Id + "#key-1",
				VerificationMethodType: types.Ed25519VerificationKey2020Type,
				Controller:             documentToUpdate.Id,
				VerificationMaterial:   didutil.GenerateEd25519VerificationKey2020VerificationMaterial(keypair.Public),
			},
		},
		Authentication:  []string{documentToUpdate.Id + "#key-2", documentToUpdate.Id + "#key-1"},
		AssertionMethod: []string{documentToUpdate.Id + "#key-2", documentToUpdate.Id + "#key-1"},
		VersionId:       uuid.NewString(),
	}

	signatures := []didutil.SignInput{
		{
			VerificationMethodID: documentToUpdate.Id + "#key-1",
			Key:                  keypair.Private,
		},
	}
	_, err = didutil.UpdateDIDDocument(suite.ctx, suite.keeper, payload, signatures)
	suite.Require().NoError(err)

	// Check if document was indexed correctly (index still have previous verification material + new verification material)
	updatedDocuments, err := suite.keeper.GetDIDsControlledBy(suite.ctx, verificationMaterial)
	suite.Require().NoError(err)

	suite.Require().Equal(documentsAmount, len(updatedDocuments))
	for _, doc := range createdDocuments {
		suite.Require().Contains(updatedDocuments, doc.Id)
	}

	// Check if new verification method was indexed
	docsControlledByNewKey, err := suite.keeper.GetDIDsControlledBy(suite.ctx, payload.VerificationMethod[0].VerificationMaterial)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(docsControlledByNewKey))
	suite.Require().Contains(docsControlledByNewKey, documentToUpdate.Id)
}
