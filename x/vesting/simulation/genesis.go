package simulation

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"swisstronik/x/vesting/types"
)

func RandomGenesisAccounts(simState *module.SimulationState) authtypes.GenesisAccounts {
	genesisAccs := make(authtypes.GenesisAccounts, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		bacc := authtypes.NewBaseAccountWithAddress(acc.Address)

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if !(int64(i) > simState.NumBonded && simState.Rand.Intn(100) < 50) {
			genesisAccs[i] = bacc
			continue
		}

		initialVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, simState.InitialStake.Sub(sdkmath.NewInt(simState.Rand.Int63n(10)))))
		var endTime int64

		startTime := simState.GenTimestamp.Unix()

		// Allow for some vesting accounts to vest very quickly while others very slowly.
		if simState.Rand.Intn(100) < 50 {
			endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*24*30))))
		} else {
			endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*12))))
		}

		bva := vestingtypes.NewBaseVestingAccount(bacc, initialVesting, endTime)

		rnd := simState.Rand.Intn(300)
		if rnd < 100 {
			genesisAccs[i] = vestingtypes.NewContinuousVestingAccountRaw(bva, startTime)
		} else if rnd < 200 {
			genesisAccs[i] = vestingtypes.NewDelayedVestingAccountRaw(bva)
		} else {
			var (
				cliffDays = int64(simulation.RandIntBetween(simState.Rand, 0, 30))
				months    = int64(simulation.RandIntBetween(simState.Rand, 1, 12))
			)
			genesisAccs[i] = types.NewMonthlyVestingAccount(bacc, initialVesting, startTime, cliffDays, months)
		}
	}

	return genesisAccs
}
