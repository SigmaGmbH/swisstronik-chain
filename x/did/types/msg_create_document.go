package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var _ sdk.Msg = &MsgCreateDIDDocument{}

func NewMsgCreateDid(payload *MsgCreateDIDDocumentPayload, signatures []*SignInfo) *MsgCreateDIDDocument {
	return &MsgCreateDIDDocument{
		Payload:    payload,
		Signatures: signatures,
	}
}

func (msg *MsgCreateDIDDocument) Route() string {
	return RouterKey
}

func (msg *MsgCreateDIDDocument) Type() string {
	return "MsgCreateDidDoc"
}

func (msg *MsgCreateDIDDocument) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (msg *MsgCreateDIDDocument) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDIDDocument) ValidateBasic() error {
	err := msg.Validate(nil)
	if err != nil {
		return ErrBasicValidation.Wrap(err.Error())
	}

	return nil
}

func (msg MsgCreateDIDDocument) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Payload, validation.Required, ValidMsgCreateDidPayloadRule(allowedNamespaces)),
		validation.Field(&msg.Signatures, IsUniqueSignInfoListByIDRule(), validation.Each(ValidSignInfoRule(allowedNamespaces))),
	)
}

func (msg *MsgCreateDIDDocument) Normalize() {
	msg.Payload.Normalize()
	NormalizeSignInfoList(msg.Signatures)
}

func (msg *MsgCreateDIDDocumentPayload) GetSignBytes() []byte {
	bytes, err := msg.Marshal()
	if err != nil {
		panic(err)
	}

	return bytes
}

func (msg *MsgCreateDIDDocumentPayload) ToDidDoc() DIDDocument {
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

func (msg MsgCreateDIDDocumentPayload) Validate(allowedNamespaces []string) error {
	err := msg.ToDidDoc().Validate(allowedNamespaces)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(&msg,
		validation.Field(&msg.VersionId, validation.Required),
	)
}

func ValidMsgCreateDidPayloadRule(allowedNamespaces []string) *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(*MsgCreateDIDDocumentPayload)
		if !ok {
			panic("ValidMsgCreateDidPayloadRule must be only applied on MsgCreateDidPayload properties")
		}

		return casted.Validate(allowedNamespaces)
	})
}

func (msg *MsgCreateDIDDocumentPayload) Normalize() {
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