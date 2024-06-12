package keeper_test

// Since this integration test breaks some other tests,
// we skip those tests

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/SigmaGmbH/librustgo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"swisstronik/crypto/deoxys"
	"swisstronik/server/config"
	"swisstronik/x/evm/types"
)

func (suite *KeeperTestSuite) TestGetNodePublicKey() {
	suite.T().Skip()

	var v1EpochStartingBlock uint64 = 2
	var v2EpochStartingBlock uint64 = 5

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	// Add 2 epochs
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	err = librustgo.AddEpoch(v2EpochStartingBlock)
	suite.Require().NoError(err)

	updatedEpochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(updatedEpochs), 3, "Should be 3 epochs")

	// Request node public key
	nodePublicKeyResponse, err := librustgo.GetNodePublicKey(uint64(0))
	suite.Require().NoError(err)
	nodePublicKey := nodePublicKeyResponse.PublicKey
	nodePublicKeyResponse, err = librustgo.GetNodePublicKey(v1EpochStartingBlock)
	suite.Require().NoError(err)
	nodePublicKeyV1 := nodePublicKeyResponse.PublicKey
	nodePublicKeyResponse, err = librustgo.GetNodePublicKey(v2EpochStartingBlock)
	suite.Require().NoError(err)
	nodePublicKeyV2 := nodePublicKeyResponse.PublicKey

	suite.Require().NotEqual(nodePublicKey, nodePublicKeyV1)
	suite.Require().NotEqual(nodePublicKeyV1, nodePublicKeyV2)
}

func (suite *KeeperTestSuite) TestKeeperGetNodePublicKey() {
	suite.T().Skip()

	var v1EpochStartingBlock uint64 = 2
	var v2EpochStartingBlock uint64 = 5

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	// Add 2 epochs
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	err = librustgo.AddEpoch(v2EpochStartingBlock)
	suite.Require().NoError(err)

	updatedEpochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(updatedEpochs), 3, "Should be 3 epochs")

	nodePublicKey, err := suite.app.EvmKeeper.GetNodePublicKey(suite.ctx, 0)
	suite.Require().NoError(err)

	for i := uint64(0); i < v1EpochStartingBlock; i++ {
		suite.Commit()
	}
	nodePublicKeyV1, err := suite.app.EvmKeeper.GetNodePublicKey(suite.ctx, v1EpochStartingBlock)
	suite.Require().NoError(err)
	for i := v1EpochStartingBlock; i < v2EpochStartingBlock; i++ {
		suite.Commit()
	}
	nodePublicKeyV2, err := suite.app.EvmKeeper.GetNodePublicKey(suite.ctx, v2EpochStartingBlock)
	suite.Require().NoError(err)

	suite.Require().NotEqual(nodePublicKey, nodePublicKeyV1)
	suite.Require().NotEqual(nodePublicKeyV1, nodePublicKeyV2)

}

func (suite *KeeperTestSuite) TestAddEpoch() {
	suite.T().Skip()
	var v1EpochStartingBlock uint64 = 2

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	epochs, err := librustgo.ListEpochs()
	suite.Require().Equal(len(epochs), 1, "Should be only one epoch")
	suite.Require().Equal(epochs[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochs[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Add new epoch key, which starts from 2nd block
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// Check updated epochs
	updatedEpochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(updatedEpochs), 2, "Should be two epochs")
	suite.Require().Equal(updatedEpochs[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(updatedEpochs[1].StartingBlock, v1EpochStartingBlock, "Incorrect epoch starting block")
}

func (suite *KeeperTestSuite) TestRemoveEpoch() {
	suite.T().Skip()
	var v1EpochStartingBlock uint64 = 2
	var v2EpochStartingBlock uint64 = 3

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	epochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochs), 1, "Should be only one epoch")
	suite.Require().Equal(epochs[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochs[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Should not be able to remove last epoch
	err = librustgo.RemoveLatestEpoch()
	suite.Require().Error(err)

	// Add another epoch
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// Check updated epochs
	epochsAfterAdd, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterAdd), 2, "Should be two epochs")
	suite.Require().Equal(epochsAfterAdd[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(epochsAfterAdd[1].StartingBlock, v1EpochStartingBlock, "Incorrect epoch starting block")

	// Should be able to remove latest epoch
	err = librustgo.RemoveLatestEpoch()
	suite.Require().NoError(err)

	epochsAfterRemoval, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterRemoval), 1, "Should be only one epoch")
	suite.Require().Equal(epochsAfterRemoval[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochsAfterRemoval[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Should be able to add epoch again with changed starting block
	err = librustgo.AddEpoch(v2EpochStartingBlock)
	suite.Require().NoError(err)

	epochsAfterNewAdd, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterNewAdd), 2, "Should be two epochs")
	suite.Require().Equal(epochsAfterNewAdd[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(epochsAfterNewAdd[1].StartingBlock, v2EpochStartingBlock, "Incorrect epoch starting block")
}

func (suite *KeeperTestSuite) TestCrossEpochInteraction() {
	suite.T().Skip()
	// test plan
	// 1. deploy contract, which writes some data in storage
	// 2. write some state
	// 3. update epoch
	// 4. try to decrypt previous state

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	// Add another epoch, starting from 20th block
	var v1EpochStartingBlock uint64 = 20
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// deploy Incrementor contract
	ctx := sdk.WrapSDKContext(suite.ctx)
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
	chainID := suite.app.EvmKeeper.ChainID()

	constructorArgs, err := types.IncrementorContract.ABI.Pack("")
	suite.Require().NoError(err)

	data := append(types.IncrementorContract.Bin, constructorArgs...)

	deployTx := types.NewSGXVMTxContract(
		chainID,
		nonce,
		nil,     // amount
		200_000, // gasLimit
		nil,     // gasPrice
		nil, nil,
		data, // input
		nil,  // accesses
	)

	deployTx.From = suite.address.Hex()
	err = deployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)

	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, deployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)

	contractAddress := crypto.CreateAddress(suite.address, nonce)
	contractAcc := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, contractAddress)
	suite.Require().Equal(uint64(1), contractAcc.Nonce)
	suite.Require().Equal(new(big.Int), contractAcc.Balance)
	suite.Require().True(contractAcc.IsContract())

	// Incrementor contract was deployed. Now we're calling `increment` function
	// to update contract state.
	nodePublicKeyResponse, err := librustgo.GetNodePublicKey(uint64(suite.ctx.BlockHeight()))
	suite.Require().NoError(err)
	initialNodePublicKey := nodePublicKeyResponse.PublicKey

	// increment initial value at the contract
	incrementArgs, err := types.IncrementorContract.ABI.Pack("increment")
	suite.Require().NoError(err)
	incrementTx := types.NewSGXVMTx(
		chainID,
		nonce,
		&contractAddress,
		nil,
		uint64(100_000),
		nil,
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(0),
		incrementArgs,
		&ethtypes.AccessList{}, // accesses
		suite.privateKey,
		initialNodePublicKey,
	)

	incrementTx.From = suite.address.Hex()
	err = incrementTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err = suite.app.EvmKeeper.HandleTx(ctx, incrementTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)

	// read contract state at block 1.
	// Use the same context.
	getStateData, err := types.IncrementorContract.ABI.Pack("state")
	suite.Require().NoError(err)
	encryptedGetStateData, err := deoxys.EncryptECDH(suite.privateKey, initialNodePublicKey, getStateData)
	suite.Require().NoError(err)

	getStorageArgs, err := json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		To:   &contractAddress,
		Data: (*hexutil.Bytes)(&encryptedGetStateData),
	})

	res, err := suite.app.EvmKeeper.EthCall(ctx, &types.EthCallRequest{
		Args:   getStorageArgs,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)
	suite.Require().Empty(res.VmError)

	resultAtBlock1, err := decryptAndParseResponse(suite.privateKey, initialNodePublicKey, res.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1), resultAtBlock1)

	// read contract state at 21st block.
	// should be able to obtain value and return the same result.
	// should not accept previous node public key
	updatedCtx := suite.ctx.WithBlockHeight(21)
	res, err = suite.app.EvmKeeper.EthCall(updatedCtx, &types.EthCallRequest{
		Args:   getStorageArgs,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)
	suite.Require().NotEmpty(res.VmError)
	suite.Require().True(strings.Contains(res.VmError, "DecryptionError"))

	// try again with updated node public key
	nodePublicKeyResponse, err = librustgo.GetNodePublicKey(uint64(updatedCtx.BlockHeight()))
	suite.Require().NoError(err)
	updatedNodePublicKey := nodePublicKeyResponse.PublicKey
	suite.Require().NotEqual(initialNodePublicKey, updatedNodePublicKey) // should use different node public keys
	encryptedGetStateData, err = deoxys.EncryptECDH(suite.privateKey, updatedNodePublicKey, getStateData)
	suite.Require().NoError(err)
	getStorageArgs, err = json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		To:   &contractAddress,
		Data: (*hexutil.Bytes)(&encryptedGetStateData),
	})

	res, err = suite.app.EvmKeeper.EthCall(updatedCtx, &types.EthCallRequest{
		Args:   getStorageArgs,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)
	suite.Require().Empty(res.VmError)
	resultAtBlock21, err := decryptAndParseResponse(suite.privateKey, updatedNodePublicKey, res.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(1), resultAtBlock21)

	// increment value at 21st block
	incrementTx = types.NewSGXVMTx(
		chainID,
		nonce,
		&contractAddress,
		nil,
		uint64(100_000),
		nil,
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(0),
		incrementArgs,
		&ethtypes.AccessList{}, // accesses
		suite.privateKey,
		updatedNodePublicKey,
	)

	incrementTx.From = suite.address.Hex()
	err = incrementTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err = suite.app.EvmKeeper.HandleTx(updatedCtx, incrementTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)

	// check value again
	res, err = suite.app.EvmKeeper.EthCall(updatedCtx, &types.EthCallRequest{
		Args:   getStorageArgs,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)
	suite.Require().Empty(res.VmError)
	updatedResult, err := decryptAndParseResponse(suite.privateKey, updatedNodePublicKey, res.Ret)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(2), updatedResult)
}

func decryptAndParseResponse(userPrivateKey, nodePublicKey, encryptedResponse []byte) (int64, error) {
	data, err := deoxys.DecryptECDH(userPrivateKey, nodePublicKey, encryptedResponse)
	if err != nil {
		return 0, err
	}

	hexDecryptedData := hexutil.Encode(data)
	var dataToParse = hexDecryptedData
	if strings.HasPrefix(hexDecryptedData, "0x") {
		dataToParse = strings.TrimPrefix(dataToParse, "0x")
	}

	return strconv.ParseInt(dataToParse, 16, 64)
}
