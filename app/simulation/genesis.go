package simulation

import (
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"swisstronik/app"
	"swisstronik/encoding"
	"swisstronik/utils"
	feemarkettypes "swisstronik/x/feemarket/types"
)

// NewDefaultGenesisState generates the default state for the application.
func simGenesisState() simapp.GenesisState {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	genesisState := app.ModuleBasics.DefaultGenesis(encCfg.Codec)

	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.BondDenom = utils.BaseDenom
	genesisState[stakingtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(stakingGenesis)

	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Params.MintDenom = utils.BaseDenom
	genesisState[minttypes.ModuleName] = encCfg.Codec.MustMarshalJSON(mintGenesis)

	govGenesis := govtypes.DefaultGenesisState()
	govGenesis.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1000)))

	// NOTE, Set feemarket disabled, so simulation transaction does not check ante handler for dynamic fee checker,
	// since all the automatically generated accounts are having random coin amounts with `stake` denom by default.
	// While the evm/feemarket modules allow gas fee only with `aswtr` denom, in order to enable transaction with `aswtr` denom,
	// there need a tricky solution that replace `stake` denom with `aswtr` base denom.
	// This problem only happens with the function call `simtestutil.AppStateFn` in `app_test.go`, we will replace it later
	// with the function that generates coins with `aswtr` instead of `stake` denom.

	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.NoBaseFee = true
	genesisState[feemarkettypes.ModuleName] = encCfg.Codec.MustMarshalJSON(feemarketGenesis)

	return genesisState
}
