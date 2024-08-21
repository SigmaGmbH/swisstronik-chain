package ante

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
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
	accountSpendable := make(map[string]*big.Int)
	denom := vtd.ek.GetParams(ctx).EvmDenom

	// Aggregate all the spending amounts per each account.
	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgHandleTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgHandleTx)(nil))
		}

		_, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, errorsmod.Wrap(err, "failed to unpack tx data")
		}

		from := msgEthTx.GetFrom()
		account := vtd.ak.GetAccount(ctx, from)
		// Only check if account is vesting account
		if _, ok = account.(*vestingmoduletypes.MonthlyVestingAccount); !ok {
			continue
		}
		fromStr := from.String()

		spendable := vtd.bk.SpendableCoin(ctx, from, denom)
		if existing, ok := accountSpendable[fromStr]; ok {
			accountSpendable[fromStr] = new(big.Int).Add(existing, spendable.Amount.BigInt())
		} else {
			accountSpendable[fromStr] = spendable.Amount.BigInt()
		}
	}

	// Check if all the vesting accounts have enough spendable amount as what they need.
	for _, msg := range tx.GetMsgs() {
		msgEthTx, _ := msg.(*evmtypes.MsgHandleTx)
		txData, _ := evmtypes.UnpackTxData(msgEthTx.Data)

		from := msgEthTx.GetFrom()
		account := vtd.ak.GetAccount(ctx, from)
		// Only check if account is vesting account
		if _, ok := account.(*vestingmoduletypes.MonthlyVestingAccount); !ok {
			continue
		}
		fromStr := from.String()
		value := txData.GetValue()
		if accountSpendable[fromStr].Cmp(value) < 0 {
			return ctx, errorsmod.Wrapf(vestingmoduletypes.ErrInsufficientUnlockedCoins, "%s < %s", accountSpendable[fromStr].String(), value.String())
		}
		accountSpendable[fromStr] = new(big.Int).Sub(accountSpendable[fromStr], value)
	}
	return next(ctx, tx, simulate)
}
