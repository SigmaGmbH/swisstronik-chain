package v1_0_8

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	stakingKeeper stakingkeeper.Keeper,
	storeKey *types.KVStoreKey,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		validators := stakingKeeper.GetAllValidators(ctx)
		ctx.Logger().Info("Updating voting power")
		for _, validator := range validators {
			store := ctx.KVStore(storeKey)
			
			deleted := false

			iterator := sdk.KVStorePrefixIterator(store, stakingtypes.ValidatorsByPowerIndexKey)
			defer iterator.Close()

			for ; iterator.Valid(); iterator.Next() {
				valAddr := stakingtypes.ParseValidatorPowerRankKey(iterator.Key())
				if bytes.Equal(valAddr, validator.GetOperator()) {
					if deleted {
						panic("Found duplicate power index key")
					} else {
						deleted = true
					}

					store.Delete(iterator.Key())
				}
			}

			stakingKeeper.SetValidatorByPowerIndex(ctx, validator)
			ctx.Logger().Info("Set the validator successfully...")

			_, err := stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			ctx.Logger().Info("Set validator set update...")
			if err != nil {
				panic(err)
			}

			ctx.Logger().Info("Validator updated...")

		}

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
