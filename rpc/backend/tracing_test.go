package backend

import (
	"fmt"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/cometbft/cometbft/libs/log"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/indexer"
	"swisstronik/rpc/backend/mocks"
	"swisstronik/utils"
	evmtypes "swisstronik/x/evm/types"
)

func (suite *BackendTestSuite) TestTraceTransaction() {
	msgHandleTx, _ := suite.buildEthereumTx()
	msgHandleTx2, _ := suite.buildEthereumTx()

	txHash := msgHandleTx.AsTransaction().Hash()
	txHash2 := msgHandleTx2.AsTransaction().Hash()

	priv, _ := ethsecp256k1.GenerateKey()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())

	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	RegisterParamsWithoutHeader(queryClient, 1)

	armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
	_ = suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
	ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())

	txEncoder := suite.backend.clientCtx.TxConfig.TxEncoder()

	msgHandleTx.From = from.String()
	_ = msgHandleTx.Sign(ethSigner, suite.signer)
	tx, _ := msgHandleTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), utils.BaseDenom)
	txBz, _ := txEncoder(tx)

	msgHandleTx2.From = from.String()
	_ = msgHandleTx2.Sign(ethSigner, suite.signer)
	tx2, _ := msgHandleTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), utils.BaseDenom)
	txBz2, _ := txEncoder(tx2)

	testCases := []struct {
		name          string
		registerMock  func()
		block         *types.Block
		responseBlock []*abci.ResponseDeliverTx
		expResult     interface{}
		expPass       bool
	}{
		{
			"fail - tx not found",
			func() {},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - block not found",
			func() {
				// var header metadata.MD
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, 1)
			},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			false,
		},
		{
			"pass - transaction found in a block with multiple transactions",
			func() {
				var (
					queryClient       = suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
					client            = suite.backend.clientCtx.Client.(*mocks.Client)
					height      int64 = 1
				)
				_, err := RegisterBlockMultipleTxs(client, height, []types.Tx{txBz, txBz2})
				suite.Require().NoError(err)
				RegisterTraceTransactionWithPredecessors(queryClient, msgHandleTx, []*evmtypes.MsgHandleTx{msgHandleTx})
			},
			&types.Block{Header: types.Header{Height: 1, ChainID: ChainID}, Data: types.Data{Txs: []types.Tx{txBz, txBz2}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash2.Hex()},
							{Key: "txIndex", Value: "1"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
		{
			"pass - transaction found",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, _ = RegisterBlock(client, 1, txBz)
				RegisterTraceTransaction(queryClient, msgHandleTx)
			},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			db := dbm.NewMemDB()
			suite.backend.indexer = indexer.NewKVIndexer(db, tmlog.NewNopLogger(), suite.backend.clientCtx)

			err := suite.backend.indexer.IndexBlock(tc.block, tc.responseBlock)
			suite.Require().NoError(err)
			txResult, err := suite.backend.TraceTransaction(txHash, nil)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expResult, txResult)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestTraceBlock() {
	msgEthTx, bz := suite.buildEthereumTx()
	emptyBlock := tmtypes.MakeBlock(1, []tmtypes.Tx{}, nil, nil)
	emptyBlock.ChainID = ChainID
	filledBlock := tmtypes.MakeBlock(1, []tmtypes.Tx{bz}, nil, nil)
	filledBlock.ChainID = ChainID
	resBlockEmpty := tmrpctypes.ResultBlock{Block: emptyBlock, BlockID: emptyBlock.LastBlockID}
	resBlockFilled := tmrpctypes.ResultBlock{Block: filledBlock, BlockID: filledBlock.LastBlockID}

	testCases := []struct {
		name            string
		registerMock    func()
		expTraceResults []*evmtypes.TxTraceResult
		resBlock        *tmrpctypes.ResultBlock
		config          *evmtypes.TraceConfig
		expPass         bool
	}{
		{
			"pass - no transaction returning empty array",
			func() {},
			[]*evmtypes.TxTraceResult{},
			&resBlockEmpty,
			&evmtypes.TraceConfig{},
			true,
		},
		{
			"fail - cannot unmarshal data",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				//client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterTraceBlock(queryClient, []*evmtypes.MsgHandleTx{msgEthTx})
				//RegisterConsensusParams(client, 1)
			},
			[]*evmtypes.TxTraceResult{},
			&resBlockFilled,
			&evmtypes.TraceConfig{},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			traceResults, err := suite.backend.TraceBlock(1, tc.config, tc.resBlock)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expTraceResults, traceResults)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
