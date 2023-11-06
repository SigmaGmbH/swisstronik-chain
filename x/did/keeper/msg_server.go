package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"swisstronik/x/did/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = &Keeper{}

const (
	DefaultAlternativeURITemplate    = "did:swtr:%s/resources/%s"
	DefaultAlternaticeURIDescription = "did-url"
)

// NewMsgServer returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &keeper
}

func (k Keeper) CreateDIDDocument(goCtx context.Context, msg *types.MsgCreateDIDDocument) (*types.MsgCreateDIDDocumentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get sign bytes before modifying payload
	signBytes := msg.Payload.GetSignBytes()

	// Normalize UUID identifiers
	msg.Normalize()

	// Validate DID doesn't exist
	if k.HasDIDDocument(ctx, msg.Payload.Id) {
		return nil, types.ErrDIDDocumentExists.Wrap(msg.Payload.Id)
	}

	// Build metadata and stateValue
	didDoc := msg.Payload.ToDidDoc()
	metadata := types.NewMetadataFromContext(ctx, msg.Payload.VersionId)
	didDocWithMetadata := types.NewDidDocWithMetadata(&didDoc, &metadata)

	// Consider DID that we are going to create during DID resolutions
	inMemoryDids := map[string]types.DIDDocumentWithMetadata{
		didDoc.Id: didDocWithMetadata,
	}

	// Check controllers' existence
	controllers := didDoc.AllControllerDIDs()
	for _, controller := range controllers {
		_, err := k.MustFindDIDDocument(ctx, inMemoryDids, controller)
		if err != nil {
			return nil, err
		}
	}

	// Verify signatures
	signers := GetSignerDIDsForDIDCreation(didDoc)
	err := k.VerifyAllSignersHaveAllValidSignatures(ctx, inMemoryDids, signBytes, signers, msg.Signatures)
	if err != nil {
		return nil, err
	}

	// Save first DID Document version
	err = k.AddNewDIDDocumentVersion(ctx, &didDocWithMetadata)
	if err != nil {
		return nil, types.ErrInternal.Wrapf(err.Error())
	}

	for _, vm := range didDoc.VerificationMethod {
		err = k.AddDIDControlledBy(ctx, vm.VerificationMaterial, didDoc.Id)
		if err != nil {
			return nil, types.ErrInternal.Wrapf(err.Error())
		}
	}

	// Build and return response
	return &types.MsgCreateDIDDocumentResponse{
		Value: &didDocWithMetadata,
	}, nil
}

func GetSignerDIDsForDIDCreation(did types.DIDDocument) []string {
	res := did.GetControllersOrSubject()
	res = append(res, did.GetVerificationMethodControllers()...)

	return types.UniqueSorted(res)
}

func (k Keeper) DeactivateDIDDocument(goCtx context.Context, msg *types.MsgDeactivateDIDDocument) (*types.MsgDeactivateDIDDocumentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get sign bytes before modifying payload
	signBytes := msg.Payload.GetSignBytes()

	// Normalize UUID identifiers
	msg.Normalize()

	// Validate DID does exist
	if !k.HasDIDDocument(ctx, msg.Payload.Id) {
		return nil, types.ErrDIDDocumentNotFound.Wrap(msg.Payload.Id)
	}

	// Retrieve DID Document state value and did
	didDoc, err := k.GetLatestDIDDocument(ctx, msg.Payload.Id)
	if err != nil {
		return nil, err
	}

	// Validate DID is not deactivated
	if didDoc.Metadata.Deactivated {
		return nil, types.ErrDIDDocumentDeactivated.Wrap(msg.Payload.Id)
	}

	// We neither create dids nor update
	inMemoryDids := map[string]types.DIDDocumentWithMetadata{}

	// Verify signatures
	signers := GetSignerDIDsForDIDCreation(*didDoc.DidDoc)
	err = k.VerifyAllSignersHaveAllValidSignatures(ctx, inMemoryDids, signBytes, signers, msg.Signatures)
	if err != nil {
		return nil, err
	}

	// Update metadata
	didDoc.Metadata.Deactivated = true
	didDoc.Metadata.Update(ctx, msg.Payload.VersionId)

	// Apply changes. We create a new version on deactivation to track deactivation time
	err = k.AddNewDIDDocumentVersion(ctx, &didDoc)
	if err != nil {
		return nil, types.ErrInternal.Wrapf(err.Error())
	}

	// Deactivate all previous versions
	var iterationErr error
	k.IterateDIDDocumentVersions(ctx, msg.Payload.Id, func(didDocWithMetadata types.DIDDocumentWithMetadata) bool {
		didDocWithMetadata.Metadata.Deactivated = true

		err := k.SetDIDDocumentVersion(ctx, &didDocWithMetadata, true)
		if err != nil {
			iterationErr = err
			return false
		}

		return true
	})

	if iterationErr != nil {
		return nil, types.ErrInternal.Wrapf(iterationErr.Error())
	}

	for _, vm := range didDoc.DidDoc.VerificationMethod {
		err = k.RemoveControlledDID(ctx, vm.VerificationMaterial, didDoc.DidDoc.Id)
		if err != nil {
			return nil, types.ErrInternal.Wrapf(err.Error())
		}
	}

	// Build and return response
	return &types.MsgDeactivateDIDDocumentResponse{
		Value: &didDoc,
	}, nil
}

func (k Keeper) UpdateDIDDocument(goCtx context.Context, msg *types.MsgUpdateDIDDocument) (*types.MsgUpdateDIDDocumentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get sign bytes before modifying payload
	signBytes := msg.Payload.GetSignBytes()

	// Normalize UUID identifiers
	msg.Normalize()

	// Check if DID exists and get latest version
	existingDidDocWithMetadata, err := k.GetLatestDIDDocument(ctx, msg.Payload.Id)
	if err != nil {
		return nil, types.ErrDIDDocumentNotFound.Wrap(err.Error())
	}

	existingDidDoc := existingDidDocWithMetadata.DidDoc

	// Validate DID is not deactivated
	if existingDidDocWithMetadata.Metadata.Deactivated {
		return nil, types.ErrDIDDocumentDeactivated.Wrap(msg.Payload.Id)
	}

	// Remove existing DID doc from associated DIDs with verification material
	for _, vm := range existingDidDoc.VerificationMethod {
		err = k.RemoveControlledDID(ctx, vm.VerificationMaterial, existingDidDoc.Id)
		if err != nil {
			return nil, types.ErrInternal.Wrapf(err.Error())
		}
	}

	// Construct the new version of the DID and temporary rename it and its self references
	// in order to consider old and new versions different DIDs during signatures validation
	updatedDidDoc := msg.Payload.ToDidDoc()
	updatedDidDoc.ReplaceDIDs(updatedDidDoc.Id, updatedDidDoc.Id+types.UpdatedPostfix)

	updatedMetadata := *existingDidDocWithMetadata.Metadata
	updatedMetadata.Update(ctx, msg.Payload.VersionId)

	updatedDidDocWithMetadata := types.NewDidDocWithMetadata(&updatedDidDoc, &updatedMetadata)

	// Consider the new version of the DID a separate DID
	inMemoryDids := map[string]types.DIDDocumentWithMetadata{updatedDidDoc.Id: updatedDidDocWithMetadata}

	// Check controllers existence
	controllers := updatedDidDoc.AllControllerDIDs()
	for _, controller := range controllers {
		_, err := k.MustFindDIDDocument(ctx, inMemoryDids, controller)
		if err != nil {
			return nil, err
		}
	}

	// Verify signatures
	// Duplicate signatures that reference the old version, make them reference a new (in memory) version
	// We can't use VerifySignatures because we can't uniquely identify a verification method corresponding to a given signInfo.
	// In other words if a signature belongs to the did being updated, there is no way to know which did version it belongs to: old or new.
	// To eliminate this problem we have to add pubkey to the signInfo in future.
	signers := GetSignerDIDsForDIDUpdate(*existingDidDoc, updatedDidDoc)
	extendedSignatures := DuplicateSignatures(msg.Signatures, existingDidDocWithMetadata.DidDoc.Id, updatedDidDoc.Id)
	err = k.VerifyAllSignersHaveAtLeastOneValidSignature(ctx, inMemoryDids, signBytes, signers, extendedSignatures, existingDidDoc.Id, updatedDidDoc.Id)
	if err != nil {
		return nil, err
	}

	// Return original id
	updatedDidDoc.ReplaceDIDs(updatedDidDoc.Id, existingDidDoc.Id)

	// Update state
	err = k.AddNewDIDDocumentVersion(ctx, &updatedDidDocWithMetadata)
	if err != nil {
		return nil, types.ErrInternal.Wrapf(err.Error())
	}

	for _, newVm := range updatedDidDoc.VerificationMethod {
		err = k.AddDIDControlledBy(ctx, newVm.VerificationMaterial, updatedDidDoc.Id)
		if err != nil {
			return nil, types.ErrInternal.Wrapf(err.Error())
		}
	}

	// Build and return response
	return &types.MsgUpdateDIDDocumentResponse{
		Value: &updatedDidDocWithMetadata,
	}, nil
}

func (k Keeper) CreateResource(goCtx context.Context, msg *types.MsgCreateResource) (*types.MsgCreateResourceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remember bytes before modifying payload
	signBytes := msg.Payload.GetSignBytes()

	msg.Normalize()

	// Validate corresponding DIDDoc exists
	did := types.JoinDID(types.DIDMethod, msg.Payload.CollectionId)
	didDoc, err := k.GetLatestDIDDocument(ctx, did)
	if err != nil {
		return nil, err
	}

	// Validate DID is not deactivated
	if didDoc.Metadata.Deactivated {
		return nil, types.ErrDIDDocumentDeactivated.Wrap(did)
	}

	// Validate Resource doesn't exist
	if k.HasResource(ctx, msg.Payload.CollectionId, msg.Payload.Id) {
		return nil, types.ErrResourceExists.Wrap(msg.Payload.Id)
	}

	// We can use the same signers as for DID creation because didDoc stays the same
	signers := GetSignerDIDsForDIDCreation(*didDoc.DidDoc)
	err = k.VerifyAllSignersHaveAllValidSignatures(ctx, map[string]types.DIDDocumentWithMetadata{}, signBytes, signers, msg.Signatures)
	if err != nil {
		return nil, err
	}

	// Build Resource
	resource := msg.Payload.ToResource()
	checksum := sha256.Sum256(resource.Resource.Data)
	resource.Metadata.Checksum = hex.EncodeToString(checksum[:])
	resource.Metadata.Created = ctx.BlockTime()
	resource.Metadata.MediaType = types.DetectMediaType(resource.Resource.Data)

	// Add default resource alternative url
	defaultAlternativeURL := types.AlternativeUri{
		Uri:         fmt.Sprintf(DefaultAlternativeURITemplate, msg.Payload.CollectionId, msg.Payload.Id),
		Description: DefaultAlternaticeURIDescription,
	}
	resource.Metadata.AlsoKnownAs = append(resource.Metadata.AlsoKnownAs, &defaultAlternativeURL)

	// Persist resource
	err = k.AddNewResourceVersion(ctx, &resource)
	if err != nil {
		return nil, types.ErrInternal.Wrapf(err.Error())
	}

	// Build and return response
	return &types.MsgCreateResourceResponse{
		Resource: resource.Metadata,
	}, nil
}

func DuplicateSignatures(signatures []*types.SignInfo, didToDuplicate string, newDid string) []*types.SignInfo {
	result := make([]*types.SignInfo, 0, len(signatures))

	for _, signature := range signatures {
		result = append(result, signature)

		did, path, query, fragment := types.MustSplitDIDUrl(signature.VerificationMethodId)
		if did == didToDuplicate {
			duplicate := types.SignInfo{
				VerificationMethodId: types.JoinDIDUrl(newDid, path, query, fragment),
				Signature:            signature.Signature,
			}

			result = append(result, &duplicate)
		}
	}

	return result
}

func GetSignerDIDsForDIDUpdate(existingDidDoc types.DIDDocument, updatedDidDoc types.DIDDocument) []string {
	signers := existingDidDoc.GetControllersOrSubject()
	signers = append(signers, updatedDidDoc.GetControllersOrSubject()...)

	existingVMMap := types.VerificationMethodListToMapByFragment(existingDidDoc.VerificationMethod)
	updatedVMMap := types.VerificationMethodListToMapByFragment(updatedDidDoc.VerificationMethod)

	for _, updatedVM := range updatedDidDoc.VerificationMethod {
		_, _, _, fragment := types.MustSplitDIDUrl(updatedVM.Id)
		existingVM, found := existingVMMap[fragment]

		// VM added
		if !found {
			signers = append(signers, updatedVM.Controller)
			continue
		}

		// Verification methods were updated
		// We have to revert renaming before comparing veriifcation methods.
		// Otherwise we will detect id and controller change
		// for non changed VMs because of `-updated` postfix.
		originalUpdatedVM := *updatedVM
		originalUpdatedVM.ReplaceDIDs(updatedDidDoc.Id, existingDidDoc.Id)

		if !reflect.DeepEqual(existingVM, originalUpdatedVM) {
			signers = append(signers, existingVM.Controller, updatedVM.Controller)
			continue
		}

		// verification methods not changed
	}

	for _, existingVM := range existingDidDoc.VerificationMethod {
		_, _, _, fragment := types.MustSplitDIDUrl(existingVM.Id)
		_, found := updatedVMMap[fragment]

		// VM removed
		if !found {
			signers = append(signers, existingVM.Controller)
			continue
		}
	}

	return types.UniqueSorted(signers)
}
