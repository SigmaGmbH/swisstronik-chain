package simulation

import (
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"swisstronik/x/vesting/keeper"
	"swisstronik/x/vesting/types"
)

const (
	OpWeightMsgCreateMonthlyVestingAccount = "op_weight_msg_create_monthly_vesting_account"

	DefaultWeightMsgCreateMonthlyVestingAccount int = 100
)

func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simulation.WeightedOperations {
	var (
		weightMsgCreateMonthlyVestingAccount int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateMonthlyVestingAccount, &weightMsgCreateMonthlyVestingAccount, nil,
		func(_ *rand.Rand) {
			weightMsgCreateMonthlyVestingAccount = DefaultWeightMsgCreateMonthlyVestingAccount
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateMonthlyVestingAccount,
			SimulateMsgCreateMonthlyVestingAccount(ak, bk, k),
		),
	}
}

func RandomizedGenState(simState *module.SimulationState) {
	authsims.RandomizedGenState(simState, authsims.RandomGenesisAccounts)
}
