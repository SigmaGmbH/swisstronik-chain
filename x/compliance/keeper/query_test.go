package keeper_test

import (
	"context"
	"encoding/base64"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/tests"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

type QuerierTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	goCtx   context.Context
	keeper  keeper.Keeper
	querier keeper.Querier
	app     *app.App

	issuerCreator sdk.AccAddress
	issuer        sdk.AccAddress
	user          sdk.AccAddress
}

func TestQuerierTestSuite(t *testing.T) {
	s := new(QuerierTestSuite)
	k, ctx := testkeeper.ComplianceKeeper(t)
	s.ctx = ctx
	s.goCtx = sdk.WrapSDKContext(ctx)
	s.keeper = *k
	s.querier = keeper.Querier{Keeper: s.keeper}

	suite.Run(t, s)
}

func (suite *QuerierTestSuite) SetupTest() {
	suite.issuerCreator = tests.RandomAccAddress()
	suite.issuer = tests.RandomAccAddress()
	suite.user = tests.RandomAccAddress()

	// Create issuer
	issuerDetails := &types.IssuerDetails{Creator: suite.issuerCreator.String(), Name: "testIssuer"}
	err := suite.keeper.SetIssuerDetails(suite.ctx, suite.issuer, issuerDetails)
	suite.Require().NoError(err)

	// Set verification status as true for issuer details
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, suite.issuer, true)
	suite.Require().NoError(err)

	// Add address details
	err = suite.keeper.SetAddressDetails(
		suite.ctx,
		suite.user,
		&types.AddressDetails{
			IsVerified: true,
			IsRevoked:  false,
		})

	// Add verification details and address details
	verificationId, err := suite.keeper.AddVerificationDetails(
		suite.ctx,
		suite.user,
		types.VerificationType_VT_KYC,
		&types.VerificationDetails{
			IssuerAddress:       suite.issuer.String(),
			OriginChain:         "test chain",
			IssuanceTimestamp:   1712018692,
			ExpirationTimestamp: 1715018692,
			OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(verificationId)
}

func (suite *QuerierTestSuite) TestSuccess() {
	// Query issuer details
	issuerRequest := &types.QueryIssuerDetailsRequest{
		IssuerAddress: suite.issuer.String(),
	}
	issuerDetails, err := suite.querier.IssuerDetails(suite.goCtx, issuerRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(issuerDetails.Details.Name, "testIssuer")

	// Query address details
	addressRequest := &types.QueryAddressDetailsRequest{
		Address: suite.user.String(),
	}
	addressDetails, err := suite.querier.AddressDetails(suite.goCtx, addressRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(addressDetails.Data.IsVerified, true)
	suite.Require().Greater(len(addressDetails.Data.Verifications), 0)

	verification := addressDetails.Data.Verifications[0]

	// Bytes are base64 encoded and passed to querier
	verificationId := base64.StdEncoding.EncodeToString(verification.VerificationId)

	// Query verification details
	verificationRequest := &types.QueryVerificationDetailsRequest{
		VerificationID: verificationId,
	}
	verificationDetails, err := suite.querier.VerificationDetails(suite.goCtx, verificationRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(verificationDetails.Details.IssuerAddress, verification.IssuerAddress)
}

func (suite *QuerierTestSuite) TestFailed() {
	anyUser := tests.RandomAccAddress()

	// Query invalid issuer details
	issuerRequest := &types.QueryIssuerDetailsRequest{
		IssuerAddress: "invalid issuer",
	}
	issuerDetails, err := suite.querier.IssuerDetails(suite.goCtx, issuerRequest)
	suite.Require().Error(err) // Failed in parsing acc address

	issuerRequest = &types.QueryIssuerDetailsRequest{
		IssuerAddress: anyUser.String(),
	}
	issuerDetails, err = suite.querier.IssuerDetails(suite.goCtx, issuerRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(issuerDetails.Details, &types.IssuerDetails{}) // Empty details, not found

	// Query invalid address details
	addressRequest := &types.QueryAddressDetailsRequest{
		Address: "invalid address",
	}
	addressDetails, err := suite.querier.AddressDetails(suite.goCtx, addressRequest)
	suite.Require().Error(err) // Failed in parsing acc address

	addressRequest = &types.QueryAddressDetailsRequest{
		Address: anyUser.String(),
	}
	addressDetails, err = suite.querier.AddressDetails(suite.goCtx, addressRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(addressDetails.Data, &types.AddressDetails{})

	// Query invalid verification details
	verificationRequest := &types.QueryVerificationDetailsRequest{
		VerificationID: "invalid verification id",
	}
	verificationDetails, err := suite.querier.VerificationDetails(suite.goCtx, verificationRequest)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "base64")

	verificationRequest = &types.QueryVerificationDetailsRequest{
		VerificationID: "7xeEYWEV2Krw4ikPFHcJZIdiJNk5AtcbTX7QqNhY7hQ=", // random verification id
	}
	verificationDetails, err = suite.querier.VerificationDetails(suite.goCtx, verificationRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(verificationDetails.Details, &types.VerificationDetails{})
}
