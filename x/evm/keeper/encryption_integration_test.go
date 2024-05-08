package keeper_test

import (
	"encoding/json"
	"github.com/SigmaGmbH/librustgo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strconv"
	"strings"
	"swisstronik/crypto/deoxys"
	"swisstronik/server/config"
	"swisstronik/x/evm/types"
)

func (suite *KeeperTestSuite) TestGetNodePublicKey() {
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
	nodePublicKey, err := suite.app.EvmKeeper.GetNodePublicKey(0)
	suite.Require().NoError(err)
	nodePublicKeyV1, err := suite.app.EvmKeeper.GetNodePublicKey(v1EpochStartingBlock)
	suite.Require().NoError(err)
	nodePublicKeyV2, err := suite.app.EvmKeeper.GetNodePublicKey(v2EpochStartingBlock)
	suite.Require().NoError(err)

	suite.Require().NotEqual(nodePublicKey, nodePublicKeyV1)
	suite.Require().NotEqual(nodePublicKeyV1, nodePublicKeyV2)
}

func (suite *KeeperTestSuite) TestAddEpoch() {
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
	// test plan
	// 1. deploy contract, which writes some data in storage
	// 2. write some state
	// 3. update epoch
	// 4. try to decrypt previous state

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	// Add another epoch
	var v1EpochStartingBlock uint64 = 20
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// deploy contract
	ctx := sdk.WrapSDKContext(suite.ctx)
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := types.SimpleStorageContract.ABI.Pack("")
	suite.Require().NoError(err)

	data := append(types.SimpleStorageContract.Bin, ctorArgs...)
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

	storageDeployTx := types.NewSGXVMTxContract(
		chainID,
		nonce,
		nil,        // amount
		gasRes.Gas, // gasLimit
		nil,        // gasPrice
		nil, nil,
		data, // input
		nil,  // accesses
	)

	storageDeployTx.From = suite.address.Hex()
	err = storageDeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)

	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, storageDeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)

	contractAddress := crypto.CreateAddress(suite.address, nonce)
	contractAcc := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, contractAddress)
	suite.Require().Equal(uint64(1), contractAcc.Nonce)
	suite.Require().Equal(new(big.Int), contractAcc.Balance)
	suite.Require().True(contractAcc.IsContract())

	nodePublicKeyRes, err := librustgo.GetNodePublicKey(uint64(suite.ctx.BlockHeader().Height))
	suite.Require().NoError(err)

	// write some data to contract
	writtenValue := int64(100)
	setStorageArgs, err := types.SimpleStorageContract.ABI.Pack("store", big.NewInt(writtenValue))
	suite.Require().NoError(err)
	setStorageTx := types.NewSGXVMTx(
		chainID,
		nonce,
		&contractAddress,
		nil,
		uint64(100_000),
		nil,
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(0),
		setStorageArgs,
		&ethtypes.AccessList{}, // accesses
		suite.privateKey,
		nodePublicKeyRes.PublicKey,
	)

	setStorageTx.From = suite.address.Hex()
	err = setStorageTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err = suite.app.EvmKeeper.HandleTx(ctx, setStorageTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)

	// TODO: Change block number
	// try to read data from contract
	//updatedBlockHeight := 100
	//nodePublicKeyV1Res, err := librustgo.GetNodePublicKey(uint64(updatedBlockHeight))
	//suite.Require().NoError(err)
	//nodePublicKeyV1 := nodePublicKeyV1Res.PublicKey
	//suite.Require().NotEqual(nodePublicKeyRes.PublicKey, nodePublicKeyV1, "should be different encryption key")

	getStorageData, err := types.SimpleStorageContract.ABI.Pack("get")
	suite.Require().NoError(err)
	encryptedData, err := deoxys.EncryptECDH(suite.privateKey, nodePublicKeyRes.PublicKey, getStorageData)
	suite.Require().NoError(err)

	getStorageArgs, err := json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		To:   &contractAddress,
		Data: (*hexutil.Bytes)(&encryptedData),
	})

	res, err := suite.queryClient.EthCall(suite.ctx, &types.EthCallRequest{
		Args:   getStorageArgs,
		GasCap: uint64(config.DefaultGasCap),
	})
	suite.Require().NoError(err)
	suite.Require().Empty(res.VmError)

	// Decrypt response
	decryptedData, err := deoxys.DecryptECDH(suite.privateKey, nodePublicKeyRes.PublicKey, res.Ret)
	suite.Require().NoError(err)
	decodedValue, err := parseIntResponse(decryptedData)
	suite.Require().NoError(err)

	suite.Require().Equal(writtenValue, decodedValue)
}

func parseIntResponse(data []byte) (int64, error) {
	hexDecryptedData := hexutil.Encode(data)
	var dataToParse = hexDecryptedData
	if strings.HasPrefix(hexDecryptedData, "0x") {
		dataToParse = strings.TrimPrefix(dataToParse, "0x")
	}

	return strconv.ParseInt(dataToParse, 16, 64)
}
