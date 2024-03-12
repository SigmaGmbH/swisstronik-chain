package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgSetAddressInfo(
	signer, userAddress, issuerAddress, issuerAlias, originChain string,
	creationTimestamp uint32,
	verificationType VerificationType,
) MsgSetAddressInfo {
	adapterData := &IssuerAdapterContractDetail{
		IssuerAlias:     issuerAlias,
		ContractAddress: issuerAddress,
	}

	verificationEntry := &VerificationEntry{
		AdapterData:         adapterData,
		OriginChain:         originChain,
		IssuanceTimestamp:   creationTimestamp,
		ExpirationTimestamp: 0,
		OriginalData:        nil,
	}

	verificationData := &VerificationData{
		VerificationType: verificationType,
		Entries:          []*VerificationEntry{verificationEntry},
	}

	addressInfo := &AddressInfo{
		Address:       userAddress,
		IsVerified:    true,
		BanData:       nil,
		Verifications: []*VerificationData{verificationData},
	}

	return MsgSetAddressInfo{
		Signer:      signer,
		Data:        addressInfo,
		UserAddress: userAddress,
	}
}

func (msg *MsgSetAddressInfo) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetAddressInfo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetAddressInfo) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}
	return nil
}
