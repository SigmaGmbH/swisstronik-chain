package v_1_0_0

import (
	"swisstronik/app/upgrades"
	didmoduletypes "swisstronik/x/did/types"

	storetypes "cosmossdk.io/store/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
)

const (
	UpgradeName = "v1.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName: UpgradeName,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			didmoduletypes.StoreKey,
			consensusparamtypes.StoreKey,
			crisistypes.ModuleName,
			icacontrollertypes.StoreKey,
		},
	},
}
