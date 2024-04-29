package types_test

import (
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"
	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/tests"
	"swisstronik/x/vesting/types"
	"testing"
	"time"
)

type VestingAccountTestSuite struct {
	suite.Suite
}

func TestVestingAccountTestSuite(t *testing.T) {
	suite.Run(t, new(VestingAccountTestSuite))
}

func (suite *VestingAccountTestSuite) TestMonthlyVestingAccountValidate() {
	now := tmtime.Now()

	privkey, _ := ethsecp256k1.GenerateKey()
	addr := sdk.AccAddress(privkey.PubKey().Address())
	privkey2, _ := ethsecp256k1.GenerateKey()

	baseAcc := authtypes.NewBaseAccountWithAddress(addr)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))

	testCases := []struct {
		name   string
		acc    authtypes.GenesisAccount
		expErr bool
	}{
		{
			"valid base account",
			baseAcc,
			false,
		},
		{
			"invalid base account",
			types.NewMonthlyVestingAccount(
				authtypes.NewBaseAccount(addr, privkey2.PubKey(), 0, 0),
				initialVesting,
				now.Unix(),
				30,
				12,
			),
			true,
		},
		{
			"valid monthly vesting account",
			types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 30, 12),
			false,
		},
		{
			name: "empty period with base vesting account",
			acc: types.NewMonthlyVestingAccountRaw(
				sdkvesting.NewBaseVestingAccount(
					authtypes.NewBaseAccountWithAddress(addr),
					initialVesting,
					now.Add(time.Hour*24).Unix(),
				),
				now.Unix(),
				now.Add(time.Hour*24).Unix(),
				sdkvesting.Periods{},
			),
			expErr: true,
		},
		{
			name: "one period with base vesting account",
			acc: types.NewMonthlyVestingAccountRaw(
				sdkvesting.NewBaseVestingAccount(
					authtypes.NewBaseAccountWithAddress(addr),
					initialVesting,
					now.Add(time.Hour*24*2).Unix(), // end time = start time + cliff days
				),
				now.Unix(),
				now.Add(time.Hour*24).Unix(),
				sdkvesting.Periods{
					sdkvesting.Period{
						Length: int64((time.Hour * 24).Seconds()),
						Amount: initialVesting,
					},
				},
			),
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(tc.expErr, tc.acc.Validate() != nil)
		})
	}
}

func (suite *VestingAccountTestSuite) TestGetVestedCoinsMonthlyVestingAcc() {
	now := tmtime.Now()
	cliffTime := now.Add(time.Hour * 24 * 10) // after 10 days
	firstMonth := now.Add(time.Hour * 24 * 30)
	firstMonthAfterCliff := cliffTime.Add(time.Hour * 24 * 30)             // after 1 month
	secondMonthAfterCliff := firstMonthAfterCliff.Add(time.Hour * 24 * 30) // after 2 months
	endTime := cliffTime.Add(time.Hour * 24 * 30 * 3)                      // after 3 months

	baseAcc, initialVesting := initBaseAccount(300)
	mva := types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 10, 3)

	testCases := []struct {
		name        string
		blockTime   time.Time
		vestedCoins sdk.Coins
	}{
		{
			"require no coins vested at the beginning",
			now,
			nil,
		},
		{
			"require no coins vested before cliff",
			cliffTime.Add(-time.Second),
			nil,
		},
		{
			"require all coins vested after the end of vesting period",
			endTime.Add(time.Second),
			initialVesting,
		},
		{
			"require no coins vested during the first month from start time",
			firstMonth,
			nil,
		},
		{
			"require no coins vested during the first month after cliff time",
			firstMonthAfterCliff.Add(-time.Second),
			nil,
		},
		{
			"require 1/3 coins vested after first month after cliff time",
			firstMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)),
		},
		{
			"require 1/3 coins vested after first month after cliff time",
			firstMonthAfterCliff.Add(time.Second),
			initialVesting.QuoInt(sdk.NewInt(3)),
		},
		{
			"require 2/3 coins vested after second months after cliff time",
			secondMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
		},
		{
			"require 2/3 coins vested before end of vesting period",
			endTime.Add(-time.Second),
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
		},
		{
			"require all coins vested at the end of vesting period",
			endTime,
			initialVesting,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vestedCoins := mva.GetVestedCoins(tc.blockTime)
			suite.Require().Equal(tc.vestedCoins, vestedCoins)
		})
	}
}

func (suite *VestingAccountTestSuite) TestGetVestingMonthlyVestingAcc() {
	now := tmtime.Now()

	cliffTime := now.Add(time.Hour * 24 * 10) // after 10 days
	firstMonth := now.Add(time.Hour * 24 * 30)
	firstMonthAfterCliff := cliffTime.Add(time.Hour * 24 * 30)             // after 1 month
	secondMonthAfterCliff := firstMonthAfterCliff.Add(time.Hour * 24 * 30) // after 2 months
	endTime := cliffTime.Add(time.Hour * 24 * 30 * 3)                      // after 3 months

	baseAcc, initialVesting := initBaseAccount(300)
	mva := types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 10, 3)

	testCases := []struct {
		name         string
		blockTime    time.Time
		vestingCoins sdk.Coins
	}{
		{
			"require all coins vesting at the beginning",
			now,
			initialVesting,
		},
		{
			"require all coins vesting before cliff",
			cliffTime.Add(-time.Second),
			initialVesting,
		},
		{
			"require no coins after the end of vesting period",
			endTime.Add(time.Second),
			sdk.Coins{},
		},
		{
			"require all coins vested during the first month from start time",
			firstMonth,
			initialVesting,
		},
		{
			"require no coins vested during the first month after cliff time",
			firstMonthAfterCliff.Add(-time.Second),
			initialVesting,
		},
		{
			"require 2/3 coins vested after first month after cliff time",
			firstMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
		},
		{
			"require 2/3 coins vested after first month after cliff time",
			firstMonthAfterCliff.Add(time.Second),
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
		},
		{
			"require 1/3 coins vested after second month after cliff time",
			secondMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)),
		},
		{
			"require 1/3 coins vested before end of vesting period",
			endTime.Add(-time.Second),
			initialVesting.QuoInt(sdk.NewInt(3)),
		},
		{
			"require no coins vested at the end of vesting period",
			endTime,
			sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vestingCoins := mva.GetVestingCoins(tc.blockTime)
			suite.Require().Equal(tc.vestingCoins, vestingCoins)
		})
	}
}

func (suite *VestingAccountTestSuite) TestSpendableCoinsMonthlyVestingAcc() {
	now := tmtime.Now()

	cliffTime := now.Add(time.Hour * 24 * 10)                  // after 10 days
	firstMonthAfterCliff := cliffTime.Add(time.Hour * 24 * 30) // after 1 month
	endTime := cliffTime.Add(time.Hour * 24 * 30 * 3)          // after 3 months

	baseAcc, initialVesting := initBaseAccount(300)
	mva := types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 10, 3)

	testCases := []struct {
		name        string
		blockTime   time.Time
		lockedCoins sdk.Coins
	}{
		{
			"require all coins locked at the beginning",
			now,
			initialVesting,
		},
		{
			"require no coins locked at the end",
			endTime,
			sdk.Coins{},
		},
		{
			"require 2/3 coins locked after first month after cliff time",
			firstMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lockedCoins := mva.LockedCoins(tc.blockTime)
			suite.Require().Equal(tc.lockedCoins, lockedCoins)
		})
	}
}

func (suite *VestingAccountTestSuite) TestTrackDelegationMonthlyVestingAcc() {
	now := tmtime.Now()

	cliffTime := now.Add(time.Hour * 24 * 10)                  // after 10 days
	firstMonthAfterCliff := cliffTime.Add(time.Hour * 24 * 30) // after 1 month
	endTime := cliffTime.Add(time.Hour * 24 * 30 * 3)          // after 3 months

	baseAcc, initialVesting := initBaseAccount(300)

	testCases := []struct {
		name               string
		blockTime          time.Time
		delegationAmount   sdk.Coins
		expDelegatedAmount sdk.Coins
		expDelegatedFree   sdk.Coins
	}{
		{
			"require the ability to delegate all the vesting coins at the beginning",
			now,
			initialVesting,
			initialVesting,
			sdk.Coins{},
		},
		{
			"require no ability to delegate all the vesting coins at the end",
			endTime,
			initialVesting,
			sdk.Coins{},
			initialVesting,
		},
		{
			"require the ability to delegate only 2/3 coins after first month after cliff time",
			firstMonthAfterCliff,
			initialVesting,
			initialVesting.QuoInt(sdk.NewInt(3)).MulInt(sdk.NewInt(2)),
			initialVesting.QuoInt(sdk.NewInt(3)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mva := types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 10, 3)
			mva.TrackDelegation(tc.blockTime, initialVesting, tc.delegationAmount)
			suite.Require().Equal(tc.expDelegatedAmount, mva.DelegatedVesting)
			suite.Require().Equal(tc.expDelegatedFree, mva.DelegatedFree)
		})
	}
}

func (suite *VestingAccountTestSuite) TestTrackUndelegationMonthlyVestingAcc() {
	now := tmtime.Now()

	cliffTime := now.Add(time.Hour * 24 * 10)                  // after 10 days
	firstMonthAfterCliff := cliffTime.Add(time.Hour * 24 * 30) // after 1 month
	endTime := cliffTime.Add(time.Hour * 24 * 30 * 3)          // after 3 months

	baseAcc, initialVesting := initBaseAccount(300)

	testCases := []struct {
		name               string
		blockTime          time.Time
		delegationAmount   sdk.Coins
		undelegationAmount sdk.Coins
		expDelegatedAmount sdk.Coins
		expDelegatedFree   sdk.Coins
	}{
		{
			"require the ability to undelegate all the vesting coins at the beginning",
			now,
			initialVesting,
			initialVesting,
			sdk.Coins{},
			sdk.Coins{},
		},
		{
			"require no ability to undelegate all the vesting coins at the end",
			endTime,
			initialVesting,
			initialVesting,
			sdk.Coins{},
			sdk.Coins{},
		},
		{
			"require the ability to undelegate only 2/3 coins after first month after cliff time",
			firstMonthAfterCliff,
			initialVesting.QuoInt(sdk.NewInt(3)),
			initialVesting.QuoInt(sdk.NewInt(3)),
			sdk.Coins{},
			sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mva := types.NewMonthlyVestingAccount(baseAcc, initialVesting, now.Unix(), 10, 3)
			mva.TrackDelegation(tc.blockTime, initialVesting, tc.delegationAmount)
			mva.TrackUndelegation(tc.undelegationAmount)
			suite.Require().Equal(tc.expDelegatedAmount, mva.DelegatedVesting)
			suite.Require().Equal(tc.expDelegatedFree, mva.DelegatedFree)
		})
	}
}

func initBaseAccount(amount int64) (*authtypes.BaseAccount, sdk.Coins) {
	from, _ := tests.RandomEthAddressWithPrivateKey()
	addr := sdk.AccAddress(from.Bytes())
	baseAcc := authtypes.NewBaseAccountWithAddress(addr)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, amount))
	return baseAcc, initialVesting
}
