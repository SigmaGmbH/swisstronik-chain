//go:build !nosgx
// +build !nosgx

package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"swisstronik/ethereum/eip712"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// this line is used by starport scaffolding # root/moduleImport

	"swisstronik/app"

	evmmoduleclient "swisstronik/client"
	evmmoduleserver "swisstronik/server"
	evmserverconfig "swisstronik/server/config"
	srvflags "swisstronik/server/flags"

	"swisstronik/utils"

	ethermint "swisstronik/types"

	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
)

const ShortBlockWindow uint32 = 20

// NewRootCmd creates a new root command for a Cosmos SDK application
func NewRootCmd() (*cobra.Command, ethermint.EncodingConfig) {
	// Initialize the SDK config the first before doing anything else.
	InitSDKConfig()

	encodingConfig := app.MakeConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.FlagBroadcastMode).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("").
		WithLedgerHasProtobuf(true)

	eip712.SetEncodingConfig(encodingConfig)

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "Swisstronik Daemon",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL.
			enabledSignModes := append(authtx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts := authtx.ConfigOptions{
				EnabledSignModes:           enabledSignModes,
				TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
			}
			txConfigWithTextual, err := authtx.NewTxConfigWithOptions(
				codec.NewProtoCodec(encodingConfig.InterfaceRegistry),
				txConfigOpts,
			)
			if err != nil {
				return err
			}
			initClientCtx = initClientCtx.WithTxConfig(txConfigWithTextual)

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customTMConfig := initTendermintConfig()
			return sdkserver.InterceptConfigsPreRunHandler(
				cmd, customAppTemplate, customAppConfig, customTMConfig,
			)
		},
	}

	txConfig := authtx.NewTxConfig(encodingConfig.Codec, authtx.DefaultSignModes)
	// gentxModule := app.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	valOperAddressCodec := encodingConfig.TxConfig.SigningContext().ValidatorAddressCodec()
	a := appCreator{encodingConfig}
	rootCmd.AddCommand(
		evmmoduleclient.ValidateChainID(
			genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, genutiltypes.DefaultMessageValidator, valOperAddressCodec),
		genutilcli.MigrateGenesisCmd(genutilcli.MigrationMap),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, valOperAddressCodec),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome, txConfig.SigningContext().ValidatorAddressCodec()),
		tmcli.NewCompletionCmd(rootCmd, true),
		NewTestnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		confixcmd.ConfigCommand(),
		EnclaveCmd(),
		DebugCmd(),
		snapshot.Cmd(a.newApp),
	)

	evmmoduleserver.AddCommands(
		rootCmd,
		evmmoduleserver.NewDefaultStartOptions(a.newApp, app.DefaultNodeHome),
		a.appExport,
		addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		sdkserver.StatusCommand(),
		queryCommand(),
		txCommand(),
		evmmoduleclient.KeyCommands(app.DefaultNodeHome),
	)
	rootCmd, err := srvflags.AddTxFlags(rootCmd)
	if err != nil {
		panic(err)
	}

	db := dbm.NewMemDB()
	chainID := utils.TestnetChainID + "-1"
	newapp := app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		simutils.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)

	initClientCtx, err = config.ReadDefaultValuesFromDefaultClientConfig(initClientCtx)
	if err != nil {
		panic(err)
	}
	if err := autoCliOpts(newapp, initClientCtx).EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd, encodingConfig
}

// autoCliOpts returns options based upon the modules in the Swisstronik app.
//
// Creates an instance of the application that is discarded to enumerate the modules.
func autoCliOpts(app *app.App, initClientCtx client.Context) autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	cliKR, err := keyring.NewAutoCLIKeyring(initClientCtx.Keyring)
	if err != nil {
		panic(err)
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdktypes.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdktypes.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdktypes.GetConfig().GetBech32ConsensusAddrPrefix()),
		Keyring:               cliKR,
		ClientCtx:             initClientCtx,
	}
}

// queryCommand returns the sub-command to send queries to the app
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		sdkserver.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// txCommand returns the sub-command to send transactions to the app
func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	// this line is used by starport scaffolding # root/arguments
}

type appCreator struct {
	encodingConfig ethermint.EncodingConfig
}

// newApp creates a new Cosmos SDK app
func (a appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(sdkserver.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(sdkserver.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := sdkserver.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	home := cast.ToString(appOpts.Get(flags.FlagHome))
	snapshotDir := filepath.Join(home, "data", "snapshots")
	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(sdkserver.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(sdkserver.FlagStateSyncSnapshotKeepRecent)),
	)

	// Setup chainId
	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if len(chainID) == 0 {
		v := viper.New()
		v.AddConfigPath(filepath.Join(home, "config"))
		v.SetConfigName("client")
		v.SetConfigType("toml")
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
		conf := new(config.ClientConfig)
		if err := v.Unmarshal(conf); err != nil {
			panic(err)
		}
		chainID = conf.ChainID
	}

	return app.New(
		logger,
		db,
		traceStore,
		true,
		skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(sdkserver.FlagInvCheckPeriod)),
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(sdkserver.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(sdkserver.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(sdkserver.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(sdkserver.FlagHaltTime))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(sdkserver.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(sdkserver.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(sdkserver.FlagIAVLCacheSize))),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(appOpts.Get(sdkserver.FlagDisableIAVLFastNode))),
		baseapp.SetChainID(chainID),
	)
}

// appExport creates a new simapp (optionally at a given height)
func (a appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var swissApp *app.App

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	if height != -1 {

		swissApp = app.New(
			logger,
			db,
			traceStore,
			height == -1, // -1: no height provided
			map[int64]bool{},
			homePath,
			uint(1),
			appOpts,
		)
	} else {
		swissApp = app.New(logger, db, traceStore, true, map[int64]bool{}, "", uint(1), appOpts)

	}

	return swissApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// TODO(DEC-1716): Set default seeds.
	cfg.P2P.Seeds = ""

	// Expose the Tendermint RPC.
	cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	cfg.RPC.CORSAllowedOrigins = []string{"*"}

	cfg.Consensus.TimeoutCommit = time.Second * 3
	// use v0 since v1 severely impacts the node's performance
	// Mempool config.
	// We specifically are using a number greater than max QPS (currently set at 5000) * ShortBlockWindow to prevent
	// a replay attack that is possible with short-term order placements and cancellations. The attack would consume
	// a users rate limit if the entry is evicted from the mempool cache as it would be possible for the transaction
	// to go through `CheckTx` again causing it to hit rate limit code against the users account.
	cfg.Mempool.CacheSize = 5000 * int(ShortBlockWindow)
	cfg.Mempool.Size = 50000
	cfg.Mempool.KeepInvalidTxsInCache = true

	// to put a higher strain on node memory, use these values:
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	// Enable pex.
	cfg.P2P.PexReactor = true

	// Enable telemetry.
	cfg.Instrumentation.Prometheus = true

	// Set default commit timeout to 500ms for faster block time.
	// Note: avoid using 1s since it's considered tne default Tendermint value
	// (https://github.com/dydxprotocol/tendermint/blob/dc03b21cf5d54c641e1d14b14fae5920fa7ba656/config/config.go#L982)
	// and will be overridden by `interceptConfigs` in `cosmos-sdk`.
	cfg.Consensus.TimeoutCommit = 500 * time.Millisecond

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.

	type CustomAppConfig struct {
		serverconfig.Config

		EVM     evmserverconfig.EVMConfig     `mapstructure:"evm"`
		JSONRPC evmserverconfig.JSONRPCConfig `mapstructure:"json-rpc"`
		TLS     evmserverconfig.TLSConfig     `mapstructure:"tls"`
	}

	customAppConfig := CustomAppConfig{
		Config:  *serverconfig.DefaultConfig(),
		EVM:     *evmserverconfig.DefaultEVMConfig(),
		JSONRPC: *evmserverconfig.DefaultJSONRPCConfig(),
		TLS:     *evmserverconfig.DefaultTLSConfig(),
	}
	customAppTemplate := serverconfig.DefaultConfigTemplate + evmserverconfig.DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}
