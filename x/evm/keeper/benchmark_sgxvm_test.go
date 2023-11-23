package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/ethereum/go-ethereum/common"

	evmcommontypes "swisstronik/types"
	"swisstronik/x/evm/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func SetupContractSGXVM(b *testing.B) (*KeeperTestSuite, common.Address) {
	suite := KeeperTestSuite{}
	suite.SetupSGXVMTestWithT(b)

	amt := sdk.Coins{evmcommontypes.NewPhotonCoinInt64(1000000000000000000)}
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
	require.NoError(b, err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
	require.NoError(b, err)

	contractAddr := suite.DeploySGXVMTestContract(b, suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
	suite.Commit()

	return &suite, contractAddr
}

func SetupSGXVMTestMessageCall(b *testing.B) (*KeeperTestSuite, common.Address) {
	suite := KeeperTestSuite{}
	suite.SetupSGXVMTestWithT(b)

	amt := sdk.Coins{evmcommontypes.NewPhotonCoinInt64(1000000000000000000)}
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
	require.NoError(b, err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
	require.NoError(b, err)

	contractAddr := suite.DeploySGXVMTestMessageCall(b)
	suite.Commit()

	return &suite, contractAddr
}

type SGXVMTxBuilder func(suite *KeeperTestSuite, contract common.Address) *types.MsgHandleTx

func DoBenchmarkSGXVM(b *testing.B, txBuilder SGXVMTxBuilder) {
	suite, contractAddr := SetupContractSGXVM(b)

	msg := txBuilder(suite, contractAddr)
	msg.From = suite.address.Hex()
	err := msg.Sign(ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID()), suite.signer)
	require.NoError(b, err)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := suite.ctx.CacheContext()

		// deduct fee first
		txData, err := types.UnpackTxData(msg.Data)
		require.NoError(b, err)

		fees := sdk.Coins{sdk.NewCoin(suite.EvmDenom(), sdkmath.NewIntFromBigInt(txData.Fee()))}
		err = authante.DeductFees(suite.app.BankKeeper, suite.ctx, suite.app.AccountKeeper.GetAccount(ctx, msg.GetFrom()), fees)
		require.NoError(b, err)

		rsp, err := suite.app.EvmKeeper.HandleTx(sdk.WrapSDKContext(ctx), msg)
		require.NoError(b, err)
		require.False(b, rsp.Failed())
	}
}

func BenchmarkTokenTransferSGXVM(b *testing.B) {
	DoBenchmarkSGXVM(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgHandleTx {
		input, err := types.ERC20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
		return types.NewSGXVMTx(suite.app.EvmKeeper.ChainID(), nonce, &contract, big.NewInt(0), 410000, big.NewInt(1), nil, nil, input, nil, suite.privateKey, suite.nodePublicKey)
	})
}

func BenchmarkEmitLogsSGXVM(b *testing.B) {
	DoBenchmarkSGXVM(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgHandleTx {
		input, err := types.ERC20Contract.ABI.Pack("benchmarkLogs", big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
		return types.NewSGXVMTx(suite.app.EvmKeeper.ChainID(), nonce, &contract, big.NewInt(0), 4100000, big.NewInt(1), nil, nil, input, nil, suite.privateKey, suite.nodePublicKey)
	})
}

func BenchmarkTokenTransferFromSGXVM(b *testing.B) {
	DoBenchmarkSGXVM(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgHandleTx {
		input, err := types.ERC20Contract.ABI.Pack("transferFrom", suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(0))
		require.NoError(b, err)
		nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
		return types.NewSGXVMTx(suite.app.EvmKeeper.ChainID(), nonce, &contract, big.NewInt(0), 410000, big.NewInt(1), nil, nil, input, nil, suite.privateKey, suite.nodePublicKey)
	})
}

func BenchmarkTokenMintSGXVM(b *testing.B) {
	DoBenchmarkSGXVM(b, func(suite *KeeperTestSuite, contract common.Address) *types.MsgHandleTx {
		input, err := types.ERC20Contract.ABI.Pack("mint", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
		require.NoError(b, err)
		nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
		return types.NewSGXVMTx(suite.app.EvmKeeper.ChainID(), nonce, &contract, big.NewInt(0), 410000, big.NewInt(1), nil, nil, input, nil, suite.privateKey, suite.nodePublicKey)
	})
}

func BenchmarkMessageCallSGXVM(b *testing.B) {
	suite, contract := SetupSGXVMTestMessageCall(b)

	input, err := types.TestMessageCall.ABI.Pack("benchmarkMessageCall", big.NewInt(10000))
	require.NoError(b, err)
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
	msg := types.NewSGXVMTx(suite.app.EvmKeeper.ChainID(), nonce, &contract, big.NewInt(0), 25000000, big.NewInt(1), nil, nil, input, nil, suite.privateKey, suite.nodePublicKey)

	msg.From = suite.address.Hex()
	err = msg.Sign(ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID()), suite.signer)
	require.NoError(b, err)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := suite.ctx.CacheContext()

		// deduct fee first
		txData, err := types.UnpackTxData(msg.Data)
		require.NoError(b, err)

		fees := sdk.Coins{sdk.NewCoin(suite.EvmDenom(), sdkmath.NewIntFromBigInt(txData.Fee()))}
		err = authante.DeductFees(suite.app.BankKeeper, suite.ctx, suite.app.AccountKeeper.GetAccount(ctx, msg.GetFrom()), fees)
		require.NoError(b, err)

		rsp, err := suite.app.EvmKeeper.HandleTx(sdk.WrapSDKContext(ctx), msg)
		require.NoError(b, err)
		require.False(b, rsp.Failed())
	}
}
