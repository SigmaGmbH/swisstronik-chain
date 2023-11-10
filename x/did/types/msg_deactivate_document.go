package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var _ sdk.Msg = &MsgDeactivateDIDDocument{}

func NewMsgDeactivateDid(payload *MsgDeactivateDIDDocumentPayload, signatures []*SignInfo) *MsgDeactivateDIDDocument {
	return &MsgDeactivateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}
}

func (msg *MsgDeactivateDIDDocument) Route() string {
	return RouterKey
}

func (msg *MsgDeactivateDIDDocument) Type() string {
	return "MsgDeactivateDidDoc"
}

func (msg *MsgDeactivateDIDDocument) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (msg *MsgDeactivateDIDDocument) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshal(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeactivateDIDDocument) ValidateBasic() error {
	err := msg.Validate(nil)
	if err != nil {
		return ErrBasicValidation.Wrap(err.Error())
	}

	return nil
}

func (msg MsgDeactivateDIDDocument) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Payload, validation.Required, ValidMsgDeactivateDidPayloadRule(allowedNamespaces)),
		validation.Field(&msg.Signatures, IsUniqueSignInfoListRule(), validation.Each(ValidSignInfoRule(allowedNamespaces))),
	)
}

func (msg *MsgDeactivateDIDDocument) Normalize() {
	msg.Payload.Normalize()
}

func (msg *MsgDeactivateDIDDocumentPayload) GetSignBytes() []byte {
	bytes, err := msg.Marshal()
	if err != nil {
		panic(err)
	}

	return bytes
}

func (msg MsgDeactivateDIDDocumentPayload) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Id, validation.Required, IsDID()),
		validation.Field(&msg.VersionId, validation.Required, IsUUID()),
	)
}

func ValidMsgDeactivateDidPayloadRule(allowedNamespaces []string) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(*MsgDeactivateDIDDocumentPayload)
		if !ok {
			panic("ValidMsgDeactivateDidPayloadRule must be only applied on MsgDeactivateDidPayload properties")
		}

		return casted.Validate(allowedNamespaces)
	})
}

func (msg *MsgDeactivateDIDDocumentPayload) Normalize() {
	msg.Id = NormalizeDID(msg.Id)
	msg.VersionId = NormalizeUUID(msg.VersionId)
}