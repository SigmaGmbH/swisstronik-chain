package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	feemarkettypes "swisstronik/x/feemarket/types"

	"cosmossdk.io/simapp"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"swisstronik/app"
	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/encoding"
	"swisstronik/server/config"
	"swisstronik/tests"
	evmcommontypes "swisstronik/types"
	evmtypes "swisstronik/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"swisstronik/utils"

	"github.com/SigmaGmbH/librustgo"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	queryClient evmtypes.QueryClient
	address     common.Address
	consAddress sdk.ConsAddress

	// for generate test tx
	clientCtx client.Context
	ethSigner ethtypes.Signer

	appCodec codec.Codec
	signer   keyring.Signer

	enableFeemarket  bool
	enableLondonHF   bool
	mintFeeCollector bool
	denom            string

	privateKey    []byte
	nodePublicKey []byte
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	if os.Getenv("benchmark") != "" {
		t.Skip("Skipping Gingko Test")
	}
	s = new(KeeperTestSuite)
	s.enableFeemarket = false
	s.enableLondonHF = true
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	chainID := utils.TestnetChainID + "-1"
	suite.app, _ = app.SetupSwissApp(checkTx, nil, chainID)
	suite.SetupApp(checkTx)
}

func (suite *KeeperTestSuite) SetupTestWithT(t require.TestingT) {
	checkTx := false
	chainID := utils.TestnetChainID + "-1"
	suite.app, _ = app.SetupSwissApp(checkTx, nil, chainID)
	suite.SetupAppWithT(checkTx, t, chainID)
}

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	chainID := utils.TestnetChainID + "-1"
	// Initialize enclave
	err := librustgo.InitializeEnclave(false)
	require.NoError(suite.T(), err)

	suite.SetupAppWithT(checkTx, suite.T(), chainID)
}

// SetupApp setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *KeeperTestSuite) SetupAppWithT(checkTx bool, t require.TestingT, chainID string) {
	// obtain node public key
	res, err := librustgo.GetNodePublicKey(0)
	suite.Require().NoError(err)
	suite.nodePublicKey = res.PublicKey

	// account key, use a constant account to keep unit test deterministic.
	ecdsaPriv, err := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	require.NoError(t, err)
	priv := &ethsecp256k1.PrivKey{
		Key: crypto.FromECDSA(ecdsaPriv),
	}

	suite.privateKey = priv.Bytes()
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewTestSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	suite.app = app.Setup(checkTx, func(app *app.App, genesis simapp.GenesisState) simapp.GenesisState {
		feemarketGenesis := feemarkettypes.DefaultGenesisState()
		if suite.enableFeemarket {
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
		} else {
			feemarketGenesis.Params.NoBaseFee = true
		}
		genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		if !suite.enableLondonHF {
			evmGenesis := evmtypes.DefaultGenesisState()
			maxInt := sdkmath.NewInt(math.MaxInt64)
			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
			evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
			evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
			evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
			genesis[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
		}

		return genesis
	})

	if suite.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(int64(params.TxGas)-1)))
		genesisState := app.NewTestGenesisState(suite.app.AppCodec())
		balances := []banktypes.Balance{
			{
				Address: suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName).String(),
				Coins:   coins,
			},
		}
		var bankGenesis banktypes.GenesisState
		suite.app.AppCodec().MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)
		// Update balances and total supply
		bankGenesis.Balances = append(bankGenesis.Balances, balances...)
		bankGenesis.Supply = bankGenesis.Supply.Add(coins...)
		genesisState[banktypes.ModuleName] = suite.app.AppCodec().MustMarshalJSON(&bankGenesis)

		// we marshal the genesisState of all module to a byte array
		stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		suite.app.InitChain(
			abci.RequestInitChain{
				ChainId:         chainID,
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: app.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	header := tmproto.Header{
		ChainID:         chainID,
		Height:          1,
		Time:            time.Now().UTC(),
		ValidatorsHash:  tmhash.Sum([]byte("validators")),
		AppHash:         tmhash.Sum([]byte("app")),
		ProposerAddress: suite.consAddress.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		DataHash:           tmhash.Sum([]byte("data")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
	}

	suite.ctx = suite.app.NewContext(checkTx, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = evmtypes.NewQueryClient(queryHelper)

	acc := &evmcommontypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = utils.BaseDenom
	err = suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
	require.NoError(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
}

func (suite *KeeperTestSuite) EvmDenom() string {
	ctx := sdk.WrapSDKContext(suite.ctx)
	rsp, _ := suite.queryClient.Params(ctx, &evmtypes.QueryParamsRequest{})
	return rsp.Params.EvmDenom
}

// Commit and begin new block
func (suite *KeeperTestSuite) Commit() {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()
	header.Height += 1
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = evmtypes.NewQueryClient(queryHelper)
}

// DeployVCContract deploy a test contract for verifying credentials and returns the contract address
func (suite *KeeperTestSuite) DeployVCContract(t require.TestingT) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	hexTxData := "0x608060405234801561001057600080fd5b50610705806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063738c4cd21461003b5780637ecc184d1461006c575b600080fd5b61005560048036038101906100509190610218565b61009d565b604051610063929190610336565b60405180910390f35b61008660048036038101906100819190610496565b610173565b604051610094929190610336565b60405180910390f35b6000606060008061040373ffffffffffffffffffffffffffffffffffffffff1686866040516100cd92919061050f565b600060405180830381855afa9150503d8060008114610108576040519150601f19603f3d011682016040523d82523d6000602084013e61010d565b606091505b509150915081610152576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161014990610574565b60405180910390fd5b60008061015e83610173565b91509150818195509550505050509250929050565b600060606000808480602001905181019061018e9190610673565b915091508181935093505050915091565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b60008083601f8401126101d8576101d76101b3565b5b8235905067ffffffffffffffff8111156101f5576101f46101b8565b5b602083019150836001820283011115610211576102106101bd565b5b9250929050565b6000806020838503121561022f5761022e6101a9565b5b600083013567ffffffffffffffff81111561024d5761024c6101ae565b5b610259858286016101c2565b92509250509250929050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061029082610265565b9050919050565b6102a081610285565b82525050565b600081519050919050565b600082825260208201905092915050565b60005b838110156102e05780820151818401526020810190506102c5565b60008484015250505050565b6000601f19601f8301169050919050565b6000610308826102a6565b61031281856102b1565b93506103228185602086016102c2565b61032b816102ec565b840191505092915050565b600060408201905061034b6000830185610297565b818103602083015261035d81846102fd565b90509392505050565b600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6103a3826102ec565b810181811067ffffffffffffffff821117156103c2576103c161036b565b5b80604052505050565b60006103d561019f565b90506103e1828261039a565b919050565b600067ffffffffffffffff8211156104015761040061036b565b5b61040a826102ec565b9050602081019050919050565b82818337600083830152505050565b6000610439610434846103e6565b6103cb565b90508281526020810184848401111561045557610454610366565b5b610460848285610417565b509392505050565b600082601f83011261047d5761047c6101b3565b5b813561048d848260208601610426565b91505092915050565b6000602082840312156104ac576104ab6101a9565b5b600082013567ffffffffffffffff8111156104ca576104c96101ae565b5b6104d684828501610468565b91505092915050565b600081905092915050565b60006104f683856104df565b9350610503838584610417565b82840190509392505050565b600061051c8284866104ea565b91508190509392505050565b7f43616e6e6f74207665726966792063726564656e7469616c0000000000000000600082015250565b600061055e6018836102b1565b915061056982610528565b602082019050919050565b6000602082019050818103600083015261058d81610551565b9050919050565b600061059f82610265565b9050919050565b6105af81610594565b81146105ba57600080fd5b50565b6000815190506105cc816105a6565b92915050565b600067ffffffffffffffff8211156105ed576105ec61036b565b5b6105f6826102ec565b9050602081019050919050565b6000610616610611846105d2565b6103cb565b90508281526020810184848401111561063257610631610366565b5b61063d8482856102c2565b509392505050565b600082601f83011261065a576106596101b3565b5b815161066a848260208601610603565b91505092915050565b6000806040838503121561068a576106896101a9565b5b6000610698858286016105bd565b925050602083015167ffffffffffffffff8111156106b9576106b86101ae565b5b6106c585828601610645565b915050925092905056fea264697066735822122080e1e1b342988f1853435195a5d23375962f495c6604ef963c796be2ab91503464736f6c63430008120033"
	data, err := hexutil.Decode(hexTxData)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var deployTx *evmtypes.MsgHandleTx
	if suite.enableFeemarket {
		deployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,       // amount
			1_000_000, // gasLimit
			nil,       // gasPrice
			suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			big.NewInt(1),
			data,                   // input
			&ethtypes.AccessList{}, // accesses
		)
	} else {
		deployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,       // amount
			1_000_000, // gasLimit
			nil,       // gasPrice
			nil, nil,
			data, // input
			nil,  // accesses
		)
	}

	deployTx.From = suite.address.Hex()
	err = deployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)

	ethTx := &evmtypes.MsgHandleTx{
		Data: deployTx.Data,
		Hash: deployTx.Hash,
		From: deployTx.From,
	}
	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, ethTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

// DeployTestContract deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestContract(t require.TestingT, owner common.Address, supply *big.Int) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := evmtypes.ERC20Contract.ABI.Pack("", owner, supply)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	data := append(evmtypes.ERC20Contract.Bin, ctorArgs...)
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)
	res, err := suite.queryClient.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          uint64(config.DefaultGasCap),
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	var erc20DeployTx *evmtypes.MsgHandleTx
	if suite.enableFeemarket {
		erc20DeployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,     // amount
			res.Gas, // gasLimit
			nil,     // gasPrice
			suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			big.NewInt(1),
			data,                   // input
			&ethtypes.AccessList{}, // accesses
		)
	} else {
		erc20DeployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,     // amount
			res.Gas, // gasLimit
			nil,     // gasPrice
			nil, nil,
			data, // input
			nil,  // accesses
		)
	}

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)

	ethTx := &evmtypes.MsgHandleTx{
		Data: erc20DeployTx.Data,
		Hash: erc20DeployTx.Hash,
		From: erc20DeployTx.From,
	}
	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, ethTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

// DeployTestMessageCall deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestMessageCall(t require.TestingT) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	data := evmtypes.TestMessageCall.Bin
	args, err := json.Marshal(&evmtypes.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)

	res, err := suite.queryClient.EstimateGas(ctx, &evmtypes.EthCallRequest{
		Args:            args,
		GasCap:          uint64(config.DefaultGasCap),
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var erc20DeployTx *evmtypes.MsgHandleTx
	if suite.enableFeemarket {
		erc20DeployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,     // amount
			res.Gas, // gasLimit
			nil,     // gasPrice
			suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			big.NewInt(1),
			data,                   // input
			&ethtypes.AccessList{}, // accesses
		)
	} else {
		erc20DeployTx = evmtypes.NewTxContract(
			chainID,
			nonce,
			nil,     // amount
			res.Gas, // gasLimit
			nil,     // gasPrice
			nil, nil,
			data, // input
			nil,  // accesses
		)
	}

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)

	ethTx := &evmtypes.MsgHandleTx{
		Data: erc20DeployTx.Data,
		Hash: erc20DeployTx.Hash,
		From: erc20DeployTx.From,
	}
	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, ethTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) TestBaseFee() {
	testCases := []struct {
		name            string
		enableLondonHF  bool
		enableFeemarket bool
		expectBaseFee   *big.Int
	}{
		{"not enable london HF, not enable feemarket", false, false, nil},
		{"enable london HF, not enable feemarket", true, false, big.NewInt(0)},
		{"enable london HF, enable feemarket", true, true, big.NewInt(1000000000)},
		{"not enable london HF, enable feemarket", false, true, nil},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest()
			suite.app.EvmKeeper.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
			evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			ethCfg := evmParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
			suite.Require().Equal(tc.expectBaseFee, baseFee)
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *KeeperTestSuite) TestGetAccountStorage() {
	testCases := []struct {
		name     string
		malleate func()
		expRes   []int
	}{
		{
			"Only one account that's not a contract (no storage)",
			func() {},
			[]int{0},
		},
		{
			"Two accounts - one contract (with storage), one wallet",
			func() {
				supply := big.NewInt(100)
				suite.DeployTestContract(suite.T(), suite.address, supply)
			},
			[]int{2, 0},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()
			i := 0
			suite.app.AccountKeeper.IterateAccounts(suite.ctx, func(account authtypes.AccountI) bool {
				ethAccount, ok := account.(evmcommontypes.EthAccountI)
				if !ok {
					// ignore non EthAccounts
					return false
				}

				addr := ethAccount.EthAddress()
				storage := suite.app.EvmKeeper.GetAccountStorage(suite.ctx, addr)

				suite.Require().Equal(tc.expRes[i], len(storage))
				i++
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetAccountOrEmpty() {
	empty := evmtypes.Account{
		Balance:  new(big.Int),
		CodeHash: evmtypes.EmptyCodeHash,
	}

	supply := big.NewInt(100)
	contractAddr := suite.DeployTestContract(suite.T(), suite.address, supply)

	testCases := []struct {
		name     string
		addr     common.Address
		expEmpty bool
	}{
		{
			"unexisting account - get empty",
			common.Address{},
			true,
		},
		{
			"existing contract account",
			contractAddr,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			res := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, tc.addr)
			if tc.expEmpty {
				suite.Require().Equal(empty, res)
			} else {
				suite.Require().NotEqual(empty, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetNonce() {
	testCases := []struct {
		name          string
		address       common.Address
		expectedNonce uint64
		malleate      func()
	}{
		{
			"account not found",
			tests.RandomEthAddress(),
			0,
			func() {},
		},
		{
			"existing account",
			suite.address,
			1,
			func() {
				suite.Require().NoError(
					suite.app.EvmKeeper.SetNonce(suite.ctx, suite.address, 1),
				)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, tc.address)
			suite.Require().Equal(tc.expectedNonce, nonce)
		})
	}
}

func (suite *KeeperTestSuite) TestSetNonce() {
	testCases := []struct {
		name     string
		address  common.Address
		nonce    uint64
		malleate func()
	}{
		{
			"new account",
			tests.RandomEthAddress(),
			10,
			func() {},
		},
		{
			"existing account",
			suite.address,
			99,
			func() {},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Require().NoError(
				suite.app.EvmKeeper.SetNonce(suite.ctx, tc.address, tc.nonce),
			)
			nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, tc.address)
			suite.Require().Equal(tc.nonce, nonce)
		})
	}
}

func (suite *KeeperTestSuite) TestGetCodeHash() {
	addr := tests.RandomEthAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name     string
		address  common.Address
		expHash  common.Hash
		malleate func()
	}{
		{
			"account not found",
			tests.RandomEthAddress(),
			common.BytesToHash(evmtypes.EmptyCodeHash),
			func() {},
		},
		{
			"account not EthAccount type, EmptyCodeHash",
			addr,
			common.BytesToHash(evmtypes.EmptyCodeHash),
			func() {},
		},
		{
			"existing account",
			suite.address,
			crypto.Keccak256Hash([]byte("codeHash")),
			func() {
				err := suite.app.EvmKeeper.SetAccountCode(suite.ctx, suite.address, []byte("codeHash"))
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			hash := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, tc.address).CodeHash
			suite.Require().Equal(tc.expHash, common.BytesToHash(hash))
		})
	}
}

func (suite *KeeperTestSuite) TestSetCode() {
	addr := tests.RandomEthAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name    string
		address common.Address
		code    []byte
		isNoOp  bool
	}{
		{
			"account not found",
			tests.RandomEthAddress(),
			[]byte("code"),
			false,
		},
		{
			"account not EthAccount type",
			addr,
			nil,
			true,
		},
		{
			"existing account",
			suite.address,
			[]byte("code"),
			false,
		},
		{
			"existing account, code deleted from store",
			suite.address,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			prev, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, tc.address)
			suite.Require().NoError(err)

			err = suite.app.EvmKeeper.SetAccountCode(suite.ctx, tc.address, tc.code)
			suite.Require().NoError(err)

			post, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, tc.address)
			suite.Require().NoError(err)

			if tc.isNoOp {
				suite.Require().Equal(prev, post)
			} else {
				suite.Require().Equal(tc.code, post)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestKeeperSetAccountCode() {
	addr := tests.RandomEthAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name string
		code []byte
	}{
		{
			"set code",
			[]byte("codecodecode"),
		},
		{
			"delete code",
			nil,
		},
	}

	suite.SetupTest()

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.app.EvmKeeper.SetAccountCode(suite.ctx, addr, tc.code)
			suite.Require().NoError(err)

			acct := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, addr)
			suite.Require().NotNil(acct)

			if tc.code != nil {
				suite.Require().True(acct.IsContract())
			}

			codeHash := crypto.Keccak256Hash(tc.code)
			suite.Require().Equal(codeHash.Bytes(), acct.CodeHash)

			code := suite.app.EvmKeeper.GetCode(suite.ctx, common.BytesToHash(acct.CodeHash))
			suite.Require().Equal(tc.code, code)
		})
	}
}

func (suite *KeeperTestSuite) TestKeeperSetCode() {
	addr := tests.RandomEthAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name     string
		codeHash []byte
		code     []byte
	}{
		{
			"set code",
			[]byte("codeHash"),
			[]byte("this is the code"),
		},
		{
			"delete code",
			[]byte("codeHash"),
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.SetCode(suite.ctx, tc.codeHash, tc.code)
			key := suite.app.GetKey(evmtypes.StoreKey)
			store := prefix.NewStore(suite.ctx.KVStore(key), evmtypes.KeyPrefixCode)
			code := store.Get(tc.codeHash)

			suite.Require().Equal(tc.code, code)
		})
	}
}

func (suite *KeeperTestSuite) TestState() {
	testCases := []struct {
		name       string
		key, value common.Hash
	}{
		{
			"set state - delete from store",
			common.BytesToHash([]byte("key")),
			common.Hash{},
		},
		{
			"set state - update value",
			common.BytesToHash([]byte("key")),
			common.BytesToHash([]byte("value")),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.SetState(suite.ctx, suite.address, tc.key, tc.value.Bytes())
			value := suite.app.EvmKeeper.GetState(suite.ctx, suite.address, tc.key)
			suite.Require().Equal(tc.value, common.BytesToHash(value))
		})
	}
}

func (suite *KeeperTestSuite) TestSuicide() {
	code := []byte("code")
	err := suite.app.EvmKeeper.SetAccountCode(suite.ctx, suite.address, code)
	suite.Require().NoError(err)

	addedCode, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, suite.address)
	suite.Require().NoError(err)

	suite.Require().Equal(code, addedCode)
	// Add state to account
	for i := 0; i < 5; i++ {
		suite.app.EvmKeeper.SetState(suite.ctx, suite.address, common.BytesToHash([]byte(fmt.Sprintf("key%d", i))), []byte(fmt.Sprintf("value%d", i)))
	}

	// Generate 2nd address
	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	suite.Require().NoError(err)
	addr2 := crypto.PubkeyToAddress(key.PublicKey)

	// Add code and state to account 2
	err = suite.app.EvmKeeper.SetAccountCode(suite.ctx, addr2, code)
	suite.Require().NoError(err)

	addedCode2, _ := suite.app.EvmKeeper.GetAccountCode(suite.ctx, addr2)
	suite.Require().Equal(code, addedCode2)
	for i := 0; i < 5; i++ {
		suite.app.EvmKeeper.SetState(suite.ctx, addr2, common.BytesToHash([]byte(fmt.Sprintf("key%d", i))), []byte(fmt.Sprintf("value%d", i)))
	}

	// Destroy first contract
	err = suite.app.EvmKeeper.DeleteAccount(suite.ctx, suite.address)
	suite.Require().NoError(err)

	// Check code is deleted
	accCode, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, suite.address)
	suite.Require().NoError(err)
	suite.Require().Nil(accCode)

	// Check state is deleted
	var storage evmtypes.Storage
	suite.app.EvmKeeper.ForEachStorage(suite.ctx, suite.address, func(key common.Hash, value []byte) bool {
		storage = append(storage, evmtypes.NewState(key, value))
		return true
	})
	suite.Require().Equal(0, len(storage))

	// Check account is deleted
	acc := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, suite.address)
	suite.Require().Equal(evmtypes.EmptyCodeHash, acc.CodeHash)

	// Check code is still present in addr2 and suicided is false
	code2, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().NotNil(code2)
}

func (suite *KeeperTestSuite) TestExist() {
	testCases := []struct {
		name     string
		address  common.Address
		malleate func()
		exists   bool
	}{
		{"success, account exists", suite.address, func() {}, true},
		{"success, account doesn't exist", tests.RandomEthAddress(), func() {}, false},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()
			exists := suite.app.EvmKeeper.GetAccount(suite.ctx, tc.address) != nil
			suite.Require().Equal(tc.exists, exists)
		})
	}
}

func (suite *KeeperTestSuite) TestEmpty() {
	testCases := []struct {
		name     string
		address  common.Address
		malleate func()
		empty    bool
	}{
		{
			"not empty, positive balance",
			suite.address,
			func() {
				err := suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(100))
				suite.Require().NoError(err)
			},
			false,
		},
		{"empty, account doesn't exist", tests.RandomEthAddress(), func() {}, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()
			isEmpty := suite.app.EvmKeeper.GetAccount(suite.ctx, tc.address) == nil
			suite.Require().Equal(tc.empty, isEmpty)
		})
	}
}

func (suite *KeeperTestSuite) CreateTestTx(msg *evmtypes.MsgHandleTx, priv cryptotypes.PrivKey) authsigning.Tx {
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	suite.Require().NoError(err)

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)

	err = msg.Sign(suite.ethSigner, tests.NewTestSigner(priv))
	suite.Require().NoError(err)

	err = txBuilder.SetMsgs(msg)
	suite.Require().NoError(err)

	return txBuilder.GetTx()
}

func (suite *KeeperTestSuite) TestForEachStorage() {
	var storage evmtypes.Storage

	testCase := []struct {
		name      string
		malleate  func()
		callback  func(key common.Hash, value []byte) (stop bool)
		expValues []common.Hash
	}{
		{
			"aggregate state",
			func() {
				for i := 0; i < 5; i++ {
					suite.app.EvmKeeper.SetState(suite.ctx, suite.address, common.BytesToHash([]byte(fmt.Sprintf("key%d", i))), []byte(fmt.Sprintf("value%d", i)))
				}
			},
			func(key common.Hash, value []byte) bool {
				storage = append(storage, evmtypes.NewState(key, value))
				return true
			},
			[]common.Hash{
				common.BytesToHash([]byte("value0")),
				common.BytesToHash([]byte("value1")),
				common.BytesToHash([]byte("value2")),
				common.BytesToHash([]byte("value3")),
				common.BytesToHash([]byte("value4")),
			},
		},
		{
			"filter state",
			func() {
				suite.app.EvmKeeper.SetState(suite.ctx, suite.address, common.BytesToHash([]byte("key")), []byte("value"))
				suite.app.EvmKeeper.SetState(suite.ctx, suite.address, common.BytesToHash([]byte("filterkey")), []byte("filtervalue"))
			},
			func(key common.Hash, value []byte) bool {
				if common.Bytes2Hex(value) == common.Bytes2Hex([]byte("filtervalue")) {
					storage = append(storage, evmtypes.NewState(key, value))
					return false
				}
				return true
			},
			[]common.Hash{
				common.BytesToHash([]byte("filtervalue")),
			},
		},
	}

	for _, tc := range testCase {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.app.EvmKeeper.ForEachStorage(suite.ctx, suite.address, tc.callback)
			suite.Require().Equal(len(tc.expValues), len(storage), fmt.Sprintf("Expected values:\n%v\nStorage Values\n%v", tc.expValues, storage))

			vals := make([]common.Hash, len(storage))
			for i := range storage {
				vals[i] = common.HexToHash(storage[i].Value)
			}

			suite.Require().ElementsMatch(tc.expValues, vals)
		})
		storage = evmtypes.Storage{}
	}
}

func (suite *KeeperTestSuite) TestSetBalance() {
	amount := big.NewInt(-10)

	testCases := []struct {
		name     string
		addr     common.Address
		malleate func()
		expErr   bool
	}{
		{
			"address without funds - invalid amount",
			suite.address,
			func() {},
			true,
		},
		{
			"mint to address",
			suite.address,
			func() {
				amount = big.NewInt(100)
			},
			false,
		},
		{
			"burn from address",
			suite.address,
			func() {
				amount = big.NewInt(60)
			},
			false,
		},
		{
			"address with funds - invalid amount",
			suite.address,
			func() {
				amount = big.NewInt(-10)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()
			err := suite.app.EvmKeeper.SetBalance(suite.ctx, tc.addr, amount)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				balance := suite.app.EvmKeeper.GetBalance(suite.ctx, tc.addr)
				suite.Require().NoError(err)
				suite.Require().Equal(amount, balance)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeleteAccount() {
	supply := big.NewInt(100)
	contractAddr := suite.DeployTestContract(suite.T(), suite.address, supply)

	testCases := []struct {
		name   string
		addr   common.Address
		expErr bool
	}{
		{
			"remove address",
			suite.address,
			false,
		},
		{
			"remove unexistent address - returns nil error",
			common.HexToAddress("unexistent_address"),
			false,
		},
		{
			"remove deployed contract",
			contractAddr,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			err := suite.app.EvmKeeper.DeleteAccount(suite.ctx, tc.addr)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				balance := suite.app.EvmKeeper.GetBalance(suite.ctx, tc.addr)
				suite.Require().Equal(new(big.Int), balance)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestNodePublicKey() {
	blockNumber := uint64(10)
	expNodePublicKey := common.BytesToHash([]byte("nodepublickey"))

	suite.app.EvmKeeper.SetNodePublicKey(suite.ctx, blockNumber, expNodePublicKey)
	nodePublicKey, err := suite.app.EvmKeeper.GetNodePublicKey(suite.ctx, blockNumber)
	suite.Require().NoError(err)
	suite.Require().Equal(expNodePublicKey, nodePublicKey)

	nodePublicKey, err = suite.app.EvmKeeper.GetNodePublicKey(suite.ctx, blockNumber+1)
	suite.Require().ErrorIs(err, evmtypes.ErrEmptyNodePublicKey)
	suite.Require().Equal(common.Hash{}, nodePublicKey)
}
