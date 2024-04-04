package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/status-im/keycard-go/hexutils"
	"swisstronik/tests"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

func (suite *KeeperTestSuite) TestSetIssuerDetails() {
	var (
		operator sdk.AccAddress
		issuer   sdk.AccAddress
		signer   sdk.AccAddress
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.MsgSetIssuerDetails
		expected func(resp *types.MsgSetIssuerDetailsResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg(
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
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "mismatch operator and signer",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg(
					operator.String(),
					"issuer address",
					"issuer name",
					"issuer description",
					"issuer url",
					"issuer logo",
					"issuer legal entity",
				)
				msg.Signer = signer.String()
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "invalid issuer",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg(
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
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg(
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
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgSetIssuerDetailsResponse{})

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
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)
				suite.Require().Equal(operator.String(), details.Operator)

				// Check if issuer's verification status is false
				verified, error := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			name: "existing issuer",
			init: func() {
				details := &types.IssuerDetails{Name: "test issuer", Operator: "operator"}
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg(
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
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
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
			resp, err := msgServer.HandleSetIssuerDetails(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateIssuerDetails() {
	var (
		issuer      sdk.AccAddress
		operator    sdk.AccAddress
		newOperator sdk.AccAddress
		signer      sdk.AccAddress
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
					"exiting operator",
					"new operator",
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
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer not exist",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
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
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "invalid issuer exist",
			init: func() {
				// Invalid issuer details were added in the store
				details := &types.IssuerDetails{Name: "test issuer"} // missing operator
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
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
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "mismatch operator and signer",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				// New signer, different from previous operator
				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				// existing operator (should pass operator, but passing signer) and new operator(no change operator)
				msg := types.NewUpdateIssuerDetailsMsg(
					signer.String(),
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
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
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
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)
				suite.Require().Equal(operator.String(), details.Operator)

				// Check if issuer was revoked
				verified, err := suite.keeper.IsAddressVerified(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().False(verified)
			},
		},
		{
			name: "new operator over existing",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				newOperator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg(
					operator.String(),
					newOperator.String(), // replace with new operator
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

				// Issuer details should exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().True(issuerExists)
				suite.Require().NoError(err)

				// Should be revoked verification if issuer address was verified
				addressDetails, err := suite.keeper.GetAddressDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(false, addressDetails.IsVerified)

				// Check if issuer details are stored correctly, especially new operator
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal("issuer name", details.Name)
				suite.Require().Equal("issuer description", details.Description)
				suite.Require().Equal("issuer url", details.Url)
				suite.Require().Equal("issuer logo", details.Logo)
				suite.Require().Equal("issuer legal entity", details.LegalEntity)
				suite.Require().Equal(newOperator.String(), details.Operator)
			},
		},
		{
			// Should revoke verification for all the accounts verified by updated issuer
			name: "past verification data still exists",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				_ = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())

				// Add address details with verification details
				_ = suite.keeper.AddVerificationDetails(
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
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "issuer not exist",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "invalid issuer exist",
			init: func() {
				// Invalid issuer details were added in the store
				details := &types.IssuerDetails{Name: "test issuer"} // missing operator
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "mismatch operator and signer",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				// New signer, different from previous operator
				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())
			},
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg(signer.String(), issuer.String())
				return &msg
			},
			expected: func(resp *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
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
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				_ = suite.keeper.SetAddressVerificationStatus(suite.ctx, issuer, true)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())

				// Add address details with verification details
				_ = suite.keeper.AddVerificationDetails(
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
