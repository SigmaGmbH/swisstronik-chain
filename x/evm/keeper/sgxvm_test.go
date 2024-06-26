package keeper_test

import (
	"encoding/json"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"swisstronik/server/config"
	"swisstronik/x/evm/keeper"
	"swisstronik/x/evm/types"
)

func (suite *KeeperTestSuite) TestNativeCurrencyTransfer() {
	var (
		err             error
		msg             *types.MsgHandleTx
		signer          ethtypes.Signer
		chainCfg        *params.ChainConfig
		expectedGasUsed uint64
		transferAmount  int64
	)

	testCases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		{
			"Transfer funds tx",
			func() {
				transferAmount = 1000
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
					big.NewInt(transferAmount),
				)
				suite.Require().NoError(err)
				expectedGasUsed = params.TxGas
			},
			false,
		},
		{
			"Exceeding balance transfer tx",
			func() {
				transferAmount = 1000
				wrongAmount := int64(100000)
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
					big.NewInt(wrongAmount),
				)
				suite.Require().NoError(err)
				expectedGasUsed = params.TxGas
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupSGXVMTest()

			keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			chainCfg = keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			signer = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

			tc.malleate()

			err := suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(transferAmount))
			suite.Require().NoError(err)

			balanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
			receiverBalanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})

			res, err := suite.app.EvmKeeper.HandleTx(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Equal(res.VmError, "evm error: OutOfFund")
				suite.Require().NoError(err)
				return
			} else {
				// Check sender's balance
				expectedBalance := balanceBefore.Sub(balanceBefore, big.NewInt(transferAmount))
				balanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
				isSenderBalanceCorrect := expectedBalance.Cmp(balanceAfter)
				suite.Require().True(isSenderBalanceCorrect == 0, "Incorrect sender's balance")

				// Check receiver's balance
				receiverBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})
				expectedReceiverBalance := receiverBalanceBefore.Add(receiverBalanceBefore, big.NewInt(transferAmount))
				isReceiverBalanceCorrect := expectedReceiverBalance.Cmp(receiverBalanceAfter)
				suite.Require().True(isReceiverBalanceCorrect == 0, "Incorrect receiver's balance")

				suite.Require().NoError(err)
				suite.Require().Equal(expectedGasUsed, res.GasUsed)
				suite.Require().False(res.Failed())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDryRun() {
	var (
		signer   ethtypes.Signer
		chainCfg *params.ChainConfig
	)

	amountToTransfer := int64(100)

	testCases := []struct {
		name   string
		commit bool
	}{
		{
			"Transfer in normal mode should update nonce and balance",
			true,
		},
		{
			"Transfer in dry mode should not update nonce and balance",
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupSGXVMTest()

			keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			chainCfg = keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			signer = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

			err := suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(amountToTransfer))
			suite.Require().NoError(err)

			cfg, err := suite.app.EvmKeeper.EVMConfig(suite.ctx, suite.ctx.BlockHeader().ProposerAddress, suite.app.EvmKeeper.ChainID())
			suite.Require().NoError(err)

			nonceBefore := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
			msg, baseFee, err := newEthMsgTx(
				nonceBefore,
				suite.ctx.BlockHeight(),
				suite.address,
				chainCfg,
				suite.signer,
				signer,
				ethtypes.AccessListTxType,
				nil,
				nil,
				big.NewInt(amountToTransfer),
			)
			suite.Require().NoError(err)

			tx := msg.AsTransaction()
			ethMessage, err := tx.AsMessage(signer, baseFee)
			suite.Require().NoError(err)

			txConfig := suite.app.EvmKeeper.TxConfig(suite.ctx, tx.Hash())
			txContext, err := keeper.CreateSGXVMContext(suite.ctx, suite.app.EvmKeeper, tx)
			suite.Require().NoError(err)

			balanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
			receiverBalanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})

			res, err := suite.app.EvmKeeper.ApplyMessageWithConfig(suite.ctx, ethMessage, tc.commit, cfg, txConfig, txContext, false)
			suite.Require().NoError(err)
			suite.Require().Empty(res.VmError)

			nonceAfter := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

			if tc.commit {
				// Check if balance & nonce were updated
				expectedBalance := balanceBefore.Sub(balanceBefore, big.NewInt(amountToTransfer))
				balanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)

				isSenderBalanceCorrect := expectedBalance.Cmp(balanceAfter)
				suite.Require().True(isSenderBalanceCorrect == 0, "Incorrect sender's balance")

				// Check receiver's balance
				receiverBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})
				expectedReceiverBalance := receiverBalanceBefore.Add(receiverBalanceBefore, big.NewInt(amountToTransfer))
				isReceiverBalanceCorrect := expectedReceiverBalance.Cmp(receiverBalanceAfter)
				suite.Require().True(isReceiverBalanceCorrect == 0, "Incorrect receiver's balance")

				// Check if nonce was updated
				suite.Require().Equal(nonceBefore+1, nonceAfter)
			} else {
				// Check if balance & nonce still the same
				// Check sender's balance
				balanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
				suite.Require().Equal(balanceBefore, balanceAfter)

				// Check receiver's balance
				receiverBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})
				suite.Require().Equal(receiverBalanceBefore, receiverBalanceAfter)

				// Check if nonce still the same
				suite.Require().Equal(nonceBefore, nonceAfter)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMultipleTransfers() {
	balanceToSet := int64(10)
	amountToTransfer := int64(1)

	suite.SetupSGXVMTest()

	keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	chainCfg := keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	err := suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(balanceToSet))
	suite.Require().NoError(err)

	cfg, err := suite.app.EvmKeeper.EVMConfig(suite.ctx, suite.ctx.BlockHeader().ProposerAddress, suite.app.EvmKeeper.ChainID())
	suite.Require().NoError(err)

	for i := 0; i < 10; i++ {
		nonceBefore := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
		msg, baseFee, err := newEthMsgTx(
			nonceBefore,
			suite.ctx.BlockHeight(),
			suite.address,
			chainCfg,
			suite.signer,
			signer,
			ethtypes.AccessListTxType,
			nil,
			nil,
			big.NewInt(amountToTransfer),
		)
		suite.Require().NoError(err)

		tx := msg.AsTransaction()
		ethMessage, err := tx.AsMessage(signer, baseFee)
		suite.Require().NoError(err)

		txConfig := suite.app.EvmKeeper.TxConfig(suite.ctx, tx.Hash())
		txContext, err := keeper.CreateSGXVMContext(suite.ctx, suite.app.EvmKeeper, tx)
		suite.Require().NoError(err)

		balanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)
		receiverBalanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})

		res, err := suite.app.EvmKeeper.ApplyMessageWithConfig(suite.ctx, ethMessage, true, cfg, txConfig, txContext, false)
		suite.Require().NoError(err)
		suite.Require().Empty(res.VmError)

		nonceAfter := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

		// Check if balance & nonce were updated
		expectedBalance := balanceBefore.Sub(balanceBefore, big.NewInt(amountToTransfer))
		balanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address)

		isSenderBalanceCorrect := expectedBalance.Cmp(balanceAfter)
		suite.Require().True(isSenderBalanceCorrect == 0, "Incorrect sender's balance")

		// Check receiver's balance
		receiverBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, common.Address{})
		expectedReceiverBalance := receiverBalanceBefore.Add(receiverBalanceBefore, big.NewInt(amountToTransfer))
		isReceiverBalanceCorrect := expectedReceiverBalance.Cmp(receiverBalanceAfter)
		suite.Require().True(isReceiverBalanceCorrect == 0, "Incorrect receiver's balance")

		// Check if nonce was updated
		suite.Require().Equal(nonceBefore+1, nonceAfter)
	}
}

func (suite *KeeperTestSuite) TestMultipleContractDeployments() {
	suite.SetupSGXVMTest()

	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := types.ERC20Contract.ABI.Pack("", suite.address, big.NewInt(10))
	suite.Require().NoError(err)

	for i := 0; i < 5; i++ {
		nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

		data := append(types.ERC20Contract.Bin, ctorArgs...)
		args, err := json.Marshal(&types.TransactionArgs{
			From: &suite.address,
			Data: (*hexutil.Bytes)(&data),
		})
		suite.Require().NoError(err)
		gasRes, err := suite.queryClient.EstimateGas(ctx, &types.EthCallRequest{
			Args:            args,
			GasCap:          uint64(config.DefaultGasCap),
			ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
		})
		suite.Require().NoError(err)

		erc20DeployTx := types.NewSGXVMTxContract(
			chainID,
			nonce,
			nil,        // amount
			gasRes.Gas, // gasLimit
			nil,        // gasPrice
			nil, nil,
			data, // input
			nil,  // accesses
		)

		erc20DeployTx.From = suite.address.Hex()
		err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
		suite.Require().NoError(err)

		rsp, err := suite.app.EvmKeeper.HandleTx(ctx, erc20DeployTx)
		suite.Require().NoError(err)
		suite.Require().Empty(rsp.VmError)

		contractAddress := crypto.CreateAddress(suite.address, nonce)
		contractAcc := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, contractAddress)
		suite.Require().Equal(uint64(1), contractAcc.Nonce)
		suite.Require().Equal(new(big.Int), contractAcc.Balance)
		suite.Require().True(contractAcc.IsContract())
	}
}

func (suite *KeeperTestSuite) TestEmptyDataWithPublicKey() {
	transferAmount := big.NewInt(1000)
	suite.SetupSGXVMTest()

	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	err := suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, transferAmount)
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	// 4 byte function selector + 32 byte public key
	data := make([]byte, 36)
	args, err := json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)
	gasRes, err := suite.queryClient.EstimateGas(ctx, &types.EthCallRequest{
		Args:            args,
		GasCap:          uint64(config.DefaultGasCap),
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	suite.Require().NoError(err)

	tx := types.NewSGXVMTx(
		chainID,
		nonce,
		&suite.address,
		transferAmount, // amount
		gasRes.Gas,     // gasLimit
		nil,            // gasPrice
		nil, nil,
		data,     // input
		nil,      // accesses,
		nil, nil, // node private and public key
	)

	tx.From = suite.address.Hex()
	err = tx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)

	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	suite.Require().True(len(rsp.Ret) != 0)
}
