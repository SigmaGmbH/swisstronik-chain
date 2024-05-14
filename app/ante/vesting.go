package ante

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	evmtypes "swisstronik/x/evm/types"
	vestingmoduletypes "swisstronik/x/vesting/types"
)

// EthVestingTransactionDecorator validates if monthly vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
	bk evmtypes.BankKeeper
	ek EVMKeeper
}

// EthVestingExpenseTracker tracks both the total transaction value to be sent across Ethereum
// messages and the maximum spendable value for a given account.
type EthVestingExpenseTracker struct {
	// Total is the total value to be spent across a transaction with one or more Ethereum message calls
	Total *big.Int
	// Spendable is the maximum value that can be spent
	Spendable *big.Int
}

// NewEthVestingTransactionDecorator returns a new EthVestingTransactionDecorator.
func NewEthVestingTransactionDecorator(ak evmtypes.AccountKeeper, bk evmtypes.BankKeeper, ek EVMKeeper) EthVestingTransactionDecorator {
	return EthVestingTransactionDecorator{
		ak: ak,
		bk: bk,
		ek: ek,
	}
}

// AnteHandle validates that monthly vesting account has sufficient unlocked balances to cover the transaction
// during the cliff and vesting period.
func (vtd EthVestingTransactionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	accountExpenses := make(map[string]*EthVestingExpenseTracker)
	denom := vtd.ek.GetParams(ctx).EvmDenom

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgHandleTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgHandleTx)(nil))
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, errorsmod.Wrap(err, "failed to unpack tx data")
		}

		from := msgEthTx.GetFrom()
		value := txData.GetValue()
		acc := vtd.ak.GetAccount(ctx, from)
		if acc == nil {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress, "account %s does not exist", acc)
		}

		if err = CheckVesting(ctx, vtd.bk, acc, accountExpenses, value, denom); err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

// CheckVesting checks if the account is monthly vesting account and if so,
// checks that the account has sufficient unlocked balances to cover transaction.
func CheckVesting(
	ctx sdk.Context,
	bankKeeper evmtypes.BankKeeper,
	account authtypes.AccountI,
	accountExpenses map[string]*EthVestingExpenseTracker,
	addedExpense *big.Int,
	denom string,
) error {
	vestingAccount, ok := account.(*vestingmoduletypes.MonthlyVestingAccount)
	if !ok {
		return nil
	}

	// Check to make sure that the account does not exceed its spendable balances.
	// This transaction would fail in processing, so we should prevent it from
	// move part the AnteHandler.
	expenses, err := UpdateAccountExpense(ctx, bankKeeper, accountExpenses, vestingAccount, addedExpense, denom)
	if err != nil {
		return err
	}

	total := expenses.Total
	spendable := expenses.Spendable

	if total.Cmp(spendable) > 0 {
		return errorsmod.Wrapf(vestingmoduletypes.ErrInsufficientUnlockedCoins, "%s < %s", spendable.String(), total.String())
	}
	return nil
}

// UpdateAccountExpense updates or sets the total spend for the given account, then
// returns the new expense.
func UpdateAccountExpense(
	ctx sdk.Context,
	bankKeeper evmtypes.BankKeeper,
	accountExpenses map[string]*EthVestingExpenseTracker,
	account *vestingmoduletypes.MonthlyVestingAccount,
	addedExpense *big.Int,
	denom string,
) (*EthVestingExpenseTracker, error) {
	address := account.GetAddress()
	addrStr := address.String()

	expenses, ok := accountExpenses[addrStr]
	// if expense tracker exists, update it by adding new expense
	if ok {
		expenses.Total = new(big.Int).Add(expenses.Total, addedExpense)
		return expenses, nil
	}

	balance := bankKeeper.GetBalance(ctx, address, denom)
	if balance.IsZero() {
		return nil, errorsmod.Wrapf(errortypes.ErrInsufficientFunds,
			"account has no balance to execute transaction: %s", addrStr)
	}

	lockedBalances := account.LockedCoins(ctx.BlockTime())
	ok, lockedBalance := lockedBalances.Find(denom)
	if !ok {
		lockedBalance = sdk.NewCoin(denom, math.ZeroInt())
	}

	spendableValue := big.NewInt(0)
	if spendableBalance, err := balance.SafeSub(lockedBalance); err == nil {
		spendableValue = spendableBalance.Amount.BigInt()
	}

	expenses = &EthVestingExpenseTracker{
		Total:     addedExpense,
		Spendable: spendableValue,
	}
	accountExpenses[addrStr] = expenses

	return expenses, nil
}
