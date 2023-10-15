package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"math"
	"math/big"
	"os"
	didtypes "swisstronik/x/did/types"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	feemarkettypes "swisstronik/x/feemarket/types"

	"swisstronik/app"
	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/encoding"
	"swisstronik/server/config"
	"swisstronik/tests"
	evmcommontypes "swisstronik/types"
	"swisstronik/x/evm/types"
	evmtypes "swisstronik/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"swisstronik/go-sgxvm"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	queryClient types.QueryClient
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
	suite.app = app.Setup(checkTx, nil)
	suite.SetupApp(checkTx)
}

func (suite *KeeperTestSuite) SetupTestWithT(t require.TestingT) {
	checkTx := false
	suite.app = app.Setup(checkTx, nil)
	suite.SetupAppWithT(checkTx, t)
}

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	// Initialize enclave
	err := librustgo.InitializeMasterKey(false)
	require.NoError(suite.T(), err)

	suite.SetupAppWithT(checkTx, suite.T())
}

// SetupApp setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *KeeperTestSuite) SetupAppWithT(checkTx bool, t require.TestingT) {
	// obtain node public key
	res, err := librustgo.GetNodePublicKey()
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
	suite.signer = tests.NewSigner(priv)

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
			evmGenesis := types.DefaultGenesisState()
			maxInt := sdkmath.NewInt(math.MaxInt64)
			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
			evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
			evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
			evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
			genesis[types.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
		}
		return genesis
	})

	if suite.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(types.DefaultEVMDenom, sdkmath.NewInt(int64(params.TxGas)-1)))
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
				ChainId:         "ethermint_9000-1",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: app.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "ethermint_9000-1",
		Time:            time.Now().UTC(),
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
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

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

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
}

func (suite *KeeperTestSuite) EvmDenom() string {
	ctx := sdk.WrapSDKContext(suite.ctx)
	rsp, _ := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
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
	types.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// DeployVCContract deploy a test contract for verifying credentials and returns the contract address
func (suite *KeeperTestSuite) DeployVCContract(t require.TestingT) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	hexTxData := "0x608060405234801561001057600080fd5b506103db806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80631d73814b14610030575b600080fd5b61004a60048036038101906100459190610257565b61004c565b005b600061040373ffffffffffffffffffffffffffffffffffffffff16826040516100759190610311565b600060405180830381855afa9150503d80600081146100b0576040519150601f19603f3d011682016040523d82523d6000602084013e6100b5565b606091505b50509050806100f9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016100f090610385565b60405180910390fd5b5050565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6101648261011b565b810181811067ffffffffffffffff821117156101835761018261012c565b5b80604052505050565b60006101966100fd565b90506101a2828261015b565b919050565b600067ffffffffffffffff8211156101c2576101c161012c565b5b6101cb8261011b565b9050602081019050919050565b82818337600083830152505050565b60006101fa6101f5846101a7565b61018c565b90508281526020810184848401111561021657610215610116565b5b6102218482856101d8565b509392505050565b600082601f83011261023e5761023d610111565b5b813561024e8482602086016101e7565b91505092915050565b60006020828403121561026d5761026c610107565b5b600082013567ffffffffffffffff81111561028b5761028a61010c565b5b61029784828501610229565b91505092915050565b600081519050919050565b600081905092915050565b60005b838110156102d45780820151818401526020810190506102b9565b60008484015250505050565b60006102eb826102a0565b6102f581856102ab565b93506103058185602086016102b6565b80840191505092915050565b600061031d82846102e0565b915081905092915050565b600082825260208201905092915050565b7f43616e6e6f74207665726966792063726564656e7469616c0000000000000000600082015250565b600061036f601883610328565b915061037a82610339565b602082019050919050565b6000602082019050818103600083015261039e81610362565b905091905056fea26469706673582212208c08e6999cc8f2aad23d0c5c507ac56bfafb5ef0c37ef749f47b378663dcf44564736f6c63430008130033"
	data, err := hexutil.Decode(hexTxData)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var deployTx *types.MsgHandleTx
	if suite.enableFeemarket {
		deployTx = types.NewTxContract(
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
		deployTx = types.NewTxContract(
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

	ethTx := &types.MsgHandleTx{
		Data: deployTx.Data,
		Hash: deployTx.Hash,
		From: deployTx.From,
	}
	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, ethTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

// Sends sample transaction to verify VC
func (suite *KeeperTestSuite) SendVerifiableCredentialTx(t require.TestingT, contractAddr common.Address) {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	txDataHex := "0x1d73814b000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001e6b901e365794a68624763694f694a465a45525451534973496e523563434936496b705856434a392e65794a325979493665794a4159323975644756346443493657794a6f64485277637a6f764c336433647935334d793576636d63764d6a41784f43396a636d566b5a57353061574673637939324d534a644c434a306558426c496a7062496c5a6c636d6c6d6157466962475644636d566b5a57353061574673496c3073496d4e795a57526c626e52705957785464574a715a574e30496a7037496d466b5a484a6c63334d694f694a7a643352794d544e7a6247786a5a484e786147706c61335268597a56794e6d67314d475232616e4a306147307765585132656e637a6354527a496e31394c434a7a645749694f694a6b6157513663336430636a6f334f56566c5956685654566c3161484e764f48524c5a6d31344f55787149697769626d4a6d496a6f784e6a6b324e6a41354d7a63344c434a7063334d694f694a6b6157513663336430636a6f334f56566c5956685654566c3161484e764f48524c5a6d31344f557871496e302e442d62523449755f4b764836576b774c59666644656c4d37446247396d34345353616d6672473872374779714e4c426a305f4c61754531617877766d655f4143446a633362766b566f4141325259425f4a566f4e44410000000000000000000000000000000000000000000000000000"
	txData, err := hexutil.Decode(txDataHex)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var tx *types.MsgHandleTx
	if suite.enableFeemarket {
		tx = evmtypes.NewSGXVMTx(
			chainID,
			nonce,
			&contractAddr,
			nil,
			500_000,
			nil,
			suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			big.NewInt(1),
			txData,
			&ethtypes.AccessList{},
			suite.privateKey,
			suite.nodePublicKey,
		)
	} else {
		tx = evmtypes.NewSGXVMTx(
			chainID,
			nonce,
			&contractAddr,
			nil,
			500_000,
			nil,
			nil,
			nil,
			txData,
			nil,
			suite.privateKey,
			suite.nodePublicKey,
		)
	}

	tx.From = suite.address.Hex()
	err = tx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	require.NoError(t, err)

	ethTx := &types.MsgHandleTx{
		Data: tx.Data,
		Hash: tx.Hash,
		From: tx.From,
	}
	rsp, err := suite.app.EvmKeeper.HandleTx(ctx, ethTx)
	require.NoError(t, err)
	require.Empty(t, rsp.VmError)
}

// DeployTestContract deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) DeployTestContract(t require.TestingT, owner common.Address, supply *big.Int) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := types.ERC20Contract.ABI.Pack("", owner, supply)
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	data := append(types.ERC20Contract.Bin, ctorArgs...)
	args, err := json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)
	res, err := suite.queryClient.EstimateGas(ctx, &types.EthCallRequest{
		Args:            args,
		GasCap:          uint64(config.DefaultGasCap),
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	var erc20DeployTx *types.MsgHandleTx
	if suite.enableFeemarket {
		erc20DeployTx = types.NewTxContract(
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
		erc20DeployTx = types.NewTxContract(
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

	ethTx := &types.MsgHandleTx{
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

	data := types.TestMessageCall.Bin
	args, err := json.Marshal(&types.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	require.NoError(t, err)

	res, err := suite.queryClient.EstimateGas(ctx, &types.EthCallRequest{
		Args:            args,
		GasCap:          uint64(config.DefaultGasCap),
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	require.NoError(t, err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var erc20DeployTx *types.MsgHandleTx
	if suite.enableFeemarket {
		erc20DeployTx = types.NewTxContract(
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
		erc20DeployTx = types.NewTxContract(
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

	ethTx := &types.MsgHandleTx{
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
	empty := types.Account{
		Balance:  new(big.Int),
		CodeHash: types.EmptyCodeHash,
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
			tests.GenerateAddress(),
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
			tests.GenerateAddress(),
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
	addr := tests.GenerateAddress()
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
			tests.GenerateAddress(),
			common.BytesToHash(types.EmptyCodeHash),
			func() {},
		},
		{
			"account not EthAccount type, EmptyCodeHash",
			addr,
			common.BytesToHash(types.EmptyCodeHash),
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
	addr := tests.GenerateAddress()
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
			tests.GenerateAddress(),
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
	addr := tests.GenerateAddress()
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
	addr := tests.GenerateAddress()
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
			key := suite.app.GetKey(types.StoreKey)
			store := prefix.NewStore(suite.ctx.KVStore(key), types.KeyPrefixCode)
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

	addedCode2, err := suite.app.EvmKeeper.GetAccountCode(suite.ctx, addr2)
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
	var storage types.Storage
	suite.app.EvmKeeper.ForEachStorage(suite.ctx, suite.address, func(key, value common.Hash) bool {
		storage = append(storage, types.NewState(key, value))
		return true
	})
	suite.Require().Equal(0, len(storage))

	// Check account is deleted
	acc := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, suite.address)
	suite.Require().Equal(types.EmptyCodeHash, acc.CodeHash)

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
		{"success, account doesn't exist", tests.GenerateAddress(), func() {}, false},
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
		{"empty, account doesn't exist", tests.GenerateAddress(), func() {}, true},
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

func (suite *KeeperTestSuite) CreateTestTx(msg *types.MsgHandleTx, priv cryptotypes.PrivKey) authsigning.Tx {
	option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionsEthereumTx{})
	suite.Require().NoError(err)

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)

	err = msg.Sign(suite.ethSigner, tests.NewSigner(priv))
	suite.Require().NoError(err)

	err = txBuilder.SetMsgs(msg)
	suite.Require().NoError(err)

	return txBuilder.GetTx()
}

func (suite *KeeperTestSuite) TestForEachStorage() {
	var storage types.Storage

	testCase := []struct {
		name      string
		malleate  func()
		callback  func(key, value common.Hash) (stop bool)
		expValues []common.Hash
	}{
		{
			"aggregate state",
			func() {
				for i := 0; i < 5; i++ {
					suite.app.EvmKeeper.SetState(suite.ctx, suite.address, common.BytesToHash([]byte(fmt.Sprintf("key%d", i))), []byte(fmt.Sprintf("value%d", i)))
				}
			},
			func(key, value common.Hash) bool {
				storage = append(storage, types.NewState(key, value))
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
			func(key, value common.Hash) bool {
				if value == common.BytesToHash([]byte("filtervalue")) {
					storage = append(storage, types.NewState(key, value))
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
		storage = types.Storage{}
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

func (suite *KeeperTestSuite) TestVerifyCredential() {
	// Add verifiable credential
	var err error

	// Create DID Document for issuer
	metadata := didtypes.Metadata{
		Created:   time.Now(),
		VersionId: "123e4567-e89b-12d3-a456-426655440000",
	}
	didUrl := "did:swtr:79UeaXUMYuhso8tKfmx9Lj"
	verificationMethods := []*didtypes.VerificationMethod{{
		Id:                     "did:swtr:79UeaXUMYuhso8tKfmx9Lj#7fe2db8637700730ce468995ee89cd986d9d2e3d07266171abdaf6d11a9c7732-1",
		VerificationMethodType: "Ed25519VerificationKey2020",
		Controller:             didUrl,
		VerificationMaterial:   "z6Mko4UTTiH1d8ZcThnBxokPaVggEQ9QkuJuEWgBTwwJAYcq",
	}}
	document := didtypes.DIDDocument{
		Id:                 didUrl,
		Controller:         []string{didUrl},
		VerificationMethod: verificationMethods,
		Authentication:     []string{"did:swtr:79UeaXUMYuhso8tKfmx9Lj#7fe2db8637700730ce468995ee89cd986d9d2e3d07266171abdaf6d11a9c7732-1"},
	}
	didDocument := didtypes.DIDDocumentWithMetadata{
		Metadata: &metadata,
		DidDoc:   &document,
	}
	err = suite.app.EvmKeeper.DIDKeeper.AddNewDIDDocumentVersion(suite.ctx, &didDocument)
	suite.Require().NoError(err)

	// Deploy contract VC.sol
	vcContract := suite.DeployVCContract(suite.T())

	// Send transaction to verify credentials
	suite.SendVerifiableCredentialTx(suite.T(), vcContract)
}
