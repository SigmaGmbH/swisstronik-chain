package app

import (
	"fmt"
	evmkeeper "swisstronik/x/evm/keeper"

	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	"swisstronik/app/upgrades"
	v2_0_0 "swisstronik/app/upgrades/v2.0.0"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"
)

var (
	// `Upgrades` defines the upgrade handlers and store loaders for the application.
	// New upgrades should be added to this slice after they are implemented.
	Upgrades = []upgrades.Upgrade{
		v2_0_0.Upgrade,
	}
)

func (app *App) SetupHandlers(
	ek *evmkeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	cdc codec.BinaryCodec) {
	app.setUpgradeHandler(ek, clientKeeper, pk, cdc)
	app.loadUpgradeStore()
}

func (app *App) setUpgradeHandler(
	ek *evmkeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	cdc codec.BinaryCodec) {
	if app.UpgradeKeeper.HasHandler(v2_0_0.UpgradeName) {
		panic(fmt.Sprintf("Cannot register duplicate upgrade handler '%s'", v2_0_0.UpgradeName))
	}

	app.UpgradeKeeper.SetUpgradeHandler(
		v2_0_0.UpgradeName,
		v2_0_0.CreateUpgradeHandler(app.mm, ek, app.configurator, app.AccountKeeper, cdc, clientKeeper),
	)
}

func (app *App) loadUpgradeStore() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if shouldLoadUpgradeStore(app, upgradeInfo) {
		for _, upgrade := range Upgrades {
			// Use upgrade store loader for the initial loading of all stores when app starts,
			// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
			// so that new stores start with the correct version (the current height of chain),
			// instead the default which is the latest version that store last committed i.e 0 for new stores.
			if upgradeInfo.Name == upgrade.UpgradeName {
				app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
			}
		}
	}
}

func shouldLoadUpgradeStore(app *App, upgradeInfo upgradetypes.Plan) bool {
	return upgradeInfo.Name == version.Version && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height)
}
