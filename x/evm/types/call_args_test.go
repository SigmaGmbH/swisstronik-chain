package types

import (
	"fmt"
	"math/big"

	"swisstronik/crypto/ethsecp256k1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (suite *TxDataTestSuite) TestCallArgsString() {
	testCases := []struct {
		name           string
		callArgs       CallArgs
		expectedString string
	}{
		{
			"empty tx args",
			CallArgs{},
			"CallArgs{To:<nil>, Gas:<nil>, Nonce:<nil>, Data:<nil>, Input:<nil>, AccessList:<nil>}",
		},
		{
			"call args with fields",
			CallArgs{
				From:       &suite.addr,
				To:         &suite.addr,
				Gas:        &suite.hexUint64,
				Nonce:      &suite.hexUint64,
				Input:      &suite.hexInputBytes,
				Data:       &suite.hexDataBytes,
				AccessList: &ethtypes.AccessList{},
			},
			fmt.Sprintf("CallArgs{To:%v, Gas:%v, Nonce:%v, Data:%v, Input:%v, AccessList:%v}",
				&suite.addr,
				&suite.hexUint64,
				&suite.hexUint64,
				&suite.hexDataBytes,
				&suite.hexInputBytes,
				&ethtypes.AccessList{}),
		},
	}
	for _, tc := range testCases {
		outputString := tc.callArgs.String()
		suite.Require().Equal(outputString, tc.expectedString)
	}
}

func (suite *TxDataTestSuite) TestConvertCallArgsEthTx() {
	testCases := []struct {
		name     string
		callArgs CallArgs
	}{
		{
			"empty tx args",
			CallArgs{},
		},
		{
			"no nil args",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
		},
		{
			"max fee per gas nil, but access list not nil",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
		},
	}
	for _, tc := range testCases {
		res := tc.callArgs.ToTransaction()
		suite.Require().NotNil(res)
	}
}

func (suite *TxDataTestSuite) TestCallArgsToEVMMessage() {
	testCases := []struct {
		name         string
		callArgs     CallArgs
		globalGasCap uint64
		baseFee      *big.Int
		expError     bool
	}{
		{
			"empty tx args",
			CallArgs{},
			uint64(0),
			nil,
			false,
		},
		{
			"specify gasPrice and (maxFeePerGas or maxPriorityFeePerGas)",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(0),
			nil,
			true,
		},
		{
			"non-1559 execution, zero gas cap",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(0),
			nil,
			false,
		},
		{
			"non-1559 execution, nonzero gas cap",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			nil,
			false,
		},
		{
			"1559-type execution, nil gas price",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             nil,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			suite.bigInt,
			false,
		},
		{
			"1559-type execution, non-nil gas price",
			CallArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			suite.bigInt,
			false,
		},
	}
	for _, tc := range testCases {
		res, err := tc.callArgs.ToMessage(tc.globalGasCap, tc.baseFee)

		if tc.expError {
			suite.Require().NotNil(err)
		} else {
			suite.Require().Nil(err)
			suite.Require().NotNil(res)
		}
	}
}

func (suite *TxDataTestSuite) TestCallArgsGetFrom() {
	testCases := []struct {
		name       string
		callArgs   CallArgs
		expAddress common.Address
	}{
		{
			"empty signature field",
			CallArgs{},
			common.Address{},
		},
		{
			"non-empty from field, but empty signature",
			CallArgs{
				From: &suite.addr,
			},
			common.Address{},
		},
	}
	for _, tc := range testCases {
		retrievedAddress := tc.callArgs.GetFrom()
		suite.Require().Equal(retrievedAddress, tc.expAddress)
	}
}

func (suite *TxDataTestSuite) TestCallArgsGetFromSignature() {
	chainId := big.NewInt(suite.hexBigInt.ToInt().Int64())
	signer := ethtypes.LatestSignerForChainID(chainId)

	privateKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	ecdsaKey, err := privateKey.ToECDSA()
	suite.Require().NoError(err)
	originalSender := common.HexToAddress(privateKey.PubKey().Address().String())

	args := CallArgs{
		ChainID: &suite.hexBigInt,
		To:      &suite.addr,
		Value:   &suite.hexBigInt,
	}
	argsAsTx := args.ToTransaction().AsTransaction()
	signedTx, err := ethtypes.SignTx(argsAsTx, signer, ecdsaKey)
	suite.Require().NoError(err)

	v, r, s := signedTx.RawSignatureValues()
	vBig := hexutil.Big(*v)
	rBig := hexutil.Big(*r)
	sBig := hexutil.Big(*s)

	signedArgs := CallArgs{
		ChainID: &suite.hexBigInt,
		To:      &suite.addr,
		Value:   &suite.hexBigInt,
		V:       &vBig,
		R:       &rBig,
		S:       &sBig,
	}

	restoredSender := signedArgs.GetFrom()
	suite.Require().Equal(originalSender, restoredSender)
}

func (suite *TxDataTestSuite) TestCallArgsGetData() {
	testCases := []struct {
		name           string
		callArgs       CallArgs
		expectedOutput []byte
	}{
		{
			"empty input and data fields",
			CallArgs{
				Data:  nil,
				Input: nil,
			},
			nil,
		},
		{
			"empty input field, non-empty data field",
			CallArgs{
				Data:  &suite.hexDataBytes,
				Input: nil,
			},
			[]byte("data"),
		},
		{
			"non-empty input and data fields",
			CallArgs{
				Data:  &suite.hexDataBytes,
				Input: &suite.hexInputBytes,
			},
			[]byte("input"),
		},
	}
	for _, tc := range testCases {
		retrievedData := tc.callArgs.GetData()
		suite.Require().Equal(retrievedData, tc.expectedOutput)
	}
}
