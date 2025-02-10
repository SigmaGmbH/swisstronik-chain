package v1_0_7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	compliancemoduletypes "swisstronik/x/compliance/types"
	evmkeeper "swisstronik/x/evm/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	complianceKeeper compliancemoduletypes.ComplianceKeeper,
	evmkeeper *evmkeeper.Keeper,
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

		// Set bytecode of Arachnid Deterministic Deployment Proxy
		// We use this dirty hack since non-EIP-155 transactions are not allowed
		// in Swisstronik Network
		proxyAddress := common.HexToAddress("0x4e59b44847b379578588920cA78FbF26c0B4956C")
		codeBytes, err := hexutil.Decode("0x7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf3")
		if err != nil {
			return vm, err
		}

		if err = evmkeeper.SetAccountCode(ctx, proxyAddress, codeBytes); err != nil {
			return vm, err
		}

		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
