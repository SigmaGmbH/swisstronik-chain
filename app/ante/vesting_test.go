package ante_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	"swisstronik/tests"
	"swisstronik/utils"
	evmtypes "swisstronik/x/evm/types"
	vestingmoduletypes "swisstronik/x/vesting/types"
)

func (suite *AnteTestSuite) TestEthVestingDecorator() {
	const (
		cliffDays = 10
		months    = 3
	)
	var (
		userEth, vaEth         common.Address
		userPrivKey, vaPrivKey cryptotypes.PrivKey
		user, va               sdk.AccAddress
		vestingAccount         *vestingmoduletypes.MonthlyVestingAccount
		initialVesting         sdk.Coins
	)

	setup := func() {
		suite.enableFeemarket = false
		suite.SetupTest() // reset

		userEth, userPrivKey = tests.RandomEthAddressWithPrivateKey()
		user = sdk.AccAddress(userEth.Bytes())

		// Create regular account
		userAccount := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, user)
		suite.Require().NoError(userAccount.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, userAccount)

		vaEth, vaPrivKey = tests.RandomEthAddressWithPrivateKey()
		va = sdk.AccAddress(vaEth.Bytes())
		amount := math.NewInt(1e17).Mul(math.NewInt(months))
		initialVesting = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount))

		// Create only vesting account without funding
		baseAccount := authtypes.NewBaseAccountWithAddress(va)
		baseAccount = suite.app.AccountKeeper.NewAccount(suite.ctx, baseAccount).(*authtypes.BaseAccount)
		vestingAccount = vestingmoduletypes.NewMonthlyVestingAccount(
			baseAccount,
			initialVesting,
			suite.ctx.BlockTime().Unix(),
			cliffDays,
			months,
		)
		suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAccount)

		suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	}

	passBlockTime := func(duration time.Duration) {
		// NOTE, do not commit blocks and deliver transactions since it's only testing ante handler
		header := suite.ctx.BlockHeader()
		header.Time = header.Time.Add(duration)
		suite.ctx = suite.app.NewContext(false, header)
		suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(evmtypes.DefaultEVMDenom, sdk.OneInt())))
		suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000000000000000))
	}

	transferCoin := func(from common.Address, fromPrivKey cryptotypes.PrivKey, amount *big.Int) sdk.Tx {
		tx := evmtypes.NewTxFromArgs(
			&evmtypes.EvmTxArgs{
				ChainID:   suite.app.EvmKeeper.ChainID(),
				Nonce:     suite.app.EvmKeeper.GetNonce(suite.ctx, from),
				Amount:    amount,
				GasLimit:  100000,
				GasPrice:  big.NewInt(150),
				GasFeeCap: big.NewInt(200),
				To:        &userEth,
			}, nil, nil,
		)
		tx.From = from.Hex()
		return suite.CreateTestTx(tx, fromPrivKey, 1, false)
	}

	testCases := []struct {
		name          string
		expectedError error
		malleate      func() sdk.Tx
	}{
		{
			name:          "success - no vesting account with deploying contract",
			expectedError: nil,
			malleate: func() sdk.Tx {
				_ = suite.app.EvmKeeper.SetBalance(suite.ctx, userEth, big.NewInt(10000000000))
				contractTx := evmtypes.NewTxContract(
					suite.app.EvmKeeper.ChainID(),
					suite.app.EvmKeeper.GetNonce(suite.ctx, userEth),
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				contractTx.From = userEth.Hex()
				return suite.CreateTestTx(contractTx, userPrivKey, 1, false)
			},
		},
		{
			name:          "error - vesting account with balance 0",
			expectedError: errortypes.ErrInsufficientFunds,
			malleate: func() sdk.Tx {
				// still do not fund to keep account with zero balance
				// should not be able to make any transactions without balance
				return transferCoin(vaEth, vaPrivKey, big.NewInt(10))
			},
		},
		{
			name:          "error - vesting account with locked balance",
			expectedError: vestingmoduletypes.ErrInsufficientUnlockedCoins,
			malleate: func() sdk.Tx {
				// fund initial vesting amount
				_ = suite.app.EvmKeeper.SetBalance(suite.ctx, vaEth, initialVesting[0].Amount.BigInt())

				// try to transfer native token with locked balance
				// all the vesting amount were locked, nothing is spendable, should be failed
				amount := big.NewInt(10000000000)
				return transferCoin(vaEth, vaPrivKey, amount)
			},
		},
		{
			name:          "error - vesting account with insufficient vested coins",
			expectedError: vestingmoduletypes.ErrInsufficientUnlockedCoins,
			malleate: func() sdk.Tx {
				// fund initial vesting amount
				_ = suite.app.EvmKeeper.SetBalance(suite.ctx, vaEth, initialVesting[0].Amount.BigInt())

				// pass block time after the first vesting
				duration := time.Duration(vestingmoduletypes.SecondsOfDay*cliffDays) * time.Second
				duration = duration + time.Duration(vestingmoduletypes.SecondsOfMonth)*time.Second
				passBlockTime(duration)

				// try to transfer native token with coins more than vested
				// some of the vesting amount were unlocked, but insufficient to transfer, should be failed
				amount := initialVesting.QuoInt(sdk.NewInt(months)).MulInt(sdk.NewInt(2))
				return transferCoin(vaEth, vaPrivKey, amount[0].Amount.BigInt())
			},
		},
		{
			name:          "success - vesting account with sufficient vested coins",
			expectedError: nil,
			malleate: func() sdk.Tx {
				// fund initial vesting amount
				_ = suite.app.EvmKeeper.SetBalance(suite.ctx, vaEth, initialVesting[0].Amount.BigInt())

				// pass block time after the first vesting
				duration := time.Duration(vestingmoduletypes.SecondsOfDay*cliffDays) * time.Second
				duration = duration + time.Duration(vestingmoduletypes.SecondsOfMonth)*time.Second
				passBlockTime(duration)

				// try to transfer native token with vested coins, should be success
				amount := initialVesting.QuoInt(sdk.NewInt(months))
				return transferCoin(vaEth, vaPrivKey, amount[0].Amount.BigInt())
			},
		},
		{
			name:          "success - vesting account with all vested coins",
			expectedError: nil,
			malleate: func() sdk.Tx {
				// fund initial vesting amount
				_ = suite.app.EvmKeeper.SetBalance(suite.ctx, vaEth, initialVesting[0].Amount.BigInt())

				// pass block time after the end of vesting period
				duration := time.Duration(vestingmoduletypes.SecondsOfDay*cliffDays) * time.Second
				duration = duration + time.Duration(vestingmoduletypes.SecondsOfMonth*months)*time.Second
				passBlockTime(duration)

				// try to transfer native token with initial vesting, should be success
				amount := initialVesting
				return transferCoin(vaEth, vaPrivKey, amount[0].Amount.BigInt())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			// Check ante handler with generated transaction message
			var err error
			suite.ctx, err = suite.anteHandler(suite.ctx, tc.malleate(), false)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectedError)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
