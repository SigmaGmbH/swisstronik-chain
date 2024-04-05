package keeper_test

import (
	"context"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/tests"
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

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("swtr", "swtrpub")
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
	// todo, operator is empty
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestSuspendedIssuer() {
	details := &types.IssuerDetails{Name: "testIssuer"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
	suite.Require().NoError(err)

	// Revoke verification status for test issuer
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, false)
	suite.Require().NoError(err)

	from, _ = tests.RandomEthAddressWithPrivateKey()
	signer := sdk.AccAddress(from.Bytes())

	// Should not allow to add verification details verified by suspended issuer
	// Even if issuer was suspended, verification data should exist
	verificationId, err := suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_KYC,
		&types.VerificationDetails{
			IssuerAddress:       issuer.String(),
			OriginChain:         "test chain",
			IssuanceTimestamp:   1712018692,
			ExpirationTimestamp: 1715018692,
			OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
		},
	)
	suite.Require().Error(err)
	suite.Require().Nil(verificationId)

	has, err := suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, []sdk.Address{issuer})
	suite.Require().NoError(err)
	suite.Require().False(has)
}

func (suite *KeeperTestSuite) TestRemovedIssuer() {
	issuerDetails := &types.IssuerDetails{Name: "testIssuer"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, issuerDetails)
	suite.Require().NoError(err)

	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)
	suite.Require().NoError(err)

	from, _ = tests.RandomEthAddressWithPrivateKey()
	signer := sdk.AccAddress(from.Bytes())

	// Add dummy verification details and address details with verifications
	err = suite.keeper.SetAddressDetails(
		suite.ctx,
		issuer,
		&types.AddressDetails{
			IsVerified: true,
			IsRevoked:  false,
		})
	verificationId, err := suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_KYC,
		&types.VerificationDetails{
			IssuerAddress:       issuer.String(),
			OriginChain:         "test chain",
			IssuanceTimestamp:   1712018692,
			ExpirationTimestamp: 1715018692,
			OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
		},
	)
	suite.Require().NoError(err)

	suite.keeper.RemoveIssuer(suite.ctx, issuer)
	i, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
	suite.Require().Equal(i, &types.IssuerDetails{})
	suite.Require().NoError(err)

	exists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
	suite.Require().False(exists)
	suite.Require().NoError(err)

	// AddressDetails for removed issuer should not exist
	addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
	suite.Require().Equal(addressDetails, &types.AddressDetails{})
	suite.Require().NoError(err)

	// If issuer was removed, all the verification details which were verified by removed issuer
	// should be removed every time when call `GetVerificationDetails` or `GetAddressDetails`.
	verificationDetails, err := suite.keeper.GetVerificationsOfType(
		suite.ctx,
		signer,
		types.VerificationType_VT_KYC,
		issuer,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(verificationDetails))

	verificationDetailsBy, err := suite.keeper.GetVerificationDetails(suite.ctx, verificationId)
	suite.Require().NoError(err)
	suite.Require().Equal(verificationDetailsBy, &types.VerificationDetails{})
}

func (suite *KeeperTestSuite) TestAddVerificationDetails() {
	details := &types.IssuerDetails{Name: "testIssuer", Operator: "testOperator"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
	suite.Require().NoError(err)

	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)
	suite.Require().NoError(err)

	from, _ = tests.RandomEthAddressWithPrivateKey()
	signer := sdk.AccAddress(from.Bytes())

	// Allow to add verification details verified by active issuer
	verificationId, err := suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_KYC,
		&types.VerificationDetails{
			IssuerAddress:       issuer.String(),
			OriginChain:         "test chain",
			IssuanceTimestamp:   1712018692,
			ExpirationTimestamp: 1715018692,
			OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(verificationId)

	has, err := suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, []sdk.Address{issuer})
	suite.Require().NoError(err)
	suite.Require().True(has)

	// No provided issuer, but has verification details
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, nil)
	suite.Require().NoError(err)
	suite.Require().True(has)
}

func (suite *KeeperTestSuite) TestAddressDetailsCRUD() {
	issuerDetails := &types.IssuerDetails{Name: "testIssuer", Operator: "testOperator"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, issuerDetails)
	suite.Require().NoError(err)

	from, _ = tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())

	addressDetails := &types.AddressDetails{IsVerified: true,
		IsRevoked: false,
		Verifications: []*types.Verification{{
			Type:           types.VerificationType_VT_KYC,
			VerificationId: nil,
			IssuerAddress:  issuer.String(),
		}}}
	err = suite.keeper.SetAddressDetails(suite.ctx, address, addressDetails)
	suite.Require().NoError(err)
	i, err := suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(addressDetails, i)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestAddressVerified() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	address := sdk.AccAddress(from.Bytes())
	details := &types.AddressDetails{IsVerified: true,
		IsRevoked:     false,
		Verifications: make([]*types.Verification, 0)}
	err := suite.keeper.SetAddressDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	i, err := suite.keeper.IsAddressVerified(suite.ctx, address)
	suite.Require().Equal(true, i)
	from2, _ := tests.RandomEthAddressWithPrivateKey()
	address2 := sdk.AccAddress(from2.Bytes())
	details2 := &types.AddressDetails{IsVerified: false,
		IsRevoked:     false,
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
		IsRevoked:  false,
		Verifications: []*types.Verification{{
			Type:           types.VerificationType_VT_KYC,
			VerificationId: nil,
			IssuerAddress:  address.String(),
		}}}
	err := suite.keeper.SetAddressDetails(suite.ctx, address, details)
	suite.Require().NoError(err)
	// set to true
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, address, true)
	suite.Require().NoError(err)
	i, err := suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(true, i.IsVerified)
	suite.Require().NoError(err)
	// set to false
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, address, false)
	suite.Require().NoError(err)
	i, err = suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(false, i.IsVerified)
	suite.Require().NoError(err)
	// NOOP
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, address, false)
	suite.Require().NoError(err)
	i, err = suite.keeper.GetAddressDetails(suite.ctx, address)
	suite.Require().Equal(false, i.IsVerified)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestSetVerificationDetails() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	issuerDetails := &types.IssuerDetails{Name: "testIssuer", Operator: "testOperator"}
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, issuerDetails)
	suite.Require().NoError(err)

	verificationDetails := &types.VerificationDetails{
		IssuerAddress:       issuer.String(),
		OriginChain:         "test chain",
		IssuanceTimestamp:   1712018692,
		ExpirationTimestamp: 1715018692,
		OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
	}
	verificationId := hexutils.HexToBytes("83456ef3b8ea6777da69d1509cf51861985e2b4e24cf7f5d4c5080996bf8cf4e")
	err = suite.keeper.SetVerificationDetails(suite.ctx, verificationId, verificationDetails)
	suite.Require().NoError(err)

	resp, err := suite.keeper.GetVerificationDetails(suite.ctx, verificationId)
	suite.Require().NoError(err)
	suite.Require().Equal(verificationDetails, resp)
}
