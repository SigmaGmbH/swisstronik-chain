package keeper_test

import (
	"context"
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"swisstronik/tests"
	vestingkeeper "swisstronik/x/vesting/keeper"
	"swisstronik/x/vesting/testutil"
	"swisstronik/x/vesting/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	goCtx         context.Context
	accountKeeper *authkeeper.AccountKeeper
	bankKeeper    *testutil.MockBankKeeper
	keeper        *vestingkeeper.Keeper
	msgServer     types.MsgServer
	querier       vestingkeeper.Querier
}

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("swtr", "swtrpub")
}

func TestVestingAccountTestSuite(t *testing.T) {
	s := new(KeeperTestSuite)
	s.Setup(t)
	suite.Run(t, s)
}

func (suite *KeeperTestSuite) Setup(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	maccPerms := map[string][]string{}
	ak := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		storeKey,
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32PrefixAccAddr,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctrl := gomock.NewController(t)
	bk := testutil.NewMockBankKeeper(ctrl)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"VestingParams",
	)
	k := vestingkeeper.NewKeeper(
		storeKey,
		memStoreKey,
		paramsSubspace,
		ak,
		bk,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	suite.keeper = k
	suite.accountKeeper = &ak
	suite.bankKeeper = bk
	suite.ctx = ctx
	suite.goCtx = sdk.WrapSDKContext(suite.ctx)
	suite.msgServer = vestingkeeper.NewMsgServerImpl(*k)
	suite.querier = vestingkeeper.Querier{Keeper: *suite.keeper}
}

func (suite *KeeperTestSuite) TestCreateMonthlyVestingAccount() {
	var (
		fromAddress sdk.AccAddress
		toAddress   sdk.AccAddress
		coins       sdk.Coins
	)
	testCases := []struct {
		name     string
		init     func()
		expected func(resp *types.MsgCreateMonthlyVestingAccountResponse, error error)
	}{
		{
			name: "create for existing account",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				fromAddress = sdk.AccAddress(from.Bytes())
				to, _ := tests.RandomEthAddressWithPrivateKey()
				toAddress = sdk.AccAddress(to.Bytes())

				// Set to account
				baseAccount := authtypes.NewBaseAccountWithAddress(toAddress)
				acc := suite.accountKeeper.NewAccount(suite.ctx, baseAccount)
				suite.accountKeeper.SetAccount(suite.ctx, acc)

				coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
				suite.bankKeeper.EXPECT().IsSendEnabledCoins(suite.ctx, coins).Return(nil)
				suite.bankKeeper.EXPECT().BlockedAddr(toAddress).Return(false)
			},
			expected: func(resp *types.MsgCreateMonthlyVestingAccountResponse, error error) {
				suite.Require().ErrorIs(error, errortypes.ErrInvalidRequest)
				suite.Require().Nil(resp)
			},
		},
		{
			name: "create a valid periodic vesting account",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				fromAddress = sdk.AccAddress(from.Bytes())
				to, _ := tests.RandomEthAddressWithPrivateKey()
				toAddress = sdk.AccAddress(to.Bytes())

				coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
				suite.bankKeeper.EXPECT().IsSendEnabledCoins(suite.ctx, coins).Return(nil)
				suite.bankKeeper.EXPECT().BlockedAddr(toAddress).Return(false)
				suite.bankKeeper.EXPECT().SendCoins(suite.ctx, fromAddress, toAddress, coins).Return(nil)
			},
			expected: func(resp *types.MsgCreateMonthlyVestingAccountResponse, error error) {
				suite.Require().NoError(error)
				suite.Require().Equal(resp, &types.MsgCreateMonthlyVestingAccountResponse{})
			},
		},
		{
			name: "blocked account",
			init: func() {
				from, _ := tests.RandomEthAddressWithPrivateKey()
				fromAddress = sdk.AccAddress(from.Bytes())
				to, _ := tests.RandomEthAddressWithPrivateKey()
				toAddress = sdk.AccAddress(to.Bytes())

				coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
				suite.bankKeeper.EXPECT().IsSendEnabledCoins(suite.ctx, coins).Return(nil)
				suite.bankKeeper.EXPECT().BlockedAddr(toAddress).Return(true)
			},
			expected: func(resp *types.MsgCreateMonthlyVestingAccountResponse, error error) {
				suite.Require().ErrorIs(error, errortypes.ErrUnauthorized)
				suite.Require().Nil(resp)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.init != nil {
				tc.init()
			}

			msg := types.NewMsgCreateMonthlyVestingAccount(
				fromAddress.String(),
				toAddress.String(),
				30,
				12,
				coins,
			)
			resp, err := suite.msgServer.HandleCreateMonthlyVestingAccount(suite.goCtx, msg)
			tc.expected(resp, err)
		})
	}
}
