package keeper_test

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/tests"
	"swisstronik/testutil"
	"swisstronik/utils"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
	feemarkettypes "swisstronik/x/feemarket/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	goCtx     context.Context
	keeper    keeper.Keeper
	validator stakingtypes.Validator
	app       *app.App
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
	suite.app, _ = app.SetupSwissApp(false, nil, chainID)

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	address := common.BytesToAddress(priv.PubKey().Address().Bytes())

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)
	suite.goCtx = sdk.WrapSDKContext(suite.ctx)
	suite.keeper = suite.app.ComplianceKeeper

	//// bond denom
	//stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	//stakingParams.BondDenom = utils.BaseDenom
	//err = suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
	//require.NoError(t, err)

	feeParams := feemarkettypes.DefaultParams()
	feeParams.MinGasPrice = sdk.NewDec(1)
	_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, feeParams)
	suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, sdk.ZeroInt().BigInt())

	// Set Validator
	valAddr := sdk.ValAddress(address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)

	validators := s.app.StakingKeeper.GetValidators(s.ctx, 2)
	// set a bonded validator that takes part in consensus
	if validators[0].Status == stakingtypes.Bonded {
		suite.validator = validators[0]
	} else {
		suite.validator = validators[1]
	}
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, t)
	suite.Require().NoError(err)
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

	has, err := suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 1715018692, []sdk.Address{issuer})
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
	details := &types.IssuerDetails{Name: "testIssuer"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	issuer := sdk.AccAddress(from.Bytes())
	err := suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
	suite.Require().NoError(err)

	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)
	suite.Require().NoError(err)

	from, _ = tests.RandomEthAddressWithPrivateKey()
	signer := sdk.AccAddress(from.Bytes())

	verificationDetails := &types.VerificationDetails{
		IssuerAddress:       issuer.String(),
		OriginChain:         "test chain",
		IssuanceTimestamp:   1712018692,
		ExpirationTimestamp: 1715018692,
		OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
	}
	// Try to add verification details without verification type
	verificationId, err := suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_UNSPECIFIED,
		verificationDetails,
	)
	suite.Require().Error(err)
	suite.Require().Nil(verificationId)

	verificationId, err = suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_CREDIT_SCORE+1,
		verificationDetails,
	)
	suite.Require().Error(err)
	suite.Require().Nil(verificationId)

	// Allow to add verification details verified by active issuer
	verificationId, err = suite.keeper.AddVerificationDetails(
		suite.ctx,
		signer,
		types.VerificationType_VT_KYC,
		verificationDetails,
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(verificationId)

	has, err := suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 1715018692, []sdk.Address{issuer})
	suite.Require().NoError(err)
	suite.Require().True(has)

	// No provided issuer, but has verification details
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 1715018692, nil)
	suite.Require().NoError(err)
	suite.Require().True(has)

	// Check if it has valid verification details
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 1715018692-1, nil)
	suite.Require().NoError(err)
	suite.Require().True(has)
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 1715018692+1, nil)
	suite.Require().NoError(err)
	suite.Require().False(has)
	// Check if it has valid verification details within current time
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 0, nil)
	suite.Require().NoError(err)
	suite.Require().False(has)
	has, err = suite.keeper.HasVerificationOfType(suite.ctx, signer, types.VerificationType_VT_KYC, 0, []sdk.Address{issuer})
	suite.Require().NoError(err)
	suite.Require().False(has)
}

func (suite *KeeperTestSuite) TestAddressDetailsCRUD() {
	issuerDetails := &types.IssuerDetails{Name: "testIssuer"}
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
	issuerDetails := &types.IssuerDetails{Name: "testIssuer"}
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

func (suite *KeeperTestSuite) TestInvalidOperatorType() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	operator := sdk.AccAddress(from.Bytes())

	err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_UNSPECIFIED)
	suite.Require().Error(err)

	err = suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR+1)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestInitialOperator() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	operator := sdk.AccAddress(from.Bytes())

	err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_INITIAL)
	suite.Require().NoError(err)

	operatorDetails, err := suite.keeper.GetOperatorDetails(suite.ctx, operator)
	suite.Require().NoError(err)
	suite.Require().Equal(operator.String(), operatorDetails.Operator)
	suite.Require().Equal(operatorDetails.OperatorType, types.OperatorType_OT_INITIAL)

	// Can not remove initial operator
	err = suite.keeper.RemoveRegularOperator(suite.ctx, operator)
	suite.Require().Error(err)

	exists, err := suite.keeper.OperatorExists(suite.ctx, operator)
	suite.Require().True(exists)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestRegularOperator() {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	operator := sdk.AccAddress(from.Bytes())

	err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
	suite.Require().NoError(err)

	operatorDetails, err := suite.keeper.GetOperatorDetails(suite.ctx, operator)
	suite.Require().NoError(err)
	suite.Require().Equal(operator.String(), operatorDetails.Operator)
	suite.Require().Equal(operatorDetails.OperatorType, types.OperatorType_OT_REGULAR)

	err = suite.keeper.RemoveRegularOperator(suite.ctx, operator)
	suite.Require().NoError(err)

	exists, err := suite.keeper.OperatorExists(suite.ctx, operator)
	suite.Require().False(exists)
	suite.Require().NoError(err)
}
