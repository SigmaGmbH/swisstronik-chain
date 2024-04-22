package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgAddOperator(operatorAddress, newOperatorAddress string) MsgAddOperator {
	return MsgAddOperator{
		Signer:   operatorAddress,
		Operator: newOperatorAddress,
	}
}

func (msg *MsgAddOperator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddOperator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	return nil
}

func (msg *MsgAddOperator) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func NewMsgRemoveOperator(operatorAddress, newOperatorAddress string) MsgRemoveOperator {
	return MsgRemoveOperator{
		Signer:   operatorAddress,
		Operator: newOperatorAddress,
	}
}

func (msg *MsgRemoveOperator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveOperator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	return nil
}

func (msg *MsgRemoveOperator) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func NewSetIssuerDetailsMsg(operatorAddress, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgSetIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
	}
	return MsgSetIssuerDetails{
		Signer:  operatorAddress,
		Issuer:  issuerAddress,
		Details: &issuerDetails,
	}
}

func (msg *MsgSetIssuerDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetIssuerDetails) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgSetIssuerDetails) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func NewUpdateIssuerDetailsMsg(operatorAddress, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgUpdateIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
	}
	return MsgUpdateIssuerDetails{
		Signer:  operatorAddress,
		Issuer:  issuerAddress,
		Details: &issuerDetails,
	}
}

func (msg *MsgUpdateIssuerDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateIssuerDetails) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgUpdateIssuerDetails) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func NewRemoveIssuerMsg(operator, issuerAddress string) MsgRemoveIssuer {
	return MsgRemoveIssuer{
		Signer: operator,
		Issuer: issuerAddress,
	}
}

func (msg *MsgRemoveIssuer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveIssuer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgRemoveIssuer) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}
