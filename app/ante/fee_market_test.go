package ante_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"swisstronik/app/ante"
	"swisstronik/tests"
	"swisstronik/types"
	evmtypes "swisstronik/x/evm/types"
)

func (suite *AnteTestSuite) TestGasWantedDecorator() {
	suite.enableFeemarket = true
	suite.SetupTest()
	dec := ante.NewGasWantedDecorator(suite.app.EvmKeeper, suite.app.FeeMarketKeeper)
	from, fromPrivKey := tests.RandomEthAddressWithPrivateKey()
	to := tests.RandomEthAddress()

	testCases := []struct {
		name              string
		expectedGasWanted uint64
		malleate          func() sdk.Tx
		expPass           bool
	}{
		{
			"Cosmos Tx",
			TestGasLimit,
			func() sdk.Tx {
				denom := evmtypes.DefaultEVMDenom
				testMsg := banktypes.MsgSend{
					FromAddress: "swtr1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "swtr1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), "stake", &testMsg)
				return txBuilder.GetTx()
			},
			true,
		},
		{
			"Ethereum Legacy Tx",
			TestGasLimit,
			func() sdk.Tx {
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil, nil, nil)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"Ethereum Access List Tx",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"Ethereum Dynamic Fee Tx (EIP1559)",
			TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), big.NewInt(100), big.NewInt(50), &emptyAccessList, nil, nil)
				return suite.CreateTestTx(msg, fromPrivKey, 1, false)
			},
			true,
		},
		{
			"Cosmos Tx - gasWanted > max block gas",
			TestGasLimit,
			func() sdk.Tx {
				denom := evmtypes.DefaultEVMDenom
				testMsg := banktypes.MsgSend{
					FromAddress: "swtr1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "swtr1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				limit := types.BlockGasLimit(suite.ctx)
				txBuilder.SetGasLimit(limit + 5)
				return txBuilder.GetTx()
			},
			false,
		},
	}

	// cumulative gas wanted from all test transactions in the same block
	var expectedGasWanted uint64

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := dec.AnteHandle(suite.ctx, tc.malleate(), false, NextFn)
			if tc.expPass {
				suite.Require().NoError(err)

				gasWanted := suite.app.FeeMarketKeeper.GetTransientGasWanted(suite.ctx)
				expectedGasWanted += tc.expectedGasWanted
				suite.Require().Equal(expectedGasWanted, gasWanted)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
