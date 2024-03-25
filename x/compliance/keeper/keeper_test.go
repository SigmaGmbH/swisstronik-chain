package keeper_test

import (
	"context"
	"swisstronik/tests"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/utils"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
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
	RunSpecs(t, "Compliance Keeper Suite")
}

func (suite *KeeperTestSuite) Setup(t *testing.T) {
	chainID := utils.TestnetChainID + "-1"
	app, _ := app.SetupSwissApp(false, nil, chainID)
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{ChainID: chainID})
	s.goCtx = sdk.WrapSDKContext(s.ctx)
	s.keeper = app.ComplianceKeeper
}

func (suite *KeeperTestSuite) TestCreateSimpleAndFetchSimpleIssuer() {
	details := &types.IssuerDetails{Name: "testIssuer"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	i, err := suite.keeper.GetIssuerDetails(suite.ctx, address)
	suite.Require().Equal(details, i)
	suite.Require().NoError(err)
	suite.keeper.RemoveIssuer(suite.ctx, address)
	i, err = suite.keeper.GetIssuerDetails(suite.ctx, address)
	suite.Require().Equal("", i.Name)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestNonExistingIssuer() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	i, err := suite.keeper.GetIssuerDetails(suite.ctx, address)
	suite.Require().Equal("", i.Name)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestAddressDetailsCRUD() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	details := &types.AddressDetails{IsVerified: true,
		IsRevoked: false,
		Verifications: []*types.Verification{{
			Type:           types.VerificationType_VT_KYC,
			VerificationId: nil,
			IssuerAddress:  from.String(),
		}}}
	err := suite.keeper.SetAddressDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	i, err := suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(details, i)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestAddressVerified() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	details := &types.AddressDetails{IsVerified: true,
		IsRevoked: false,
		Verifications: make([]*types.Verification, 0)}
	err := suite.keeper.SetAddressDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	i, err := suite.keeper.IsAddressVerified(suite.ctx, address)
	suite.Require().Equal(true, i)
	from2, _ := tests.RandomEthAddressWithPrivateKey()
	address2 := sdk.AccAddress(from2.Bytes())
	details2 := &types.AddressDetails{IsVerified: false,
		IsRevoked: false,
		Verifications: make([]*types.Verification, 0)}
	err = suite.keeper.SetAddressDetails(suite.ctx, address2, details2)
	suite.Require().NoError(err)
	i, err = suite.keeper.IsAddressVerified(suite.ctx, address2)
	suite.Require().Equal(false, i)
}


func (suite *KeeperTestSuite) TestAddressDetailsSetVerificationStatus() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	details := &types.AddressDetails{
		IsVerified: false,
		IsRevoked: false,
		Verifications: []*types.Verification{{
			Type:           types.VerificationType_VT_KYC,
			VerificationId: nil,
			IssuerAddress:  from.String(),
		}}}
	err := suite.keeper.SetAddressDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	// set to true
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx,address, true)
	suite.Require().NoError(err)
	i, err := suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(true, i.IsVerified)
	suite.Require().NoError(err)
	// set to false
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx,address, false)
	suite.Require().NoError(err)
	i, err = suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(false, i.IsVerified)
	suite.Require().NoError(err)
	// NOOP
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx,address, false)
	suite.Require().NoError(err)
	i, err = suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(false, i.IsVerified)
	suite.Require().NoError(err)
}
