package app

import (
	didmoduletypes "swisstronik/x/did/types"
	evmkeeper "swisstronik/x/evm/keeper"
	evmmoduletypes "swisstronik/x/evm/types"
	feemarketmoduletypes "swisstronik/x/feemarket/types"
	vestmoduletypes "swisstronik/x/vesting/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	m "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
)

func SetupHandlers(
	app *App,
	ek *evmkeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	cdc codec.BinaryCodec) {
	setUpgradeHandler(app, ek, clientKeeper, pk, cdc)

	loadUpgradeStore(app)
}

func setUpgradeHandler(
	app *App,
	ek *evmkeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	cdc codec.BinaryCodec) {
	// Set param key table for params module migration
	for _, subspace := range app.ParamsKeeper.GetSubspaces() {
		subspace := subspace

		app.Logger().Info("Setting up upgrade handler for " + subspace.Name())

		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable() //nolint:staticcheck
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
		case didmoduletypes.ModuleName:
			keyTable = didmoduletypes.ParamKeyTable()
		case evmmoduletypes.ModuleName:
			keyTable = evmmoduletypes.ParamKeyTable()
		case feemarketmoduletypes.ModuleName:
			keyTable = feemarketmoduletypes.ParamKeyTable()
		case vestmoduletypes.ModuleName:
			keyTable = vestmoduletypes.ParamKeyTable()
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable()
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}

	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	app.UpgradeKeeper.SetUpgradeHandler(
		version.Version,
		func(ctx sdk.Context, plan upgradetypes.Plan, vm m.VersionMap) (m.VersionMap, error) {
			app.Logger().Info("Running upgrade handler for " + version.Version)

			// Migrate Tendermint consensus parameters from x/params module to a
			// dedicated x/consensus module.
			baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper)

			// Include this when migrating to ibc-go v7 (optional)
			// source: https://github.com/cosmos/ibc-go/blob/v7.2.0/docs/migrations/v6-to-v7.md
			// prune expired tendermint consensus states to save storage space
			if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, cdc, clientKeeper); err != nil {
				return nil, err
			}
			// !! ATTENTION !!

			// Add EIP contained in Shanghai hard fork to the extra EIPs
			// in the EVM parameters. This enables using the PUSH0 opcode and
			// thus supports Solidity v0.8.20.
			//
			app.Logger().Info("adding EIP 3855 to EVM parameters")
			err := EnableEIPs(ctx, ek, 3855)
			if err != nil {
				app.Logger().Error("error while enabling EIPs", "error", err)
			}

			app.Logger().Debug("running module migrations ...")
			return app.mm.RunMigrations(ctx, app.configurator, vm)
		},
	)
}

// EnableEIPs enables the given EIPs in the EVM parameters.
func EnableEIPs(ctx sdk.Context, ek *evmkeeper.Keeper, eips ...int64) error {
	evmParams := ek.GetParams(ctx)
	evmParams.ExtraEIPs = append(evmParams.ExtraEIPs, eips...)

	return ek.SetParams(ctx, evmParams)
}

func loadUpgradeStore(app *App) {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if shouldLoadUpgradeStore(app, upgradeInfo) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				didmoduletypes.StoreKey,
				consensusparamtypes.StoreKey,
				crisistypes.ModuleName,
			},
		}
		// Use upgrade store loader for the initial loading of all stores when app starts,
		// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
		// so that new stores start with the correct version (the current height of chain),
		// instead the default which is the latest version that store last committed i.e 0 for new stores.
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func shouldLoadUpgradeStore(app *App, upgradeInfo upgradetypes.Plan) bool {
	return upgradeInfo.Name == version.Version && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height)
}
