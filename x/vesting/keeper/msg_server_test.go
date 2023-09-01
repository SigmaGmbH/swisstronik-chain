package keeper_test

import (
	"testing"
	"time"

	"swisstronik/cmd/swisstronikd/cmd"
	"swisstronik/x/vesting/keeper"
	"swisstronik/x/vesting/types"

	"swisstronik/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"swisstronik/testutil"
)

var (
	fromAddr   = "swtr1qa2h6a27waactkrc6dyxrn2jzfjjfg24dgxzu8"
	to1Addr    = "swtr1ctvs7dql3e7pl7j4zwuck7n8jc3vrh5kr6ng8g"
	to1AddrAcc = sdk.AccAddress([]byte(to1Addr))

	to2Addr    = "swtr13gvyhac4qwtqjkdpzxzarlvpsz3zxld5v2ae58"
	zeroCoin   = sdk.NewInt64Coin("uswtr", 0)
	periodCoin = sdk.NewInt64Coin("uswtr", 200000000)
)

func TestCreatingMonthlyVestingAccount(t *testing.T) {
	cmd.InitSDKConfig()

	// setup the app
	app, genAcc := simapp.Setup(t, false)
	
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "swisstronik_1291-1"})
	msgServer := keeper.NewMsgServerImpl(app.VestingKeeper)

	toAcc := app.AccountKeeper.NewAccountWithAddress(ctx, to1AddrAcc)
	app.AccountKeeper.SetAccount(ctx, toAcc)

	existingAddr := genAcc.GetAddress().String()
	toAddr := toAcc.GetAddress().String()

	// prefund account
	coinsToMint := sdk.NewCoins(periodCoin)
	testutil.FundAccount(app.BankKeeper, ctx, genAcc.GetAddress(), coinsToMint)

	testCases := []struct {
		name      string
		preRun    func()
		input     *types.MsgCreateMonthlyVestingAccount
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty from address",
			input: types.NewMsgCreateMonthlyVestingAccount(
				"",
				to1Addr,
				time.Now().Unix(),
				sdk.NewCoins(periodCoin),
				10,
			),
			expErr:    true,
			expErrMsg: "invalid 'from' address",
		},
		{
			name: "empty to address",
			input: types.NewMsgCreateMonthlyVestingAccount(
				fromAddr,
				"",
				time.Now().Unix(),
				sdk.NewCoins(periodCoin),
				10,
			),
			expErr:    true,
			expErrMsg: "invalid 'to' address",
		},
		{
			name: "invalid start time",
			input: types.NewMsgCreateMonthlyVestingAccount(
				fromAddr,
				to1Addr,
				0,
				sdk.NewCoins(periodCoin),
				10,
			),
			expErr:    true,
			expErrMsg: "invalid start time",
		},
		{
			name: "invalid months",
			input: types.NewMsgCreateMonthlyVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				sdk.NewCoins(periodCoin),
				0,
			),
			expErr:    true,
			expErrMsg: "invalid months",
		},
		{
			name: "invalid amount",
			input: types.NewMsgCreateMonthlyVestingAccount(
				fromAddr,
				to1Addr,
				time.Now().Unix(),
				sdk.NewCoins(zeroCoin),
				10,
			),
			expErr:    true,
			expErrMsg: "invalid amount",
		},
		{
			name: "create for existing account",
			preRun: func() {

			},
			input: types.NewMsgCreateMonthlyVestingAccount(
				existingAddr,
				toAddr,
				time.Now().Unix(),
				sdk.NewCoins(periodCoin),
				10,
			),
			expErr:    true,
			expErrMsg: "already exists",
		},
		{
			name: "create a valid monthly vesting account",
			input: types.NewMsgCreateMonthlyVestingAccount(
				existingAddr,
				to2Addr,
				time.Now().Unix(),
				sdk.NewCoins(periodCoin),
				10,
			),
			expErr:    false,
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CreateMonthlyVestingAccount(ctx, tc.input)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
