package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"swisstronik/tests"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

type ProposalTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper

	issuerCreator sdk.AccAddress
	validIssuer   sdk.AccAddress
}

func TestProposalTestSuite(t *testing.T) {
	s := new(ProposalTestSuite)
	k, ctx := testkeeper.ComplianceKeeper(t)
	s.ctx = ctx
	s.keeper = *k

	suite.Run(t, s)
}

func (suite *ProposalTestSuite) SetupTest() {
	suite.issuerCreator = tests.RandomAccAddress()
	suite.validIssuer = tests.RandomAccAddress()

	// Set issuer details
	issuerDetails := &types.IssuerDetails{Name: "testIssuer"}
	err := suite.keeper.SetIssuerDetails(suite.ctx, suite.issuerCreator, suite.validIssuer, issuerDetails)
	suite.Require().NoError(err)

	// Set verification status as true for issuer details
	err = suite.keeper.SetAddressVerificationStatus(suite.ctx, suite.validIssuer, true)
	suite.Require().NoError(err)
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("compliance", (&types.VerifyIssuerProposal{}).ProposalRoute())
	suite.Require().Equal("VerifyIssuer", (&types.VerifyIssuerProposal{}).ProposalType())
}

func (suite *ProposalTestSuite) TestVerifyIssuerProposal() {
	testCases := []struct {
		name          string
		title         string
		description   string
		issuerAddress string
		expected      bool
	}{
		{"verified issuer", "test", "description", suite.validIssuer.String(), true},
		{"not verified issuer", "test", "description", tests.RandomAccAddress().String(), true},
		{"invalid issuer", "test", "description", "invalid address", false},
	}

	for _, tc := range testCases {
		tx := types.NewVerifyIssuerProposal(tc.title, tc.description, tc.issuerAddress)
		err := tx.ValidateBasic()

		if tc.expected {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}
