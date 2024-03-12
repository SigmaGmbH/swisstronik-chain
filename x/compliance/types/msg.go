package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (msg *MsgSetVerificationData) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetVerificationData) ValidateBasic() error {
	if len(msg.Data.Entries) == 0 {
		return ErrBasicValidation
	}

	return nil
}

func (msg *MsgSetVerificationData) GetSigners() []sdk.AccAddress {
	msg.
}
