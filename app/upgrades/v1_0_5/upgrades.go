package v1_0_5

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"swisstronik/utils"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	bankKeeper bankkeeper.Keeper,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		if utils.IsTestnet(ctx.ChainID()) {
			// Fund faucet coins
			const amount = 5000000
			var faucetAddresses = []string{
				"swtr10t6y08uyqvf6u3r73u83xawkas3tfq7hnt84xq",
				"swtr14tvy0a64rdgte8vms6qc9ea0t0g2r94rnl08n8",
				"swtr17fs5hku6f69cl07kea4qqlwnxvtnrguruxmz9f",
				"swtr1rds7cwl2rkrx6c6f6gcxghz4lx9h59tkj88nvv",
				"swtr1rtecg6ewfn2q0fs0s6yzsrwhm32du3ps3wjqhu",
			}
			for _, address := range faucetAddresses {
				faucetCoins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(amount, 18)))
				err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, faucetCoins)
				if err != nil {
					return vm, err
				}
				err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.MustAccAddressFromBech32(address), faucetCoins)
				if err != nil {
					return vm, err
				}
				ctx.Logger().Info("Sent faucet coins", "address", address, "faucetCoins", faucetCoins.String())
			}
		}

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}
		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
