package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var _ sdk.Msg = &MsgUpdateDIDDocument{}

func NewMsgUpdateDid(payload *MsgUpdateDIDDocumentPayload, signatures []*SignInfo) *MsgUpdateDIDDocument {
	return &MsgUpdateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}
}

func (msg *MsgUpdateDIDDocument) Route() string {
	return RouterKey
}

func (msg *MsgUpdateDIDDocument) Type() string {
	return "MsgUpdateDidDoc"
}

func (msg *MsgUpdateDIDDocument) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (msg *MsgUpdateDIDDocument) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateDIDDocument) ValidateBasic() error {
	err := msg.Validate(nil)
	if err != nil {
		return ErrBasicValidation.Wrap(err.Error())
	}

	return nil
}

func (msg MsgUpdateDIDDocument) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Payload, validation.Required, ValidMsgUpdateDidPayloadRule(allowedNamespaces)),
		validation.Field(&msg.Signatures, IsUniqueSignInfoListRule(), validation.Each(ValidSignInfoRule(allowedNamespaces))),
	)
}

func (msg *MsgUpdateDIDDocument) Normalize() {
	msg.Payload.Normalize()
	NormalizeSignInfoList(msg.Signatures)
}

func (msg *MsgUpdateDIDDocumentPayload) GetSignBytes() []byte {
	bytes, err := msg.Marshal()
	if err != nil {
		panic(err)
	}

	return bytes
}

func (msg *MsgUpdateDIDDocumentPayload) ToDidDoc() DIDDocument {
	return DIDDocument{
		Context:              msg.Context,
		Id:                   msg.Id,
		Controller:           msg.Controller,
		VerificationMethod:   msg.VerificationMethod,
		Authentication:       msg.Authentication,
		AssertionMethod:      msg.AssertionMethod,
		CapabilityInvocation: msg.CapabilityInvocation,
		CapabilityDelegation: msg.CapabilityDelegation,
		KeyAgreement:         msg.KeyAgreement,
		AlsoKnownAs:          msg.AlsoKnownAs,
		Service:              msg.Service,
	}
}

func (msg MsgUpdateDIDDocumentPayload) Validate(allowedNamespaces []string) error {
	err := msg.ToDidDoc().Validate(allowedNamespaces)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(&msg,
		validation.Field(&msg.VersionId, validation.Required),
	)
}

func ValidMsgUpdateDidPayloadRule(allowedNamespaces []string) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(*MsgUpdateDIDDocumentPayload)
		if !ok {
			panic("ValidMsgUpdateDidPayloadRule must be only applied on MsgUpdateDidPayload properties")
		}

		return casted.Validate(allowedNamespaces)
	})
}

func (msg *MsgUpdateDIDDocumentPayload) Normalize() {
	msg.Id = NormalizeDID(msg.Id)
	for _, vm := range msg.VerificationMethod {
		vm.Controller = NormalizeDID(vm.Controller)
		vm.Id = NormalizeDIDUrl(vm.Id)
	}
	for _, s := range msg.Service {
		s.Id = NormalizeDIDUrl(s.Id)
	}
	msg.Controller = NormalizeDIDList(msg.Controller)
	msg.Authentication = NormalizeDIDUrlList(msg.Authentication)
	msg.AssertionMethod = NormalizeDIDUrlList(msg.AssertionMethod)
	msg.CapabilityInvocation = NormalizeDIDUrlList(msg.CapabilityInvocation)
	msg.CapabilityDelegation = NormalizeDIDUrlList(msg.CapabilityDelegation)
	msg.KeyAgreement = NormalizeDIDUrlList(msg.KeyAgreement)

	msg.VersionId = NormalizeUUID(msg.VersionId)
}