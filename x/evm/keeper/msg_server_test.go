package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"swisstronik/x/evm/types"
)

func (suite *KeeperTestSuite) TestEthereumTx() {
	var (
		err             error
		msg             *types.MsgHandleTx
		signer          ethtypes.Signer
		chainCfg        *params.ChainConfig
		expectedGasUsed uint64
	)

	testCases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		{
			"Deploy contract tx",
			func() {
				msg, err = suite.createContractMsgTx(
					suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
					signer,
					chainCfg,
					big.NewInt(1),
				)
				suite.Require().NoError(err)
				expectedGasUsed = params.TxGasContractCreation
			},
			false,
		},
		{
			"Transfer funds tx",
			func() {
				msg, _, err = newEthMsgTx(
					suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
					suite.ctx.BlockHeight(),
					suite.address,
					chainCfg,
					suite.signer,
					signer,
					ethtypes.AccessListTxType,
					nil,
					nil,
					big.NewInt(0),
				)
				suite.Require().NoError(err)
				expectedGasUsed = params.TxGas
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			chainCfg = keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			signer = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

			tc.malleate()
			res, err := suite.app.EvmKeeper.HandleTx(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(expectedGasUsed, res.GasUsed)
			suite.Require().False(res.Failed())
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name          string
		request       *types.MsgUpdateParams
		expectErr     bool
		compareParams bool
	}{
		{
			name:          "fail - invalid authority",
			request:       &types.MsgUpdateParams{Authority: "foobar"},
			expectErr:     true,
			compareParams: false,
		},
		{
			name: "pass - valid Update msg",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			expectErr:     false,
			compareParams: false,
		},
		{
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					EvmDenom:            "newdenom",
					EnableCreate:        false,
					EnableCall:          false,
					ChainConfig:         types.DefaultChainConfig(),
					ExtraEIPs:           nil,
					AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
				},
			},
			expectErr:     false,
			compareParams: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run("MsgUpdateParams", func() {
			paramsBefore := suite.app.EvmKeeper.GetParams(suite.ctx)

			_, err := suite.app.EvmKeeper.UpdateParams(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			if tc.compareParams {
				paramsAfter := suite.app.EvmKeeper.GetParams(suite.ctx)
				suite.Require().Equal(paramsBefore.EvmDenom, paramsAfter.EvmDenom)
				suite.Require().NotEqual(paramsBefore.EnableCreate, paramsAfter.EnableCreate)
				suite.Require().NotEqual(paramsBefore.EnableCall, paramsAfter.EnableCall)
			}
		})
	}
}
