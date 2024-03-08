package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateMonthlyVestingAccount = "create_monthly_vesting_account"

var _ sdk.Msg = &MsgCreateMonthlyVestingAccount{}

func NewMsgCreateMonthlyVestingAccount(creator string, toAddress string, startTime int64, amount sdk.Coins, month int64) *MsgCreateMonthlyVestingAccount {
	return &MsgCreateMonthlyVestingAccount{
		Creator:   creator,
		ToAddress: toAddress,
		StartTime: startTime,
		Amount:    amount,
		Month:     month,
	}
}

func (msg *MsgCreateMonthlyVestingAccount) Route() string {
	return RouterKey
}

func (msg *MsgCreateMonthlyVestingAccount) Type() string {
	return TypeMsgCreateMonthlyVestingAccount
}

func (msg *MsgCreateMonthlyVestingAccount) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateMonthlyVestingAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateMonthlyVestingAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
