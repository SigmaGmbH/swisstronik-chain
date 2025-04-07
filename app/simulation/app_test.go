package simulation

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	swtrapp "swisstronik/app"
	"swisstronik/encoding"
	"swisstronik/tests"

	commontypes "swisstronik/types"
)

func TestFullAppSimulation(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = commontypes.PrefixedChainID
	config.Commit = true // commit by default
	config.NumBlocks = 5 // default value is 500

	db, dir, logger, skip, err := simtestutil.SetupSimulation(
		config,
		"leveldb-app-sim",
		"Simulation",
		true, // simcli.FlagVerboseValue,
		true, // simcli.FlagEnabledValue,
	)
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = swtrapp.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := swtrapp.New(
		logger,
		db, nil, true, map[int64]bool{},
		swtrapp.DefaultNodeHome, 5,
		encoding.MakeConfig(swtrapp.ModuleBasics),
		appOptions,
		fauxMerkleModeOpt,
		baseapp.SetChainID(config.ChainID),
	)
	require.Equal(t, "swisstronik", app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), simGenesisState()),
		tests.RandomSimulationEthAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
