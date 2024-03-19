package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewSetIssuerDetailsMsg(operator, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgSetIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
		Operator:    operator,
	}
	return MsgSetIssuerDetails{
		Signer:        operator,
		IssuerAddress: issuerAddress,
		Details:       &issuerDetails,
	}
}

func (msg *MsgSetIssuerDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetIssuerDetails) ValidateBasic() error {
	signerAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}

	operatorAddr, err := sdk.AccAddressFromBech32(msg.Details.Operator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	if !operatorAddr.Equals(signerAddr) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "operator and signer address mismatch (%s)", err)
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

func NewUpdateIssuerDetailsMsg(existingOperator, newOperator, issuerAddress, issuerName, issuerDescription, issuerURL, issuerLogo, issuerLegalEntity string) MsgUpdateIssuerDetails {
	issuerDetails := IssuerDetails{
		Name:        issuerName,
		Description: issuerDescription,
		Url:         issuerURL,
		Logo:        issuerLogo,
		LegalEntity: issuerLegalEntity,
		Operator:    newOperator,
	}
	return MsgUpdateIssuerDetails{
		Signer:        existingOperator,
		IssuerAddress: issuerAddress,
		Details:       &issuerDetails,
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

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
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
		Signer:        operator,
		IssuerAddress: issuerAddress,
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

	_, err = sdk.AccAddressFromBech32(msg.IssuerAddress)
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
