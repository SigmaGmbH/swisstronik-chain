package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	simappparams "cosmossdk.io/simapp/params"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/streaming"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	evmante "swisstronik/app/ante"
	"swisstronik/app/upgrades/v1_0_3"
	"swisstronik/app/upgrades/v1_0_4"
	"swisstronik/app/upgrades/v1_0_5"
	"swisstronik/app/upgrades/v1_0_6"
	"swisstronik/app/upgrades/v1_0_7"
	"swisstronik/docs"
	"swisstronik/encoding"
	srvflags "swisstronik/server/flags"
	evmcommontypes "swisstronik/types"
	compliancemodule "swisstronik/x/compliance"
	compliancemodulekeeper "swisstronik/x/compliance/keeper"
	compliancemoduletypes "swisstronik/x/compliance/types"
	"swisstronik/x/evm"
	evmkeeper "swisstronik/x/evm/keeper"
	evmtypes "swisstronik/x/evm/types"
	"swisstronik/x/feemarket"
	feemarketkeeper "swisstronik/x/feemarket/keeper"
	feemarkettypes "swisstronik/x/feemarket/types"
	vestingmodule "swisstronik/x/vesting"
	vestingmodulekeeper "swisstronik/x/vesting/keeper"
	vestingmodulesimulation "swisstronik/x/vesting/simulation"
	vestingmoduletypes "swisstronik/x/vesting/types"
)

const (
	AccountAddressPrefix = "swtr"
	Name                 = "swisstronik"
)

// this line is used by starport scaffolding # stargate/wasm/app/enabledProposals

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		// compliancemoduleclient.VerifyIssuerProposalHandler, this functionality is disabled for now
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		ica.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vestingmodule.AppModuleBasic{},
		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{},
		mint.AppModuleBasic{},
		// groupmodule.AppModuleBasic{},
		consensus.AppModuleBasic{},
		compliancemodule.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		icatypes.ModuleName:            nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		evmtypes.ModuleName:            {authtypes.Minter, authtypes.Burner}, // used for secure addition and subtraction of balance using module account
		// this line is used by starport scaffolding # stargate/app/maccPerms
	}
)

var (
	_ servertypes.Application = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
	RegisterCoinDenominations()
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper       authkeeper.AccountKeeper
	AuthzKeeper         authzkeeper.Keeper
	BankKeeper          bankkeeper.Keeper
	CapabilityKeeper    *capabilitykeeper.Keeper
	StakingKeeper       stakingkeeper.Keeper
	SlashingKeeper      slashingkeeper.Keeper
	MintKeeper          mintkeeper.Keeper
	DistrKeeper         distrkeeper.Keeper
	GovKeeper           govkeeper.Keeper
	CrisisKeeper        crisiskeeper.Keeper
	UpgradeKeeper       upgradekeeper.Keeper
	ParamsKeeper        paramskeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper      evidencekeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	FeeGrantKeeper      feegrantkeeper.Keeper
	// GroupKeeper           groupkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper

	EvmKeeper       *evmkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper

	VestingKeeper    vestingmodulekeeper.Keeper
	ComplianceKeeper compliancemodulekeeper.Keeper
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

	// ModuleManager is the module manager
	ModuleManager *module.Manager

	// sm is the simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig simappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Codec
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	// Setup Mempool and Proposal Handlers
	baseAppOptions = append(baseAppOptions, func(app *baseapp.BaseApp) {
		mempool := mempool.NoOpMempool{}
		app.SetMempool(mempool)
		handler := baseapp.NewDefaultProposalHandler(mempool, app)
		app.SetPrepareProposal(handler.PrepareProposalHandler())
		app.SetProcessProposal(handler.ProcessProposalHandler())
	})

	// NOTE we use custom transaction decoder that supports the sdk.Tx interface instead of sdk.StdTx
	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		ibcexported.StoreKey, ibctransfertypes.StoreKey, icahosttypes.StoreKey, capabilitytypes.StoreKey,
		consensusparamtypes.StoreKey, group.StoreKey, icacontrollertypes.StoreKey, crisistypes.StoreKey,
		evmtypes.StoreKey, feemarkettypes.StoreKey,
		vestingmoduletypes.StoreKey, compliancemoduletypes.StoreKey,
		// this line is used by starport scaffolding # stargate/app/storeKey
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey, feemarkettypes.TransientKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, compliancemoduletypes.MemStoreKey)

	// load state streaming if enabled
	if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, logger, keys); err != nil {
		fmt.Printf("failed to load state streaming: %s", err)
		os.Exit(1)
	}

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		cdc,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	// get authority address
	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authAddr)
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	// this line is used by starport scaffolding # stargate/app/scopedKeeper

	// Sealing prevents other modules from creating scoped sub-keepers
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		evmcommontypes.ProtoAccount,
		maccPerms,
		sdk.Bech32PrefixAccAddr,
		authAddr,
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedModuleAccountAddrs(),
		authAddr,
	)

	app.StakingKeeper = *stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authAddr,
	)

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		&app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		&app.StakingKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		app.LegacyAmino(),
		keys[slashingtypes.StoreKey],
		&app.StakingKeeper,
		authAddr,
	)

	app.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)

	app.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authAddr,
	)

	feeMarketSs := app.GetSubspace(feemarkettypes.ModuleName)
	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec, authtypes.NewModuleAddress(govtypes.ModuleName),
		keys[feemarkettypes.StoreKey], tkeys[feemarkettypes.TransientKey], feeMarketSs,
	)

	app.ComplianceKeeper = *compliancemodulekeeper.NewKeeper(
		keys[compliancemoduletypes.StoreKey],
		keys[compliancemoduletypes.MemStoreKey],
		app.GetSubspace(compliancemoduletypes.ModuleName),
	)
	complianceModule := compliancemodule.NewAppModule(appCodec, app.ComplianceKeeper)

	// Set authority to x/gov module account to only expect the module account to update params
	evmSs := app.GetSubspace(evmtypes.ModuleName)
	app.EvmKeeper = evmkeeper.NewKeeper(
		appCodec, keys[evmtypes.StoreKey], tkeys[evmtypes.TransientKey], authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.FeeMarketKeeper, app.ComplianceKeeper, evmSs,
	)

	// ... other modules keepers

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, app.MsgServiceRouter(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)
	var icaControllerIBCModule ibcporttypes.IBCModule
	icaControllerIBCModule = icacontroller.NewIBCMiddleware(icaControllerIBCModule, app.ICAControllerKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		&app.StakingKeeper,
		app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(&app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(compliancemoduletypes.RouterKey, compliancemodule.NewComplianceProposalHandler(&app.ComplianceKeeper))
	govConfig := govtypes.Config{
		MaxMetadataLen: 5000,
	}
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		&app.StakingKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authAddr,
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// insert governance hooks receivers here
		),
	)

	app.VestingKeeper = *vestingmodulekeeper.NewKeeper(
		keys[vestingmoduletypes.StoreKey],
		keys[vestingmoduletypes.MemStoreKey],
		app.GetSubspace(vestingmoduletypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
	)
	vestingModule := vestingmodule.NewAppModule(appCodec, app.VestingKeeper, app.AccountKeeper, app.BankKeeper)

	// this line is used by starport scaffolding # stargate/app/keeperDefinition

	/**** IBC Routing ****/

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerIBCModule)
	// this line is used by starport scaffolding # ibc/app/router
	app.IBCKeeper.SetRouter(ibcRouter)

	/**** Module Hooks ****/

	// register hooks after all modules have been initialized

	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			// insert staking hooks receivers here
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	/**** Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, vestingmodulesimulation.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		staking.NewAppModule(appCodec, &app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(&app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),

		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, minttypes.DefaultInflationCalculationFn, app.GetSubspace(minttypes.ModuleName)),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		transferModule,
		icaModule,
		feemarket.NewAppModule(app.FeeMarketKeeper, feeMarketSs),
		evm.NewAppModule(app.EvmKeeper, app.AccountKeeper, evmSs),
		vestingModule,
		complianceModule,
		// this line is used by starport scaffolding # stargate/app/appModule
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingmoduletypes.ModuleName,
		compliancemoduletypes.ModuleName,
		consensusparamtypes.ModuleName,

		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingmoduletypes.ModuleName,
		compliancemoduletypes.ModuleName,
		consensusparamtypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.ModuleManager.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		// NOTE: feemarket need to be initialized before genutil module:
		// gentx transactions use MinGasPriceDecorator.AnteHandle
		genutiltypes.ModuleName,
		icatypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		crisistypes.ModuleName,
		vestingmoduletypes.ModuleName,
		compliancemoduletypes.ModuleName,
		consensusparamtypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

	// Uncomment if you want to set a custom migration order here.
	// app.ModuleManager.SetOrderMigrations(custom order)

	app.ModuleManager.RegisterInvariants(&app.CrisisKeeper)
	// app.ModuleManager.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager.RegisterServices(app.configurator)

	// add test gRPC service for testing gRPC queries in isolation
	// testdata.RegisterTestServiceServer(app.GRPCQueryRouter(), testdata.TestServiceImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	// create the simulation manager and define the order of the modules for deterministic simulations
	overrideModules := map[string]module.AppModuleSimulation{
		//authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))
	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	maxGasWanted := cast.ToUint64(appOpts.Get(srvflags.EVMMaxTxGasWanted))
	app.setAnteHandler(encodingConfig.TxConfig, maxGasWanted)
	app.setPostHandler()
	app.SetEndBlocker(app.EndBlocker)
	app.setupUpgradeHandlers()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper
	// this line is used by starport scaffolding # stargate/app/beforeInitReturn

	return app
}

func (app *App) setAnteHandler(txConfig client.TxConfig, maxGasWanted uint64) {
	anteHandler, err := evmante.NewAnteHandler(evmante.HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		FeegrantKeeper:         app.FeeGrantKeeper,
		SigGasConsumer:         evmante.DefaultSigVerificationGasConsumer,
		IBCKeeper:              app.IBCKeeper,
		EvmKeeper:              app.EvmKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		MaxTxGasWanted:         maxGasWanted,
		ExtensionOptionChecker: evmcommontypes.HasDynamicFeeExtensionOption,
		TxFeeChecker:           evmante.NewDynamicFeeChecker(app.EvmKeeper),
	})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}

func (app *App) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// GetBaseApp implements the TestingApp interface.
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper implements the TestingApp interface.
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetStakingKeeperSDK implements the TestingApp interface.
func (app *App) GetStakingKeeperSDK() stakingkeeper.Keeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *App) GetTxConfig() client.TxConfig {
	cfg := encoding.MakeConfig(ModuleBasics)
	return cfg.TxConfig
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.ModuleManager.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.ModuleManager.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *App) BlockedModuleAccountAddrs() map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	docs.RegisterOpenAPIService(Name, apiSvr.Router)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService implements the Application.RegisterNodeService method.
func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())
	paramsKeeper.Subspace(vestingmoduletypes.ModuleName)
	paramsKeeper.Subspace(compliancemoduletypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	return paramsKeeper
}

func (app *App) setupUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		v1_0_3.UpgradeName,
		v1_0_3.CreateUpgradeHandler(
			app.ModuleManager, app.configurator,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v1_0_4.UpgradeName,
		v1_0_4.CreateUpgradeHandler(
			app.ModuleManager, app.configurator,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v1_0_5.UpgradeName,
		v1_0_5.CreateUpgradeHandler(
			app.ModuleManager,
			app.BankKeeper,
			app.configurator,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v1_0_6.UpgradeName,
		v1_0_6.CreateUpgradeHandler(
			app.ModuleManager, app.configurator,
		),
	)

	app.UpgradeKeeper.SetUpgradeHandler(
		v1_0_7.UpgradeName,
		v1_0_7.CreateUpgradeHandler(
			app.ModuleManager,
			app.ComplianceKeeper,
			app.EvmKeeper,
			app.configurator,
		),
	)

	// When a planned update height is reached, the old binary will panic
	// writing on disk the height and name of the update that triggered it
	// This will read that value, and execute the preparations for the upgrade.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	if upgradeInfo.Name == v1_0_3.UpgradeName {
		// Use upgrade store loader for the initial loading of all stores when app starts,
		// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
		// so that new stores start with the correct version (the current height of chain),
		// instead the default which is the latest version that store last committed i.e 0 for new stores.
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storetypes.StoreUpgrades{}))
	}

	if upgradeInfo.Name == v1_0_4.UpgradeName {
		// Use upgrade store loader for the initial loading of all stores when app starts,
		// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
		// so that new stores start with the correct version (the current height of chain),
		// instead the default which is the latest version that store last committed i.e 0 for new stores.
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storetypes.StoreUpgrades{}))
	}

	if upgradeInfo.Name == v1_0_5.UpgradeName {
		// Use upgrade store loader for the initial loading of all stores when app starts,
		// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
		// so that new stores start with the correct version (the current height of chain),
		// instead the default which is the latest version that store last committed i.e 0 for new stores.
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storetypes.StoreUpgrades{}))
	}
}

// SimulationManager implements runtime.AppI
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

func RegisterCoinDenominations() {
	_ = sdk.RegisterDenom("swtr", sdk.OneDec())
	_ = sdk.RegisterDenom("aswtr", sdk.NewDecWithPrec(1, 18))
}
