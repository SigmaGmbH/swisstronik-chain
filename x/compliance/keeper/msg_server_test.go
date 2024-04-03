package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		malleate func() *types.MsgSetIssuerDetails
		expected func(resp *types.MsgSetIssuerDetailsResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgSetIssuerDetails {
				msg := types.NewSetIssuerDetailsMsg("operator address", "issuer address", "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
			},
		},
		{
			name: "mismatch operator and signer",
			malleate: func() *types.MsgSetIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())

				msg := types.NewSetIssuerDetailsMsg(operator.String(), "issuer address", "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				msg.Signer = signer.String()
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
			},
		},
		{
			name: "invalid issuer",
			malleate: func() *types.MsgSetIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				msg := types.NewSetIssuerDetailsMsg(operator.String(), "issuer address", "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
			},
		},
		{
			name: "success",
			malleate: func() *types.MsgSetIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())

				msg := types.NewSetIssuerDetailsMsg(operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().NoError(error)

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
			},
		},
		{
			name: "existing issuer",
			malleate: func() *types.MsgSetIssuerDetails {
				details := &types.IssuerDetails{Name: "test issuer", Operator: "operator"}
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())
				msg := types.NewSetIssuerDetailsMsg(operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(resp *types.MsgSetIssuerDetailsResponse, error error) {
				suite.Require().Error(error)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
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
		malleate func() *types.MsgUpdateIssuerDetails
		expected func(response *types.MsgUpdateIssuerDetailsResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgUpdateIssuerDetails {
				msg := types.NewUpdateIssuerDetailsMsg("exiting operator", "new operator", "issuer address", "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "issuer not exist",
			malleate: func() *types.MsgUpdateIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				msg := types.NewUpdateIssuerDetailsMsg(operator.String(), operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "invalid issuer exist",
			malleate: func() *types.MsgUpdateIssuerDetails {
				// Invalid issuer details were added in the store
				details := &types.IssuerDetails{Name: "test issuer"} // missing operator
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				msg := types.NewUpdateIssuerDetailsMsg(operator.String(), operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "mismatch operator and signer",
			malleate: func() *types.MsgUpdateIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				// New signer, different from previous operator
				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())

				// existing operator (should pass operator, but passing signer) and new operator(no change operator)
				msg := types.NewUpdateIssuerDetailsMsg(signer.String(), operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "success",
			malleate: func() *types.MsgUpdateIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				msg := types.NewUpdateIssuerDetailsMsg(operator.String(), operator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().NoError(err)

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
			},
		},
		{
			name: "new operator over existing",
			malleate: func() *types.MsgUpdateIssuerDetails {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				newOperator = sdk.AccAddress(from.Bytes())

				msg := types.NewUpdateIssuerDetailsMsg(operator.String(), newOperator.String(), issuer.String(), "issuer name", "issuer description", "issuer url", "issuer logo", "issuer legal entity")
				return &msg
			},
			expected: func(_ *types.MsgUpdateIssuerDetailsResponse, err error) {
				suite.Require().NoError(err)

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
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
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
		malleate func() *types.MsgRemoveIssuer
		expected func(resp *types.MsgRemoveIssuerResponse, error error)
	}{
		{
			name: "invalid fields",
			malleate: func() *types.MsgRemoveIssuer {
				msg := types.NewRemoveIssuerMsg("operator", "issuer address")
				return &msg
			},
			expected: func(_ *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "issuer not exist",
			malleate: func() *types.MsgRemoveIssuer {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(_ *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "invalid issuer exist",
			malleate: func() *types.MsgRemoveIssuer {
				// Invalid issuer details were added in the store
				details := &types.IssuerDetails{Name: "test issuer"} // missing operator
				from, _ := tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				from, _ = tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(_ *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "mismatch operator and signer",
			malleate: func() *types.MsgRemoveIssuer {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				// New signer, different from previous operator
				from, _ = tests.RandomEthAddressWithPrivateKey()
				signer = sdk.AccAddress(from.Bytes())

				msg := types.NewRemoveIssuerMsg(signer.String(), issuer.String())
				return &msg
			},
			expected: func(_ *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().Error(err)
			},
		},
		{
			name: "success",
			malleate: func() *types.MsgRemoveIssuer {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				operator = sdk.AccAddress(from.Bytes())

				from, _ = tests.RandomEthAddressWithPrivateKey()
				issuer = sdk.AccAddress(from.Bytes())
				details := &types.IssuerDetails{Name: "test issuer", Operator: operator.String()}
				_ = suite.keeper.SetIssuerDetails(suite.ctx, issuer, details)

				msg := types.NewRemoveIssuerMsg(operator.String(), issuer.String())
				return &msg
			},
			expected: func(_ *types.MsgRemoveIssuerResponse, err error) {
				suite.Require().NoError(err)

				// Issuer should not exist
				issuerExists, err := suite.keeper.IssuerExists(suite.ctx, issuer)
				suite.Require().False(issuerExists)
				suite.Require().NoError(err)

				// Same for issuer details
				details, err := suite.keeper.GetIssuerDetails(suite.ctx, issuer)
				suite.Require().NoError(err)
				suite.Require().Equal(details, &types.IssuerDetails{})
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msgServer := keeper.NewMsgServerImpl(suite.keeper)
			msg := tc.malleate()
			resp, err := msgServer.HandleRemoveIssuer(sdk.WrapSDKContext(suite.ctx), msg)
			tc.expected(resp, err)
		})
	}
}
