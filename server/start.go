// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://swisstronik/blob/main/LICENSE
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	abciserver "github.com/cometbft/cometbft/abci/server"
	tcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cometbft/cometbft/rpc/client/local"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"

	ethmetricsexp "github.com/ethereum/go-ethereum/metrics/exp"

	errorsmod "cosmossdk.io/errors"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/cosmos/cosmos-sdk/server/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"swisstronik/indexer"
	ethdebug "swisstronik/rpc/namespaces/ethereum/debug"
	"swisstronik/server/config"
	srvflags "swisstronik/server/flags"
	ethermint "swisstronik/types"
)

// DBOpener is a function to open `application.db`, potentially with customized options.
type DBOpener func(opts types.AppOptions, rootDir string, backend dbm.BackendType) (dbm.DB, error)

// StartOptions defines options that can be customized in `StartCmd`
type StartOptions struct {
	AppCreator      types.AppCreator
	DefaultNodeHome string
	DBOpener        DBOpener
}

// NewDefaultStartOptions use the default db opener provided in tm-db.
func NewDefaultStartOptions(appCreator types.AppCreator, defaultNodeHome string) StartOptions {
	return StartOptions{
		AppCreator:      appCreator,
		DefaultNodeHome: defaultNodeHome,
		DBOpener:        openDB,
	}
}

// StartCmd runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmd(opts StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with CometBFT in or out of process. By
default, the application will run with CometBFT in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent',
'pruning-keep-every', and 'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.
`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)

			// Bind flags to the Context's Viper so the app construction can set
			// options accordingly.
			err := serverCtx.Viper.BindPFlags(cmd.Flags())
			if err != nil {
				return err
			}

			_, err = server.GetPruningOptionsFromFlags(serverCtx.Viper)
			return err
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			withTM, _ := cmd.Flags().GetBool(srvflags.WithCometBFT)
			if !withTM {
				serverCtx.Logger.Info("starting ABCI without CometBFT")
				return wrapCPUProfile(serverCtx, func() error {
					return startStandAlone(serverCtx, opts)
				})
			}

			serverCtx.Logger.Info("Unlocking keyring")

			// fire unlock precess for keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if keyringBackend == keyring.BackendFile {
				_, err = clientCtx.Keyring.List()
				if err != nil {
					return err
				}
			}

			serverCtx.Logger.Info("starting ABCI with CometBFT")

			// amino is needed here for backwards compatibility of REST routes
			err = wrapCPUProfile(serverCtx, func() error {
				return startInProcess(serverCtx, clientCtx, opts)
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, opts.DefaultNodeHome, "The application home directory")
	cmd.Flags().Bool(srvflags.WithCometBFT, true, "Run abci app embedded in-process with tendermint")
	cmd.Flags().String(srvflags.Address, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(srvflags.Transport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(srvflags.TraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(server.FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photon;0.0001stake)") //nolint:lll
	cmd.Flags().IntSlice(server.FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(server.FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(server.FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(server.FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(srvflags.CPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(server.FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(server.FlagPruning, pruningtypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(server.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(server.FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')") //nolint:lll
	cmd.Flags().Uint(server.FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(server.FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune CometBFT blocks")
	cmd.Flags().String(srvflags.AppDBBackend, "", "The type of database for application and snapshots databases")

	cmd.Flags().Bool(srvflags.GRPCOnly, false, "Start the node in gRPC query only mode without CometBFT process")
	cmd.Flags().Bool(srvflags.GRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(srvflags.GRPCAddress, serverconfig.DefaultGRPCAddress, "the gRPC server address to listen on")
	cmd.Flags().Bool(srvflags.GRPCWebEnable, true, "Define if the gRPC-Web server should be enabled. (Note: gRPC must also be enabled.)")

	cmd.Flags().Bool(srvflags.RPCEnable, false, "Defines if Cosmos-sdk REST server should be enabled")
	cmd.Flags().Bool(srvflags.EnabledUnsafeCors, false, "Defines if CORS should be enabled (unsafe - use it at your own risk)")

	cmd.Flags().Bool(srvflags.JSONRPCEnable, true, "Define if the JSON-RPC server should be enabled")
	cmd.Flags().StringSlice(srvflags.JSONRPCAPI, config.GetDefaultAPINamespaces(), "Defines a list of JSON-RPC namespaces that should be enabled")
	cmd.Flags().String(srvflags.JSONRPCAddress, config.DefaultJSONRPCAddress, "the JSON-RPC server address to listen on")
	cmd.Flags().String(srvflags.JSONWsAddress, config.DefaultJSONRPCWsAddress, "the JSON-RPC WS server address to listen on")
	cmd.Flags().Uint64(srvflags.JSONRPCGasCap, config.DefaultGasCap, "Sets a cap on gas that can be used in eth_call/estimateGas unit is aphoton (0=infinite)")     //nolint:lll
	cmd.Flags().Float64(srvflags.JSONRPCTxFeeCap, config.DefaultTxFeeCap, "Sets a cap on transaction fee that can be sent via the RPC APIs (1 = default 1 photon)") //nolint:lll
	cmd.Flags().Int32(srvflags.JSONRPCFilterCap, config.DefaultFilterCap, "Sets the global cap for total number of filters that can be created")
	cmd.Flags().Duration(srvflags.JSONRPCEVMTimeout, config.DefaultEVMTimeout, "Sets a timeout used for eth_call (0=infinite)")
	cmd.Flags().Duration(srvflags.JSONRPCHTTPTimeout, config.DefaultHTTPTimeout, "Sets a read/write timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Duration(srvflags.JSONRPCHTTPIdleTimeout, config.DefaultHTTPIdleTimeout, "Sets a idle timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Bool(srvflags.JSONRPCAllowUnprotectedTxs, config.DefaultAllowUnprotectedTxs, "Allow for unprotected (non EIP155 signed) transactions to be submitted via the node's RPC when the global parameter is disabled") //nolint:lll
	cmd.Flags().Int32(srvflags.JSONRPCLogsCap, config.DefaultLogsCap, "Sets the max number of results can be returned from single `eth_getLogs` query")
	cmd.Flags().Int32(srvflags.JSONRPCBlockRangeCap, config.DefaultBlockRangeCap, "Sets the max block range allowed for `eth_getLogs` query")
	cmd.Flags().Int(srvflags.JSONRPCMaxOpenConnections, config.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener") //nolint:lll
	cmd.Flags().Bool(srvflags.JSONRPCEnableIndexer, false, "Enable the custom tx indexer for json-rpc")
	cmd.Flags().Bool(srvflags.JSONRPCAllowIndexerGap, true, "Allow block gap for the custom tx indexer for json-rpc")
	cmd.Flags().Bool(srvflags.JSONRPCEnableMetrics, false, "Define if EVM rpc metrics server should be enabled")

	cmd.Flags().String(srvflags.EVMTracer, config.DefaultEVMTracer, "the EVM tracer type to collect execution traces from the EVM transaction execution (json|struct|access_list|markdown)") //nolint:lll
	cmd.Flags().Uint64(srvflags.EVMMaxTxGasWanted, config.DefaultMaxTxGasWanted, "the gas wanted for each eth tx returned in ante handler in check tx mode")                                 //nolint:lll

	cmd.Flags().String(srvflags.TLSCertPath, "", "the cert.pem file path for the server TLS configuration")
	cmd.Flags().String(srvflags.TLSKeyPath, "", "the key.pem file path for the server TLS configuration")

	cmd.Flags().Uint64(server.FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(server.FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")

	// add support for all CometBFT-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startStandAlone(svrCtx *server.Context, opts StartOptions) error {
	addr := svrCtx.Viper.GetString(srvflags.Address)
	transport := svrCtx.Viper.GetString(srvflags.Transport)
	home := svrCtx.Viper.GetString(flags.FlagHome)

	db, err := opts.DBOpener(svrCtx.Viper, home, server.GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			svrCtx.Logger.Error("error closing db", "error", err.Error())
		}
	}()

	traceWriterFile := svrCtx.Viper.GetString(srvflags.TraceStore)
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return err
	}

	app := opts.AppCreator(svrCtx.Logger, db, traceWriter, svrCtx.Viper)
	defer func() {
		if err := app.Close(); err != nil {
			svrCtx.Logger.Error("close application failed", "error", err.Error())
		}
	}()

	config, err := config.GetConfig(svrCtx.Viper)
	if err != nil {
		svrCtx.Logger.Error("failed to get server config", "error", err.Error())
		return err
	}

	if err := config.ValidateBasic(); err != nil {
		svrCtx.Logger.Error("invalid server config", "error", err.Error())
		return err
	}

	_, err = startTelemetry(config)
	if err != nil {
		return err
	}

	cmtApp := server.NewCometABCIWrapper(app)
	svr, err := abciserver.NewServer(addr, transport, cmtApp)
	if err != nil {
		return fmt.Errorf("error creating listener: %v", err)
	}

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger.With("server", "abci")})
	g, ctx := getCtx(svrCtx, false)

	g.Go(func() error {
		if err := svr.Start(); err != nil {
			svrCtx.Logger.Error("failed to start out-of-process ABCI server", "err", err)
			return err
		}

		// Wait for the calling process to be canceled or close the provided context,
		// so we can gracefully stop the ABCI server.
		<-ctx.Done()
		svrCtx.Logger.Info("stopping the ABCI server...")
		return svr.Stop()
	})

	return g.Wait()
}

// legacyAminoCdc is used for the legacy REST API
func startInProcess(svrCtx *server.Context, clientCtx client.Context, opts StartOptions) (err error) {
	cfg := svrCtx.Config
	home := cfg.RootDir
	logger := svrCtx.Logger
	g, ctx := getCtx(svrCtx, true)

	db, err := opts.DBOpener(svrCtx.Viper, home, server.GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		logger.Error("failed to open DB", "error", err.Error())
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.With("error", err).Error("error closing db")
		}
	}()

	traceWriterFile := svrCtx.Viper.GetString(srvflags.TraceStore)
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		logger.Error("failed to open trace writer", "error", err.Error())
		return err
	}

	config, err := config.GetConfig(svrCtx.Viper)
	if err != nil {
		logger.Error("failed to get server config", "error", err.Error())
		return err
	}

	if err := config.ValidateBasic(); err != nil {
		logger.Error("invalid server config", "error", err.Error())
		return err
	}

	app := opts.AppCreator(svrCtx.Logger, db, traceWriter, svrCtx.Viper)
	defer func() {
		if err := app.Close(); err != nil {
			logger.Error("close application failed", "error", err.Error())
		}
	}()

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		logger.Error("failed load or gen node key", "error", err.Error())
		return err
	}

	genDocProvider := GenDocProvider(cfg)

	var (
		tmNode   *node.Node
		gRPCOnly = svrCtx.Viper.GetBool(srvflags.GRPCOnly)
	)

	if gRPCOnly {
		logger.Info("starting node in query only mode; CometBFT is disabled")
		config.GRPC.Enable = true
		config.JSONRPC.EnableIndexer = false
	} else {
		logger.Info("starting node with ABCI CometBFT in-process")

		cmtApp := server.NewCometABCIWrapper(app)
		tmNode, err = node.NewNodeWithContext(
			ctx,
			cfg,
			pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
			nodeKey,
			proxy.NewLocalClientCreator(cmtApp),
			genDocProvider,
			cmtcfg.DefaultDBProvider,
			node.DefaultMetricsProvider(cfg.Instrumentation),
			servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger.With("server", "node")},
		)
		if err != nil {
			logger.Error("failed init node", "error", err.Error())
			return err
		}

		if err := tmNode.Start(); err != nil {
			logger.Error("failed start tendermint server", "error", err.Error())
			return err
		}

		defer func() {
			if tmNode.IsRunning() {
				_ = tmNode.Stop()
			}
		}()
	}

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC or JSONRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local tendermint RPC client.
	if (config.API.Enable || config.GRPC.Enable || config.JSONRPC.Enable || config.JSONRPC.EnableIndexer) && tmNode != nil {
		clientCtx = clientCtx.WithClient(local.New(tmNode))

		app.RegisterTxService(clientCtx)
		app.RegisterTendermintService(clientCtx)
		app.RegisterNodeService(clientCtx, config.Config)
	}

	metrics, err := startTelemetry(config)
	if err != nil {
		return err
	}

	// Enable metrics if JSONRPC is enabled and --metrics is passed
	// Flag not added in config to avoid user enabling in config without passing in CLI
	if config.JSONRPC.Enable && svrCtx.Viper.GetBool(srvflags.JSONRPCEnableMetrics) {
		ethmetricsexp.Setup(config.JSONRPC.MetricsAddress)
	}

	var idxer ethermint.EVMTxIndexer
	if config.JSONRPC.EnableIndexer {
		idxDB, err := OpenIndexerDB(home, server.GetAppDBBackend(svrCtx.Viper))
		if err != nil {
			logger.Error("failed to open evm indexer DB", "error", err.Error())
			return err
		}

		idxLogger := logger.With("indexer", "evm")
		idxer = indexer.NewKVIndexer(idxDB, idxLogger, clientCtx)
		indexerService := NewEVMIndexerService(idxer, clientCtx.Client.(rpcclient.Client), config.JSONRPC.AllowIndexerGap)
		indexerService.SetLogger(servercmtlog.CometLoggerWrapper{Logger: idxLogger})

		g.Go(func() error {
			return indexerService.Start()
		})
	}

	if config.API.Enable || config.JSONRPC.Enable {
		chainID := svrCtx.Viper.GetString(flags.FlagChainID)
		if chainID == "" {
			genDoc, err := genDocProvider()
			if err != nil {
				return err
			}
			chainID = genDoc.ChainID
		}
		clientCtx = clientCtx.
			WithHomeDir(home).
			WithChainID(chainID)
	}

	grpcSrv, clientCtx, err := startGrpcServer(ctx, svrCtx, clientCtx, g, config.GRPC, app)
	if err != nil {
		return err
	}
	if grpcSrv != nil {
		defer grpcSrv.GracefulStop()
	}

	apiSrv := startAPIServer(ctx, svrCtx, clientCtx, g, config.Config, app, grpcSrv, metrics)
	if apiSrv != nil {
		defer apiSrv.Close()
	}

	clientCtx, httpSrv, httpSrvDone, err := StartJSONRPCServer(svrCtx, clientCtx, g, config, genDocProvider, idxer)
	if httpSrv != nil {
		defer func() {
			shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFn()
			if err := httpSrv.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP server shutdown produced a warning", "error", err.Error())
			} else {
				logger.Info("HTTP server shut down, waiting 5 sec")
				select {
				case <-time.Tick(5 * time.Second):
				case <-httpSrvDone:
				}
			}
		}()
	}

	// At this point it is safe to block the process if we're in query only mode as
	// we do not need to start Rosetta or handle any CometBFT related processes.

	return g.Wait()
}

func openDB(_ types.AppOptions, rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

// OpenIndexerDB opens the custom eth indexer db, using the same db backend as the main app
func OpenIndexerDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("evmindexer", backendType, dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile == "" {
		return
	}

	filePath := filepath.Clean(traceWriterFile)
	return os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o600,
	)
}

func startTelemetry(cfg config.Config) (*telemetry.Metrics, error) {
	if !cfg.Telemetry.Enabled {
		return nil, nil
	}
	return telemetry.New(cfg.Telemetry)
}

// wrapCPUProfile runs callback in a goroutine, then wait for quit signals.
func wrapCPUProfile(ctx *server.Context, callback func() error) error {
	if cpuProfile := ctx.Viper.GetString(srvflags.CPUProfile); cpuProfile != "" {
		fp, err := ethdebug.ExpandHome(cpuProfile)
		if err != nil {
			ctx.Logger.Debug("failed to get filepath for the CPU profile file", "error", err.Error())
			return err
		}
		f, err := os.Create(fp)
		if err != nil {
			return err
		}

		ctx.Logger.Info("starting CPU profiler", "profile", cpuProfile)
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}

		defer func() {
			ctx.Logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()
			if err := f.Close(); err != nil {
				ctx.Logger.Info("failed to close cpu-profile file", "profile", cpuProfile, "err", err.Error())
			}
		}()
	}

	return callback()
}

func getCtx(svrCtx *server.Context, block bool) (*errgroup.Group, context.Context) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	server.ListenForQuitSignals(g, block, cancelFn, svrCtx.Logger)
	return g, ctx
}

func startGrpcServer(
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	g *errgroup.Group,
	config serverconfig.GRPCConfig,
	app types.Application,
) (*grpc.Server, client.Context, error) {
	if !config.Enable {
		// return grpcServer as nil if gRPC is disabled
		return nil, clientCtx, nil
	}
	_, _, err := net.SplitHostPort(config.Address)
	if err != nil {
		return nil, clientCtx, errorsmod.Wrapf(err, "invalid grpc address %s", config.Address)
	}

	maxSendMsgSize := config.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = serverconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := config.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = serverconfig.DefaultGRPCMaxRecvMsgSize
	}

	// if gRPC is enabled, configure gRPC client for gRPC gateway and json-rpc
	grpcClient, err := grpc.Dial(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize),
		),
	)
	if err != nil {
		return nil, clientCtx, err
	}
	// Set `GRPCClient` to `clientCtx` to enjoy concurrent grpc query.
	// only use it if gRPC server is enabled.
	clientCtx = clientCtx.WithGRPCClient(grpcClient)
	svrCtx.Logger.Debug("gRPC client assigned to client context", "address", config.Address)

	grpcSrv, err := servergrpc.NewGRPCServer(clientCtx, app, config)
	if err != nil {
		return nil, clientCtx, err
	}

	// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
	// that the server is gracefully shut down.
	g.Go(func() error {
		return servergrpc.StartGRPCServer(ctx, svrCtx.Logger.With("module", "grpc-server"), config, grpcSrv)
	})
	return grpcSrv, clientCtx, nil
}

func startAPIServer(
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	g *errgroup.Group,
	svrCfg serverconfig.Config,
	app types.Application,
	grpcSrv *grpc.Server,
	metrics *telemetry.Metrics,
) *api.Server {
	if !svrCfg.API.Enable {
		return nil
	}

	apiSrv := api.New(clientCtx, svrCtx.Logger.With("server", "api"), grpcSrv)
	app.RegisterAPIRoutes(apiSrv, svrCfg.API)

	if svrCfg.Telemetry.Enabled {
		apiSrv.SetTelemetry(metrics)
	}

	g.Go(func() error {
		return apiSrv.Start(ctx, svrCfg)
	})
	return apiSrv
}

func StartJSONRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	g *errgroup.Group,
	config config.Config,
	genDocProvider node.GenesisDocProvider,
	idxer ethermint.EVMTxIndexer,
) (ctx client.Context, httpSrv *http.Server, httpSrvDone chan struct{}, err error) {
	ctx = clientCtx
	if !config.JSONRPC.Enable {
		return
	}

	genDoc, err := genDocProvider()
	if err != nil {
		return ctx, httpSrv, httpSrvDone, err
	}

	ctx = clientCtx.WithChainID(genDoc.ChainID)
	g.Go(func() error {
		httpSrv, httpSrvDone, err = StartJSONRPC(svrCtx, clientCtx, g, &config, idxer)
		return err
	})
	return
}

// returns a function which returns the genesis doc from the genesis file.
func GenDocProvider(cfg *cmtcfg.Config) func() (*cmttypes.GenesisDoc, error) {
	return func() (*cmttypes.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}
}
