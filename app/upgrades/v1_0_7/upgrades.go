package v1_0_7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	compliancemoduletypes "swisstronik/x/compliance/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	complianceKeeper compliancemoduletypes.ComplianceKeeper,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		var migrationError error

		// Link verification id -> holder
		complianceKeeper.IterateAddressDetails(ctx, func(addr sdk.AccAddress) (continue_ bool) {
			addressDetails, err := complianceKeeper.GetAddressDetails(ctx, addr)
			if err != nil {
				migrationError = err
				return false
			}

			for _, verification := range addressDetails.Verifications {
				if err = complianceKeeper.LinkVerificationToHolder(ctx, addr, verification.VerificationId); err != nil {
					migrationError = err
					return false
				}
			}
			return true
		})
		if migrationError != nil {
			return vm, migrationError
		}

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
