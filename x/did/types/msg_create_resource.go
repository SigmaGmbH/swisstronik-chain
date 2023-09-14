package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var _ sdk.Msg = &MsgCreateResource{}

func NewMsgCreateResource(payload *MsgCreateResourcePayload, signatures []*SignInfo) *MsgCreateResource {
	return &MsgCreateResource{
		Payload:    payload,
		Signatures: signatures,
	}
}

func (msg *MsgCreateResource) Route() string {
	return RouterKey
}

func (msg *MsgCreateResource) Type() string {
	return "MsgCreateResource"
}

func (msg *MsgCreateResource) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (msg *MsgCreateResource) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateResource) ValidateBasic() error {
	err := msg.Validate([]string{})
	if err != nil {
		return ErrBasicValidation.Wrap(err.Error())
	}

	return nil
}

func (msg MsgCreateResource) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.Payload, validation.Required, ValidMsgCreateResourcePayload()),
		validation.Field(&msg.Signatures, IsUniqueSignInfoListRule(), validation.Each(ValidSignInfoRule(allowedNamespaces))),
	)
}

func (msg *MsgCreateResource) Normalize() {
	msg.Payload.Normalize()
	NormalizeSignInfoList(msg.Signatures)
}

func (msg *MsgCreateResourcePayload) GetSignBytes() []byte {
	bytes, err := msg.Marshal()
	if err != nil {
		panic(err)
	}

	return bytes
}

func (msg *MsgCreateResourcePayload) ToResource() ResourceWithMetadata {
	return ResourceWithMetadata{
		Metadata: &ResourceMetadata{
			CollectionId: msg.CollectionId,
			Id:           msg.Id,
			Name:         msg.Name,
			Version:      msg.Version,
			ResourceType: msg.ResourceType,
			AlsoKnownAs:  msg.AlsoKnownAs,
		},
		Resource: &Resource{
			Data: msg.Data,
		},
	}
}

// Validation

func (msg MsgCreateResourcePayload) Validate() error {
	return validation.ValidateStruct(&msg,
		validation.Field(&msg.CollectionId, validation.Required, IsID()),
		validation.Field(&msg.Id, validation.Required, IsUUID()),
		validation.Field(&msg.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&msg.Version, validation.Length(1, 64)),
		validation.Field(&msg.ResourceType, validation.Required, validation.Length(1, 64)),
		validation.Field(&msg.AlsoKnownAs, validation.Each(ValidAlternativeURI())),
		validation.Field(&msg.Data, validation.Required, validation.Length(1, 200*1024)), // 200KB
	)
}

func ValidMsgCreateResourcePayload() *CustomErrorRule {
	return NewCustomErrorRule(func(value interface{}) error {
		casted, ok := value.(*MsgCreateResourcePayload)
		if !ok {
			panic("ValidMsgCreateResourcePayload must be only applied on MsgCreateDidPayload properties")
		}

		return casted.Validate()
	})
}

func (msg *MsgCreateResourcePayload) Normalize() {
	msg.CollectionId = NormalizeID(msg.CollectionId)
	msg.Id = NormalizeUUID(msg.Id)
}