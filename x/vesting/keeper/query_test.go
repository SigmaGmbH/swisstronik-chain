package keeper_test

import (
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"swisstronik/tests"
	"swisstronik/x/vesting/types"
)

func (suite *KeeperTestSuite) TestBalances() {
	now := tmtime.Now()
	var (
		toAddress sdk.AccAddress
		coins     sdk.Coins
	)
	testCases := []struct {
		name     string
		init     func()
		malleate func() *types.QueryBalancesRequest
		expected func(resp *types.QueryBalancesResponse, error error)
	}{
		{
			name: "nil request",
			malleate: func() *types.QueryBalancesRequest {
				return nil
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().ErrorIs(error, status.Error(codes.InvalidArgument, "invalid request"))
				suite.Require().Nil(resp)
			},
		},
		{
			name: "empty request",
			malleate: func() *types.QueryBalancesRequest {
				return &types.QueryBalancesRequest{}
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().ErrorContains(error, "empty address string is not allowed")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "invalid address",
			malleate: func() *types.QueryBalancesRequest {
				return &types.QueryBalancesRequest{
					Address: "invalid address",
				}
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().ErrorContains(error, "decoding bech32")
				suite.Require().Nil(resp)
			},
		},
		{
			name: "unknown address",
			init: func() {
				toAddress = tests.RandomAccAddress()
			},
			malleate: func() *types.QueryBalancesRequest {
				return &types.QueryBalancesRequest{
					Address: toAddress.String(),
				}
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().ErrorIs(error, errortypes.ErrUnknownAddress)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "not found vesting account",
			init: func() {
				toAddress = tests.RandomAccAddress()

				// Set to account
				baseAccount := authtypes.NewBaseAccountWithAddress(toAddress)
				acc := suite.accountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.accountKeeper.SetAccount(suite.ctx, acc)
			},
			malleate: func() *types.QueryBalancesRequest {
				return &types.QueryBalancesRequest{
					Address: toAddress.String(),
				}
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().ErrorIs(error, types.ErrNotFoundVestingAccount)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "success",
			init: func() {
				toAddress = tests.RandomAccAddress()

				// Set to account
				baseAccount := authtypes.NewBaseAccountWithAddress(toAddress)
				acc := suite.accountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.accountKeeper.SetAccount(suite.ctx, acc)

				coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))

				vestingAccount := types.NewMonthlyVestingAccount(
					baseAccount,
					coins,
					now.Unix(),
					30,
					12,
				)
				suite.accountKeeper.SetAccount(suite.ctx, vestingAccount)
			},
			malleate: func() *types.QueryBalancesRequest {
				return &types.QueryBalancesRequest{
					Address: toAddress.String(),
				}
			},
			expected: func(resp *types.QueryBalancesResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp.Locked, coins)
				suite.Require().Equal(resp.Unvested, coins)
				suite.Require().Nil(resp.Vested)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.init != nil {
				tc.init()
			}

			req := tc.malleate()
			resp, err := suite.querier.Balances(suite.goCtx, req)
			tc.expected(resp, err)
		})
	}
}
