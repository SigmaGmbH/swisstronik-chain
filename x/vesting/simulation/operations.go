package simulation

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
	"swisstronik/utils"
	"swisstronik/x/vesting/keeper"
	"swisstronik/x/vesting/types"
)

func SimulateMsgCreateMonthlyVestingAccount(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		fromAccount, _ := simtypes.RandomAcc(r, accs)
		toAccount, _ := simtypes.RandomAcc(r, accs)

		amount, _ := simtypes.RandPositiveInt(r, math.NewInt(1e18))

		// Generate create monthly vesting account message
		msg := types.MsgCreateMonthlyVestingAccount{
			FromAddress: fromAccount.Address.String(),
			ToAddress:   toAccount.Address.String(),
			CliffDays:   int64(simtypes.RandIntBetween(r, 1, 30)),
			Months:      int64(simtypes.RandIntBetween(r, 1, 12)),
			Amount:      sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount)),
		}

		interfaceRegistry := codectypes.NewInterfaceRegistry()
		txConfig := tx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), tx.DefaultSignModes)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txConfig,
			Cdc:             nil,
			Msg:             &msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      fromAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: nil,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
