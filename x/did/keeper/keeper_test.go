package keeper_test

import (
	"testing"
	"context"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/stretchr/testify/suite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"swisstronik/simapp"
	didutil "swisstronik/testutil/did"
	"swisstronik/x/did/keeper"
	"swisstronik/x/did/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	goCtx		context.Context
	keeper      keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	app, _ := simapp.Setup(t, false)
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{ChainID: "swisstronik_1291-1"})
	s.goCtx = sdk.WrapSDKContext(s.ctx)
	s.keeper = app.DIDKeeper
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
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
	msg := didutil.CreateDIDMessage(payload, signatures)
	_, err := suite.keeper.CreateDIDDocument(suite.goCtx, msg)
	suite.Require().NoError(err)

	// check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(msg.Payload.ToDidDoc(), *resp.Value.DidDoc) 
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
	msg := didutil.CreateDIDMessage(payload, signatures)
	_, err := suite.keeper.CreateDIDDocument(suite.goCtx, msg)
	suite.Require().NoError(err)

	// Check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(msg.Payload.ToDidDoc(), *resp.Value.DidDoc)
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
	msg := didutil.CreateDIDMessage(payload, signatures)
	_, err := suite.keeper.CreateDIDDocument(suite.goCtx, msg)
	suite.Require().NoError(err)

	// Check if document was created
	resp, err := suite.keeper.DIDDocument(suite.goCtx, &types.QueryDIDDocumentRequest{Id: did})
	suite.Require().NoError(err)
	suite.Require().Equal(msg.Payload.ToDidDoc(), *resp.Value.DidDoc)
}