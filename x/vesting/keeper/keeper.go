package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"swisstronik/x/vesting/types"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

func NewKeeper(
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetMonthlyVestingAccount(ctx sdk.Context, address sdk.AccAddress) (*types.MonthlyVestingAccount, error) {
	acc := k.accountKeeper.GetAccount(ctx, address)
	if acc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrUnknownAddress, "account at %s does not exist", address.String())
	}

	vestingAccount, ok := acc.(*types.MonthlyVestingAccount)
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrNotFoundVestingAccount, address.String())
	}

	return vestingAccount, nil
}
