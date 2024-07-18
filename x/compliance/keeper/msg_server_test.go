package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/status-im/keycard-go/hexutils"

	"swisstronik/tests"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

func (suite *KeeperTestSuite) TestAddOperator() {
	var (
		operator    sdk.AccAddress
		newOperator sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgAddOperator
		expected func(resp *types.MsgAddOperatorResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgAddOperator {
				msg := types.NewMsgAddOperator(
					"operator address",
					"new operator address",
				)
				return &msg
			},
			expected: func(resp *types.MsgAddOperatorResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "operator not exist",
			init: func() {
				operator = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgAddOperator {
				msg := types.NewMsgAddOperator(
					operator.String(),
					operator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgAddOperatorResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrNotOperator)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				newOperator = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgAddOperator {
				msg := types.NewMsgAddOperator(
					operator.String(),
					newOperator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgAddOperatorResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgAddOperatorResponse{})

				// Operator should exist
				exist, err := suite.keeper.OperatorExists(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().True(exist)

				// Check operator details
				details, err := suite.keeper.GetOperatorDetails(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().Equal(newOperator.String(), details.Operator)
			},
		},
		{
			name: "existing operator",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgAddOperator {
				msg := types.NewMsgAddOperator(
					operator.String(),
					operator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgAddOperatorResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrInvalidOperator)
				suite.Require().Nil(resp)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleAddOperator(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveOperator() {
	var (
		operator    sdk.AccAddress
		newOperator sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgRemoveOperator
		expected func(resp *types.MsgRemoveOperatorResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgRemoveOperator {
				msg := types.NewMsgRemoveOperator(
					"operator address",
					"new operator address",
				)
				return &msg
			},
			expected: func(resp *types.MsgRemoveOperatorResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "operator not exist",
			init: func() {
				operator = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgRemoveOperator {
				msg := types.NewMsgRemoveOperator(
					operator.String(),
					operator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgRemoveOperatorResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrNotOperatorOrIssuerCreator)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				newOperator = tests.RandomAccAddress()
				err = suite.keeper.AddOperator(suite.ctx, newOperator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgRemoveOperator {
				msg := types.NewMsgRemoveOperator(
					operator.String(),
					newOperator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgRemoveOperatorResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgRemoveOperatorResponse{})

				// Operator should exist
				exist, err := suite.keeper.OperatorExists(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().False(exist)

				// Check operator details
				details, err := suite.keeper.GetOperatorDetails(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().Equal(details, &types.OperatorDetails{})
			},
		},
		{
			name: "remove itself",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgRemoveOperator {
				msg := types.NewMsgRemoveOperator(
					operator.String(),
					operator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgRemoveOperatorResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrInvalidOperator)
				suite.Require().Nil(resp)

				// Operator should exist
				exist, err := suite.keeper.OperatorExists(suite.ctx, operator)
				suite.Require().NoError(err)
				suite.Require().True(exist)

				// Check operator details
				details, err := suite.keeper.GetOperatorDetails(suite.ctx, operator)
				suite.Require().NoError(err)
				suite.Require().Equal(operator.String(), details.Operator)
			},
		},
		{
			name: "remove initial operator",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_INITIAL)
				suite.Require().NoError(err)

				newOperator = tests.RandomAccAddress()

				err = suite.keeper.AddOperator(suite.ctx, newOperator, types.OperatorType_OT_INITIAL)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgRemoveOperator {
				msg := types.NewMsgRemoveOperator(
					operator.String(),
					newOperator.String(),
				)
				return &msg
			},
			expected: func(resp *types.MsgRemoveOperatorResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrNotAuthorized)
				suite.Require().Nil(resp)

				// Operator should exist
				exist, err := suite.keeper.OperatorExists(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().True(exist)

				// Check operator details
				details, err := suite.keeper.GetOperatorDetails(suite.ctx, newOperator)
				suite.Require().NoError(err)
				suite.Require().Equal(newOperator.String(), details.Operator)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleRemoveOperator(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestSetVerificationStatus() {
	var (
		operator sdk.AccAddress
		issuer   sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgSetVerificationStatus
		expected func(resp *types.MsgSetVerificationStatusResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgSetVerificationStatus {
				msg := types.NewMsgSetVerificationStatus(
					"operator address",
					"issuer address",
					true,
				)
				return &msg
			},
			expected: func(resp *types.MsgSetVerificationStatusResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "operator not exist",
			init: func() {
				operator = tests.RandomAccAddress()
				issuer = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgSetVerificationStatus {
				msg := types.NewMsgSetVerificationStatus(
					operator.String(),
					issuer.String(),
					true,
				)
				return &msg
			},
			expected: func(resp *types.MsgSetVerificationStatusResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrNotOperator)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer not exist",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgSetVerificationStatus {
				msg := types.NewMsgSetVerificationStatus(
					operator.String(),
					issuer.String(),
					true,
				)
				return &msg
			},
			expected: func(resp *types.MsgSetVerificationStatusResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrInvalidIssuer)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()

				details := &types.IssuerDetails{Creator: tests.RandomAccAddress().String(), Name: "testIssuer"}
				err = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgSetVerificationStatus {
				msg := types.NewMsgSetVerificationStatus(
					operator.String(),
					issuer.String(),
					true,
				)
				return &msg
			},
			expected: func(resp *types.MsgSetVerificationStatusResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgSetVerificationStatusResponse{})

				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().True(verified)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleSetVerificationStatus(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestCreateIssuer() {
	var (
		operator sdk.AccAddress
		creator  sdk.AccAddress
		issuer   sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgCreateIssuer
		expected func(resp *types.MsgCreateIssuerResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgCreateIssuer {
				msg := types.NewCreateIssuerMsg(
					"operator address",
					"issuer address",
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgCreateIssuerResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "invalid issuer",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgCreateIssuer {
				msg := types.NewCreateIssuerMsg(
					operator.String(),
					"issuer address",
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgCreateIssuerResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success-operator as creator",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgCreateIssuer {
				msg := types.NewCreateIssuerMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgCreateIssuerResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgCreateIssuerResponse{})

				// Issuer should exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().True(issuerExists)
				suite.Require().NoError(err)

				// Should be revoked verification if issuer address was verified
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(false, addressDetails.IsVerified)

				// Check if issuer details are stored correctly
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(operator.String(), details.Creator)
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)

				// Check if issuer's verification status is false
				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			name: "success-any creator",
			init: func() {
				creator = tests.RandomAccAddress()
				issuer = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgCreateIssuer {
				msg := types.NewCreateIssuerMsg(
					creator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgCreateIssuerResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgCreateIssuerResponse{})

				// Issuer should exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().True(issuerExists)
				suite.Require().NoError(err)

				// Should be revoked verification if issuer address was verified
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(false, addressDetails.IsVerified)

				// Check if issuer details are stored correctly
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().NotEqual(operator.String(), details.Creator)
				suite.Require().Equal(creator.String(), details.Creator)
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)

				// Check if issuer's verification status is false
				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			name: "existing issuer",
			init: func() {
				details := &types.IssuerDetails{Creator: tests.RandomAccAddress().String(), Name: "test issuer"}
				issuer = tests.RandomAccAddress()
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				operator = tests.RandomAccAddress()

				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgCreateIssuer {
				msg := types.NewCreateIssuerMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgCreateIssuerResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrInvalidIssuer)
				suite.Require().Nil(resp)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleCreateIssuer(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateIssuerDetails() {
	var (
		creator  sdk.AccAddress
		issuer   sdk.AccAddress
		operator sdk.AccAddress
		signer   sdk.AccAddress
	)

	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgUpdateIssuerDetails
		expected func(response *types.MsgUpdateIssuerDetailsResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					"operator address",
					"issuer address",
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().ErrorContains(err, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer creator does not match",
			init: func() {
				issuer = tests.RandomAccAddress()
				creator = tests.RandomAccAddress()
				operator = tests.RandomAccAddress()

				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: creator.String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().ErrorIs(err, types.ErrNotOperatorOrIssuerCreator)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer does not exist",
			init: func() {
				issuer = tests.RandomAccAddress()
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().ErrorIs(err, types.ErrInvalidIssuer)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success-operator as creator",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: operator.String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().NoError(err)
				suite.Require().Equal(resp, &types.MsgUpdateIssuerDetailsResponse{})

				// Issuer should exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().True(issuerExists)
				suite.Require().NoError(err)

				// Should be revoked verification if issuer address was verified
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(false, addressDetails.IsVerified)

				// Check if issuer details are stored correctly
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(operator.String(), details.Creator)
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)

				// Check if issuer was revoked
				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			name: "success-any creator",
			init: func() {
				creator = tests.RandomAccAddress()
				issuer = tests.RandomAccAddress()

				details := &types.IssuerDetails{Creator: creator.String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					creator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().NoError(err)
				suite.Require().Equal(resp, &types.MsgUpdateIssuerDetailsResponse{})

				// Issuer should exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().True(issuerExists)
				suite.Require().NoError(err)

				// Should be revoked verification if issuer address was verified
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(false, addressDetails.IsVerified)

				// Check if issuer details are stored correctly
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().NotEqual(operator.String(), details.Creator)
				suite.Require().Equal(creator.String(), details.Creator)
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)

				// Check if issuer was revoked
				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			// Should revoke verification for all the accounts verified by updated issuer
			name: "past verification data still exists",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				creator = tests.RandomAccAddress()
				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: creator.String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				_ = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)

				signer = tests.RandomAccAddress()

				// Add address details with verification details
				_, _ = suite.keeper.AddVerificationDetails(
					suite.ctx,
					signer,
					types.VerificationType_VT_KYC,
					&types.VerificationDetails{
						IssuerAddress:       issuer.String(), // use same issuer address
						OriginChain:         "test chain",
						IssuanceTimestamp:   1712018692,
						ExpirationTimestamp: 1715018692,
						OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
					},
				)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
					issuer.String(),
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				return &msg
			},
			expected: func(resp *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().NoError(err)
				suite.Require().Equal(resp, &types.MsgUpdateIssuerDetailsResponse{})

				// Skip duplicated checks
				// Check if verification data still exists
				details, err := suite.keeper.GetVerificationsOfType(
					suite.ctx,
					signer,
					types.VerificationType_VT_KYC,
					issuer,
				)
				suite.Require().NoError(err)
				suite.Require().Equal(1, len(details))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleUpdateIssuerDetails(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveIssuer() {
	var (
		issuer   sdk.AccAddress
		operator sdk.AccAddress
		signer   sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgRemoveIssuer
		expected func(resp *types.MsgRemoveIssuerResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg("operator", "issuer address")
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().ErrorContains(err, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer not exist",
			init: func() {
				issuer = tests.RandomAccAddress()
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().ErrorIs(err, types.ErrInvalidIssuer)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "not operator",
			init: func() {
				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: tests.RandomAccAddress().String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				operator = tests.RandomAccAddress()
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().ErrorIs(err, types.ErrNotOperatorOrIssuerCreator)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: tests.RandomAccAddress().String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().NoError(err)
				suite.Require().Equal(resp, &types.MsgRemoveIssuerResponse{})

				// Issuer should not exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().False(issuerExists)
				suite.Require().NoError(err)

				// Same for issuer details
				issuerDetails, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(issuerDetails, &types.IssuerDetails{})

				// Address details for removed issuer should not exit
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(addressDetails, &types.AddressDetails{})
			},
		},
		{
			// Should revoke verification for all the accounts verified by removed issuer
			name: "account was suspended",
			init: func() {
				operator = tests.RandomAccAddress()
				err := suite.keeper.AddOperator(suite.ctx, operator, types.OperatorType_OT_REGULAR)
				suite.Require().NoError(err)

				issuer = tests.RandomAccAddress()
				details := &types.IssuerDetails{Creator: tests.RandomAccAddress().String(), Name: "test issuer"}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				_ = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)

				signer = tests.RandomAccAddress()

				// Add address details with verification details
				_, _ = suite.keeper.AddVerificationDetails(
					suite.ctx,
					signer,
					types.VerificationType_VT_KYC,
					&types.VerificationDetails{
						IssuerAddress:       issuer.String(), // use same issuer address
						OriginChain:         "test chain",
						IssuanceTimestamp:   1712018692,
						ExpirationTimestamp: 1715018692,
						OriginalData:        hexutils.HexToBytes("B639DF194671CDE06EFAA368A404F72E3306DF0359117AC7E78EC2BE04B7629D"),
					},
				)
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().NoError(err)
				suite.Require().Equal(resp, &types.MsgRemoveIssuerResponse{})

				// Skip duplicated checks
				// Check if verification data for removed issuer was removed
				details, err := suite.keeper.GetVerificationsOfType(
					suite.ctx,
					signer,
					types.VerificationType_VT_KYC,
					issuer,
				)
				suite.Require().NoError(err)
				suite.Require().Equal(0, len(details))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			if tc.init != nil {
				tc.init()
			}
			msg := tc.malleate()
			resp, err := msgServer.HandleRemoveIssuer(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}
