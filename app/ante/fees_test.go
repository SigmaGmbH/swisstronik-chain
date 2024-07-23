package ante_test

import (
	"math/big"

	"swisstronik/app/ante"
	"swisstronik/tests"
	evmtypes "swisstronik/x/evm/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

var execTypes = []struct {
	name      string
	isCheckTx bool
	simulate  bool
}{
	{"deliverTx", false, false},
	{"deliverTxSimulate", false, true},
}

func (suite *AnteTestSuite) TestMinGasPriceDecorator() {
	denom := evmtypes.DefaultEVMDenom
	testMsg := banktypes.MsgSend{
		FromAddress: "swtr13sllcdsqhjektac5r6h50dvjrthm0yt6zw3q4s",
		ToAddress:   "swtr1734tyvkylw3f7vc9xmwxp6g5n79qvsrvjhsvs4",
		Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
	}

	testCases := []struct {
		name                string
		malleate            func() sdk.Tx
		expPass             bool
		errMsg              string
		allowPassOnSimulate bool
	}{
		{
			"invalid cosmos tx type",
			func() sdk.Tx {
				return &invalidTx{}
			},
			false,
			"invalid transaction type",
			false,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			false,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			false,
		},
		{
			"valid cosmos tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			false,
		},
		{
			"invalid cosmos tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			true,
		},
		{
			"invalid cosmos tx with wrong denom",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), "stake", &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			true,
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			suite.Run(et.name+"_"+tc.name, func() {
				// suite.SetupTest(et.isCheckTx)
				ctx := suite.ctx.WithIsReCheckTx(et.isCheckTx)
				dec := ante.NewMinGasPriceDecorator(suite.app.FeeMarketKeeper, suite.app.EvmKeeper)
				_, err := dec.AnteHandle(ctx, tc.malleate(), et.simulate, NextFn)

				if tc.expPass || (et.simulate && tc.allowPassOnSimulate) {
					suite.Require().NoError(err, tc.name)
				} else {
					suite.Require().Error(err, tc.name)
					suite.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}

func (suite *AnteTestSuite) TestEthMinGasPriceDecorator() {
	denom := evmtypes.DefaultEVMDenom
	from, privKey := tests.RandomEthAddressWithPrivateKey()
	to := tests.RandomEthAddress()
	emptyAccessList := ethtypes.AccessList{}

	testCases := []struct {
		name     string
		malleate func() sdk.Tx
		expPass  bool
		errMsg   string
	}{
		{
			"invalid tx type",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				return &invalidTx{}
			},
			false,
			"invalid message type",
		},
		{
			"wrong tx type",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				testMsg := banktypes.MsgSend{
					FromAddress: "swtr13sllcdsqhjektac5r6h50dvjrthm0yt6zw3q4s",
					ToAddress:   "swtr1734tyvkylw3f7vc9xmwxp6g5n79qvsrvjhsvs4",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"invalid message type",
		},
		{
			"valid: invalid tx type with MinGasPrices = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				return &invalidTx{}
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 0, gasPrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(10), nil, nil, nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(10), nil, nil, nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"invalid legacy tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices = 0, EffectivePrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(0), big.NewInt(0), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid dynamic tx with MinGasPrices = 0, EffectivePrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(100), big.NewInt(50), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid dynamic tx with MinGasPrices < EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(100), big.NewInt(100), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"invalid dynamic tx with MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(0), big.NewInt(0), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"invalid dynamic tx with MinGasPrices > BaseFee, MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				feemarketParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				feemarketParams.BaseFee = sdkmath.NewInt(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, feemarketParams)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(1000), big.NewInt(0), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices > BaseFee, MinGasPrices < EffectivePrice (big GasTipCap)",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

				feemarketParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				feemarketParams.BaseFee = sdkmath.NewInt(10)
				_ = suite.app.FeeMarketKeeper.SetParams(suite.ctx, feemarketParams)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(1000), big.NewInt(101), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			suite.Run(et.name+"_"+tc.name, func() {
				// suite.SetupTest(et.isCheckTx)
				suite.SetupTest()
				dec := ante.NewEthMinGasPriceDecorator(suite.app.FeeMarketKeeper, suite.app.EvmKeeper)
				_, err := dec.AnteHandle(suite.ctx, tc.malleate(), et.simulate, NextFn)

				if tc.expPass {
					suite.Require().NoError(err, tc.name)
				} else {
					suite.Require().Error(err, tc.name)
					suite.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}

func (suite *AnteTestSuite) TestEthMempoolFeeDecorator() {
	// TODO: add test
}
