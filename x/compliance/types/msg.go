package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

func NewMsgSetAddressInfo(signer, userAddress, issuerAddress string) MsgSetAddressInfo {
	ethUserAddress := common.HexToAddress(userAddress)
	ethIssuerAddress := common.HexToAddress(issuerAddress)

	adapterData := &IssuerAdapterContractDetail{
		IssuerAlias:     issuerAddress,
		ContractAddress: ethIssuerAddress.Bytes(),
	}

	verificationEntry := &VerificationEntry{
		AdapterData:         adapterData,
		OriginChain:         "swisstronik",
		IssuanceTimestamp:   uint32(time.Now().Unix()),
		ExpirationTimestamp: 0,
		OriginalData:        nil,
	}

	verificationData := &VerificationData{
		VerificationType: VerificationType_VT_KYC,
		Entries:          []*VerificationEntry{verificationEntry},
	}

	addressInfo := &AddressInfo{
		Address:       ethUserAddress.Bytes(),
		IsVerified:    true,
		BanData:       nil,
		Verifications: []*VerificationData{verificationData},
	}

	return MsgSetAddressInfo{
		Signer:      signer,
		Data:        addressInfo,
		UserAddress: ethUserAddress.String(),
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
