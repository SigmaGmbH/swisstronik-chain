package v_1_0_0

import (
	"context"

	swissapp "swisstronik/app"
	didmoduletypes "swisstronik/x/did/types"
	evmmoduletypes "swisstronik/x/evm/types"
	feemarketmoduletypes "swisstronik/x/feemarket/types"
	vestmoduletypes "swisstronik/x/vesting/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	evmkeeper "swisstronik/x/evm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func CreateUpgradeHandler(
	app *swissapp.App,
	mm *module.Manager,
	ek *evmkeeper.Keeper,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	cdc codec.BinaryCodec,
	clientKeeper ibctmmigrations.ClientKeeper,
) upgradetypes.UpgradeHandler {
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
		default:
			continue
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}

	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	return func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)

		ctx.Logger().Info("Running upgrade handler for " + version.Version)

		baseapp.MigrateParams(ctx, baseAppLegacySS, &app.ConsensusParamsKeeper.ParamsStore)

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
		ctx.Logger().Info("adding EIP 3855 to EVM parameters")
		err := EnableEIPs(ctx, ek, 3855)
		if err != nil {
			ctx.Logger().Error("error while enabling EIPs", "error", err)
		}

		ctx.Logger().Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// EnableEIPs enables the given EIPs in the EVM parameters.
func EnableEIPs(ctx sdk.Context, ek *evmkeeper.Keeper, eips ...int64) error {
	evmParams := ek.GetParams(ctx)
	evmParams.ExtraEIPs = append(evmParams.ExtraEIPs, eips...)

	return ek.SetParams(ctx, evmParams)
}
