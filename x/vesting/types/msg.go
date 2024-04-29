package types

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

const TypeMsgCreateMonthlyVestingAccount = "create_monthly_vesting_account"

var _ sdk.Msg = &MsgCreateMonthlyVestingAccount{}

func NewMsgCreateMonthlyVestingAccount(fromAddress string, toAddress string, cliffDays, months int64, amount sdk.Coins) *MsgCreateMonthlyVestingAccount {
	return &MsgCreateMonthlyVestingAccount{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		CliffDays:   cliffDays,
		Months:      months,
		Amount:      amount,
	}
}

func (msg *MsgCreateMonthlyVestingAccount) Route() string {
	return RouterKey
}

func (msg *MsgCreateMonthlyVestingAccount) Type() string {
	return TypeMsgCreateMonthlyVestingAccount
}

func (msg *MsgCreateMonthlyVestingAccount) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgCreateMonthlyVestingAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateMonthlyVestingAccount) ValidateBasic() error {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid from address")
	}

	if equal := bytes.Compare(from.Bytes(), common.Address{}.Bytes()); equal == 0 {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "from address cannot be the zero address")
	}

	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid to address")
	}

	if equal := bytes.Compare(to.Bytes(), common.Address{}.Bytes()); equal == 0 {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "to address cannot be the zero address")
	}

	if msg.CliffDays <= 0 {
		return errorsmod.Wrapf(ErrInvalidParam, "cliff days cannot be zero or negative")
	}

	if msg.Months <= 1 {
		return errorsmod.Wrapf(ErrInvalidParam, "months should be at least one")
	}

	if !msg.Amount.IsAllPositive() {
		return errorsmod.Wrapf(ErrInvalidParam, "amount should be at least one")
	}

	return nil
}
