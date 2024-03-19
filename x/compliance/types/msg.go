package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewSetIssuerDetailsMsg(operator sdk.Address, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgSetIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
		Operator:    operator.Bytes(),
	}
	return MsgSetIssuerDetails{
		Operator:      operator.String(),
		IssuerAddress: issuerAddress,
		Details:       &issuerDetails,
	}
}

func (msg *MsgSetIssuerDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetIssuerDetails) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgSetIssuerDetails) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func NewUpdateIssuerDetailsMsg(operator sdk.Address, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgUpdateIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
		Operator:    operator.Bytes(),
	}
	return MsgUpdateIssuerDetails{
		Operator:      operator.String(),
		IssuerAddress: issuerAddress,
		Details:       &issuerDetails,
	}
}

func (msg *MsgUpdateIssuerDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateIssuerDetails) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgUpdateIssuerDetails) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func NewRemoveIssuerDetailsMsg(operator sdk.Address, issuerAddress string) MsgRemoveIssuer {
	return MsgRemoveIssuer{
		Operator:      operator.String(),
		IssuerAddress: issuerAddress,
	}
}

func (msg *MsgRemoveIssuer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveIssuer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	return nil
}

func (msg *MsgRemoveIssuer) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
