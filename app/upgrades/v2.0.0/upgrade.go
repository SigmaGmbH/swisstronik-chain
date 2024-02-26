package v_2_0_0

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	evmkeeper "swisstronik/x/evm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	_ *evmkeeper.Keeper,
	configurator module.Configurator,
	_ authkeeper.AccountKeeper,
	_ codec.BinaryCodec,
	_ ibctmmigrations.ClientKeeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)

		ctx.Logger().Info("Running upgrade handler for " + version.Version)
		ctx.Logger().Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
