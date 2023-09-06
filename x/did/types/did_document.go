package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func NewDIDDocument(
	context []string, id string, controller []string, verificationMethod []*VerificationMethod,
	authentication []string, assertionMethod []string, capabilityInvocation []string, capabilityDelegation []string,
	keyAgreement []string, service []*Service, alsoKnownAs []string,
) *DIDDocument {
	return &DIDDocument{
		Context:              context,
		Id:                   id,
		Controller:           controller,
		VerificationMethod:   verificationMethod,
		Authentication:       authentication,
		AssertionMethod:      assertionMethod,
		CapabilityInvocation: capabilityInvocation,
		CapabilityDelegation: capabilityDelegation,
		KeyAgreement:         keyAgreement,
		Service:              service,
		AlsoKnownAs:          alsoKnownAs,
	}
}

// AllControllerDIDs returns controller DIDs used in both did.controllers and did.verification_method.controller
func (doc *DIDDocument) AllControllerDIDs() []string {
	result := doc.Controller
	result = append(result, doc.GetVerificationMethodControllers()...)

	return UniqueSorted(result)
}

// ReplaceDIDs replaces ids in all controller and id fields
func (doc *DIDDocument) ReplaceDIDs(old, new string) {
	// Controllers
	ReplaceInSlice(doc.Controller, old, new)

	// Id
	if doc.Id == old {
		doc.Id = new
	}

	// Verification methods
	for _, method := range doc.VerificationMethod {
		method.ReplaceDIDs(old, new)
	}

	// Services
	for _, service := range doc.Service {
		service.ReplaceDIDs(old, new)
	}

	// Verification relationships
	doc.Authentication = ReplaceDIDInDIDURLList(doc.Authentication, old, new)
	doc.AssertionMethod = ReplaceDIDInDIDURLList(doc.AssertionMethod, old, new)
	doc.CapabilityInvocation = ReplaceDIDInDIDURLList(doc.CapabilityInvocation, old, new)
	doc.CapabilityDelegation = ReplaceDIDInDIDURLList(doc.CapabilityDelegation, old, new)
	doc.KeyAgreement = ReplaceDIDInDIDURLList(doc.KeyAgreement, old, new)
}

func (doc *DIDDocument) GetControllersOrSubject() []string {
	result := doc.Controller

	if len(result) == 0 {
		result = append(result, doc.Id)
	}

	return result
}

func (doc *DIDDocument) GetVerificationMethodControllers() []string {
	result := make([]string, 0, len(doc.VerificationMethod))

	for _, vm := range doc.VerificationMethod {
		result = append(result, vm.Controller)
	}

	return result
}

func (doc DIDDocument) Validate(allowedNamespaces []string) error {
	return validation.ValidateStruct(&doc,
		validation.Field(&doc.Id, validation.Required, IsDID(allowedNamespaces)),
		validation.Field(&doc.Controller, IsUniqueStrList(), validation.Each(IsDID(allowedNamespaces))),
		validation.Field(&doc.VerificationMethod,
			IsUniqueVerificationMethodListByIDRule(), validation.Each(ValidVerificationMethodRule(doc.Id, allowedNamespaces)),
		),

		validation.Field(&doc.Authentication,
			IsUniqueStrList(), validation.Each(IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(doc.Id)),
		),
		validation.Field(&doc.AssertionMethod,
			IsUniqueStrList(), validation.Each(IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(doc.Id)),
		),
		validation.Field(&doc.CapabilityInvocation,
			IsUniqueStrList(), validation.Each(IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(doc.Id)),
		),
		validation.Field(&doc.CapabilityDelegation,
			IsUniqueStrList(), validation.Each(IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(doc.Id)),
		),
		validation.Field(&doc.KeyAgreement,
			IsUniqueStrList(), validation.Each(IsDIDUrl(allowedNamespaces, Empty, Empty, Required), HasPrefix(doc.Id)),
		),

		validation.Field(&doc.Service, IsUniqueServiceListByIDRule(), validation.Each(ValidServiceRule(doc.Id, allowedNamespaces))),
		validation.Field(&doc.AlsoKnownAs, IsUniqueStrList(), validation.Each(IsURI())),
	)
}

func NewMetadataFromContext(ctx sdk.Context, version string) Metadata {
	created := ctx.BlockTime()
	return Metadata{Created: created, Deactivated: false, VersionId: version}
}

func (m *Metadata) Update(ctx sdk.Context, version string) {
	updated := ctx.BlockTime()
	m.Updated = &updated
	m.VersionId = version
}

func NewDidDocWithMetadata(doc *DIDDocument, metadata *Metadata) DIDDocumentWithMetadata {
	return DIDDocumentWithMetadata{DidDoc: doc, Metadata: metadata}
}

func (d *DIDDocumentWithMetadata) ReplaceDids(prev, new string) {
	d.DidDoc.ReplaceDIDs(prev, new)
}