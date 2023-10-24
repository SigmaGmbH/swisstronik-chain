package keeper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"swisstronik/x/did/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDIDDocumentCount get the total number of did
func (k Keeper) GetDIDDocumentCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	key := types.StrBytes(types.DocumentCountKey)
	valueBytes := store.Get(key)

	// Count doesn't exist: no element
	if valueBytes == nil {
		return 0
	}

	// Parse bytes
	count, err := strconv.ParseUint(string(valueBytes), 10, 64)
	if err != nil {
		// Panic because the count should be always formattable to iint64
		panic("cannot decode count")
	}

	return count
}

// SetDIDDocumentCount set the total number of did
func (k Keeper) SetDIDDocumentCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)

	key := types.StrBytes(types.DocumentCountKey)
	valueBytes := []byte(strconv.FormatUint(count, 10))

	store.Set(key, valueBytes)
}

func (k Keeper) AddNewDIDDocumentVersion(ctx sdk.Context, didDoc *types.DIDDocumentWithMetadata) error {
	// Check if the diddoc version already exists
	if k.HasDIDDocumentVersion(ctx, didDoc.DidDoc.Id, didDoc.Metadata.VersionId) {
		return types.ErrDIDDocumentExists.Wrapf(
			"diddoc version already exists for did %s, version %s",
			didDoc.DidDoc.Id,
			didDoc.Metadata.VersionId,
		)
	}

	// Link to the previous version if it exists
	if k.HasDIDDocument(ctx, didDoc.DidDoc.Id) {
		latestVersionID, err := k.GetLatestDIDDocumentVersion(ctx, didDoc.DidDoc.Id)
		if err != nil {
			return err
		}

		latestVersion, err := k.GetDIDDocumentVersion(ctx, didDoc.DidDoc.Id, latestVersionID)
		if err != nil {
			return err
		}

		// Update version links
		latestVersion.Metadata.NextVersionId = didDoc.Metadata.VersionId
		didDoc.Metadata.PreviousVersionId = latestVersion.Metadata.VersionId

		// Update previous version with override
		err = k.SetDIDDocumentVersion(ctx, &latestVersion, true)
		if err != nil {
			return err
		}
	}

	// Update latest version
	err := k.SetLatestDIDDocumentVersion(ctx, didDoc.DidDoc.Id, didDoc.Metadata.VersionId)
	if err != nil {
		return err
	}

	// Write new version (no override)
	return k.SetDIDDocumentVersion(ctx, didDoc, false)
}

func (k Keeper) GetLatestDIDDocument(ctx sdk.Context, did string) (types.DIDDocumentWithMetadata, error) {
	latestVersionID, err := k.GetLatestDIDDocumentVersion(ctx, did)
	if err != nil {
		return types.DIDDocumentWithMetadata{}, err
	}

	latestVersion, err := k.GetDIDDocumentVersion(ctx, did, latestVersionID)
	if err != nil {
		return types.DIDDocumentWithMetadata{}, err
	}

	return latestVersion, nil
}

// SetDIDDocumentVersion set a specific version of DID Document in the store
func (k Keeper) SetDIDDocumentVersion(ctx sdk.Context, value *types.DIDDocumentWithMetadata, override bool) error {
	if !override && k.HasDIDDocumentVersion(ctx, value.DidDoc.Id, value.Metadata.VersionId) {
		return types.ErrDIDDocumentExists.Wrap("diddoc version already exists")
	}

	// Create the diddoc version
	store := ctx.KVStore(k.storeKey)

	key := types.GetDocumentVersionKey(value.DidDoc.Id, value.Metadata.VersionId)
	valueBytes := k.cdc.MustMarshal(value)
	store.Set(key, valueBytes)

	return nil
}

// GetDIDDocumentVersion returns a version of DID Document from its ID
func (k Keeper) GetDIDDocumentVersion(ctx sdk.Context, id, version string) (types.DIDDocumentWithMetadata, error) {
	store := ctx.KVStore(k.storeKey)

	if !k.HasDIDDocumentVersion(ctx, id, version) {
		return types.DIDDocumentWithMetadata{}, sdkerrors.ErrNotFound.Wrap("diddoc version not found")
	}

	var value types.DIDDocumentWithMetadata
	valueBytes := store.Get(types.GetDocumentVersionKey(id, version))
	k.cdc.MustUnmarshal(valueBytes, &value)

	return value, nil
}

func (k Keeper) GetAllDIDDocumentVersions(ctx sdk.Context, did string) ([]*types.Metadata, error) {
	store := ctx.KVStore(k.storeKey)

	result := make([]*types.Metadata, 0)

	versionIterator := sdk.KVStorePrefixIterator(store, types.GetDocumentVersionsPrefix(did))
	defer closeIteratorOrPanic(versionIterator)

	for ; versionIterator.Valid(); versionIterator.Next() {
		// Get the diddoc
		var didDoc types.DIDDocumentWithMetadata
		k.cdc.MustUnmarshal(versionIterator.Value(), &didDoc)

		result = append(result, didDoc.Metadata)
	}

	return result, nil
}

// SetLatestDIDDocumentVersion sets the latest version ID value for a DID document
func (k Keeper) SetLatestDIDDocumentVersion(ctx sdk.Context, did, version string) error {
	// Update counter. We use latest version as existence indicator.
	if !k.HasLatestDIDDocumentVersion(ctx, did) {
		count := k.GetDIDDocumentCount(ctx)
		k.SetDIDDocumentCount(ctx, count+1)
	}

	store := ctx.KVStore(k.storeKey)

	key := types.GetLatestDocumentVersionKey(did)
	valueBytes := types.StrBytes(version)
	store.Set(key, valueBytes)

	return nil
}

// GetLatestDIDDocumentVersion returns the latest version id value for a DID document
func (k Keeper) GetLatestDIDDocumentVersion(ctx sdk.Context, id string) (string, error) {
	store := ctx.KVStore(k.storeKey)

	if !k.HasLatestDIDDocumentVersion(ctx, id) {
		return "", sdkerrors.ErrNotFound.Wrap(id)
	}

	return string(store.Get(types.GetLatestDocumentVersionKey(id))), nil
}

func (k Keeper) HasDIDDocument(ctx sdk.Context, id string) bool {
	return k.HasLatestDIDDocumentVersion(ctx, id)
}

func (k Keeper) HasLatestDIDDocumentVersion(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetLatestDocumentVersionKey(id))
}

func (k Keeper) HasDIDDocumentVersion(ctx sdk.Context, id, version string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetDocumentVersionKey(id, version))
}

func (k Keeper) IterateDIDs(ctx sdk.Context, callback func(did string) (continue_ bool)) {
	store := ctx.KVStore(k.storeKey)
	latestVersionIterator := sdk.KVStorePrefixIterator(store, types.GetLatestDocumentVersionPrefix())
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		// Get did from key
		key := string(latestVersionIterator.Key())
		did := strings.Join(strings.Split(key, ":")[1:], ":")

		if !callback(did) {
			break
		}
	}
}

func (k Keeper) IterateDIDDocumentVersions(ctx sdk.Context, did string, callback func(version types.DIDDocumentWithMetadata) (continue_ bool)) {
	store := ctx.KVStore(k.storeKey)
	versionIterator := sdk.KVStorePrefixIterator(store, types.GetDocumentVersionsPrefix(did))
	defer closeIteratorOrPanic(versionIterator)

	for ; versionIterator.Valid(); versionIterator.Next() {
		var didDoc types.DIDDocumentWithMetadata
		k.cdc.MustUnmarshal(versionIterator.Value(), &didDoc)

		if !callback(didDoc) {
			break
		}
	}
}

func (k Keeper) IterateAllDIDDocumentVersions(ctx sdk.Context, callback func(version types.DIDDocumentWithMetadata) (continue_ bool)) {
	store := ctx.KVStore(k.storeKey)
	allVersionsIterator := sdk.KVStorePrefixIterator(store, []byte(types.DocumentVersionKey))
	defer closeIteratorOrPanic(allVersionsIterator)

	for ; allVersionsIterator.Valid(); allVersionsIterator.Next() {
		var didDoc types.DIDDocumentWithMetadata
		k.cdc.MustUnmarshal(allVersionsIterator.Value(), &didDoc)

		if !callback(didDoc) {
			break
		}
	}
}

// GetAllDIDDocuments returns all did
// Loads all DIDs in memory. Used only for genesis export.
func (k Keeper) GetAllDIDDocuments(ctx sdk.Context) ([]*types.DIDDocumentVersionSet, error) {
	var didDocs []*types.DIDDocumentVersionSet
	var err error

	k.IterateDIDs(ctx, func(did string) bool {
		var latestVersion string
		latestVersion, err = k.GetLatestDIDDocumentVersion(ctx, did)
		if err != nil {
			return false
		}

		didDocVersionSet := types.DIDDocumentVersionSet{
			LatestVersion: latestVersion,
		}

		k.IterateDIDDocumentVersions(ctx, did, func(version types.DIDDocumentWithMetadata) bool {
			didDocVersionSet.DidDocs = append(didDocVersionSet.DidDocs, &version)

			return true
		})

		didDocs = append(didDocs, &didDocVersionSet)

		return true
	})

	if err != nil {
		return nil, err
	}

	return didDocs, nil
}

func (k *Keeper) FindDIDDocument(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, did string) (res types.DIDDocumentWithMetadata, found bool, err error) {
	// Look in inMemory dict
	value, found := inMemoryDIDs[did]
	if found {
		return value, true, nil
	}

	// Look in state
	if k.HasDIDDocument(ctx, did) {
		value, err := k.GetLatestDIDDocument(ctx, did)
		if err != nil {
			return types.DIDDocumentWithMetadata{}, false, err
		}

		return value, true, nil
	}

	return types.DIDDocumentWithMetadata{}, false, nil
}

func (k *Keeper) MustFindDIDDocument(ctx sdk.Context, inMemoryDIDDocs map[string]types.DIDDocumentWithMetadata, did string) (res types.DIDDocumentWithMetadata, err error) {
	res, found, err := k.FindDIDDocument(ctx, inMemoryDIDDocs, did)
	if err != nil {
		return types.DIDDocumentWithMetadata{}, err
	}

	if !found {
		return types.DIDDocumentWithMetadata{}, types.ErrDIDDocumentNotFound.Wrap(did)
	}

	return res, nil
}

func (k *Keeper) FindVerificationMethod(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, didURL string) (res types.VerificationMethod, found bool, err error) {
	did, _, _, _ := types.MustSplitDIDUrl(didURL)

	didDoc, found, err := k.FindDIDDocument(ctx, inMemoryDIDs, did)
	if err != nil || !found {
		return types.VerificationMethod{}, found, err
	}

	for _, vm := range didDoc.DidDoc.VerificationMethod {
		if vm.Id == didURL {
			return *vm, true, nil
		}
	}

	return types.VerificationMethod{}, false, nil
}

func (k *Keeper) MustFindVerificationMethod(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, didURL string) (res types.VerificationMethod, err error) {
	res, found, err := k.FindVerificationMethod(ctx, inMemoryDIDs, didURL)
	if err != nil {
		return types.VerificationMethod{}, err
	}

	if !found {
		return types.VerificationMethod{}, types.ErrVerificationMethodNotFound.Wrap(didURL)
	}

	return res, nil
}

func (k *Keeper) VerifySignature(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, message []byte, signature types.SignInfo) error {
	verificationMethod, err := k.MustFindVerificationMethod(ctx, inMemoryDIDs, signature.VerificationMethodId)
	if err != nil {
		return err
	}

	err = types.VerifySignature(verificationMethod, message, signature.Signature)
	if err != nil {
		return types.ErrInvalidSignature.Wrapf("method id: %s", signature.VerificationMethodId)
	}

	return nil
}

func (k *Keeper) VerifyAllSignersHaveAllValidSignatures(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, message []byte, signers []string, signatures []*types.SignInfo) error {
	for _, signer := range signers {
		signatures := types.FindSignInfosBySigner(signatures, signer)

		if len(signatures) == 0 {
			return types.ErrSignatureNotFound.Wrapf("signer: %s", signer)
		}

		for _, signature := range signatures {
			err := k.VerifySignature(ctx, inMemoryDIDs, message, signature)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// VerifyAllSignersHaveAtLeastOneValidSignature verifies that all signers have at least one valid signature.
// Omit didToBeUpdated and updatedDID if not updating a DID. Otherwise those values will be used to better format error messages.
func (k *Keeper) VerifyAllSignersHaveAtLeastOneValidSignature(ctx sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata,
	message []byte, signers []string, signatures []*types.SignInfo, didToBeUpdated string, updatedDID string,
) error {
	for _, signer := range signers {
		signaturesBySigner := types.FindSignInfosBySigner(signatures, signer)
		signerForErrorMessage := GetSignerIDForErrorMessage(signer, didToBeUpdated, updatedDID)

		if len(signaturesBySigner) == 0 {
			return types.ErrSignatureNotFound.Wrapf("there should be at least one signature by %s", signerForErrorMessage)
		}

		found := false
		for _, signature := range signaturesBySigner {
			err := k.VerifySignature(ctx, inMemoryDIDs, message, signature)
			if err == nil {
				found = true
				break
			}
		}

		if !found {
			return types.ErrInvalidSignature.Wrapf("there should be at least one valid signature by %s", signerForErrorMessage)
		}
	}

	return nil
}

func GetSignerIDForErrorMessage(signerID string, existingVersionID string, updatedVersionID string) interface{} {
	if signerID == existingVersionID {
		return existingVersionID + " (previous version)"
	}

	if signerID == updatedVersionID {
		return existingVersionID + " (new version)"
	}

	return signerID
}

func closeIteratorOrPanic(iterator sdk.Iterator) {
	err := iterator.Close()
	if err != nil {
		panic(err.Error())
	}
}

// GetResourceCount get the total number of resource
func (k Keeper) GetResourceCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	byteKey := types.StrBytes(types.ResourceCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	count, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		// Panic because the count should be always formattable to int64
		panic("cannot decode count")
	}

	return count
}

// SetResourceCount set the total number of resource
func (k Keeper) SetResourceCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	byteKey := types.StrBytes(types.ResourceCountKey)

	// Set bytes
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

func (k Keeper) AddNewResourceVersion(ctx sdk.Context, resource *types.ResourceWithMetadata) error {
	// Find previous version and upgrade backward and forward version links
	previousResourceVersionHeader, found := k.GetLastResourceVersionMetadata(ctx, resource.Metadata.CollectionId, resource.Metadata.Name, resource.Metadata.ResourceType)
	if found {
		// Set links
		previousResourceVersionHeader.NextVersionId = resource.Metadata.Id
		resource.Metadata.PreviousVersionId = previousResourceVersionHeader.Id

		// Update previous version
		err := k.UpdateResourceMetadata(ctx, &previousResourceVersionHeader)
		if err != nil {
			return err
		}
	}

	// Set new version
	err := k.SetResource(ctx, resource)
	return err
}

// SetResource create or update a specific resource in the store
func (k Keeper) SetResource(ctx sdk.Context, resource *types.ResourceWithMetadata) error {
	if !k.HasResource(ctx, resource.Metadata.CollectionId, resource.Metadata.Id) {
		count := k.GetResourceCount(ctx)
		k.SetResourceCount(ctx, count+1)
	}

	store := ctx.KVStore(k.storeKey)

	// Set metadata
	metadataKey := types.GetResourceMetadataKey(resource.Metadata.CollectionId, resource.Metadata.Id)
	metadataBytes := k.cdc.MustMarshal(resource.Metadata)
	store.Set(metadataKey, metadataBytes)

	// Set data
	dataKey := types.GetResourceDataKey(resource.Metadata.CollectionId, resource.Metadata.Id)
	store.Set(dataKey, resource.Resource.Data)

	return nil
}

// GetResource returns a resource from its id
func (k Keeper) GetResource(ctx sdk.Context, collectionID string, id string) (types.ResourceWithMetadata, error) {
	if !k.HasResource(ctx, collectionID, id) {
		return types.ResourceWithMetadata{}, sdkerrors.ErrNotFound.Wrap("resource " + collectionID + ":" + id)
	}

	store := ctx.KVStore(k.storeKey)

	metadataBytes := store.Get(types.GetResourceMetadataKey(collectionID, id))
	var metadata types.ResourceMetadata
	if err := k.cdc.Unmarshal(metadataBytes, &metadata); err != nil {
		return types.ResourceWithMetadata{}, sdkerrors.ErrInvalidType.Wrap(err.Error())
	}

	dataBytes := store.Get(types.GetResourceDataKey(collectionID, id))
	data := types.Resource{Data: dataBytes}

	return types.ResourceWithMetadata{
		Metadata: &metadata,
		Resource: &data,
	}, nil
}

func (k Keeper) GetResourceMetadata(ctx sdk.Context, collectionID string, id string) (types.ResourceMetadata, error) {
	if !k.HasResource(ctx, collectionID, id) {
		return types.ResourceMetadata{}, sdkerrors.ErrNotFound.Wrap("resource " + collectionID + ":" + id)
	}

	store := ctx.KVStore(k.storeKey)

	metadataBytes := store.Get(types.GetResourceMetadataKey(collectionID, id))
	var metadata types.ResourceMetadata
	if err := k.cdc.Unmarshal(metadataBytes, &metadata); err != nil {
		return types.ResourceMetadata{}, sdkerrors.ErrInvalidType.Wrap(err.Error())
	}

	return metadata, nil
}

// HasResource checks if the resource exists in the store
func (k Keeper) HasResource(ctx sdk.Context, collectionID string, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetResourceMetadataKey(collectionID, id))
}

func (k Keeper) GetResourceCollection(ctx sdk.Context, collectionID string) []*types.ResourceMetadata {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetResourceMetadataCollectionPrefix(collectionID))

	var resources []*types.ResourceMetadata

	defer closeIteratorOrPanic(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.ResourceMetadata
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		resources = append(resources, &val)

	}

	return resources
}

func (k Keeper) GetLastResourceVersionMetadata(ctx sdk.Context, collectionID, name, resourceType string) (types.ResourceMetadata, bool) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GetResourceMetadataCollectionPrefix(collectionID))

	defer closeIteratorOrPanic(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var metadata types.ResourceMetadata
		k.cdc.MustUnmarshal(iterator.Value(), &metadata)

		if metadata.Name == name && metadata.ResourceType == resourceType && metadata.NextVersionId == "" {
			return metadata, true
		}
	}

	return types.ResourceMetadata{}, false
}

// UpdateResourceMetadata update the metadata of a resource. Returns an error if the resource doesn't exist
func (k Keeper) UpdateResourceMetadata(ctx sdk.Context, metadata *types.ResourceMetadata) error {
	if !k.HasResource(ctx, metadata.CollectionId, metadata.Id) {
		return sdkerrors.ErrNotFound.Wrap("resource " + metadata.CollectionId + ":" + metadata.Id)
	}

	store := ctx.KVStore(k.storeKey)

	// Set metadata
	metadataKey := types.GetResourceMetadataKey(metadata.CollectionId, metadata.Id)
	metadataBytes := k.cdc.MustMarshal(metadata)
	store.Set(metadataKey, metadataBytes)

	return nil
}

func (k Keeper) IterateAllResourceMetadatas(ctx sdk.Context, callback func(metadata types.ResourceMetadata) (continue_ bool)) {
	headerIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StrBytes(types.ResourceMetadataKey))
	defer closeIteratorOrPanic(headerIterator)

	for headerIterator.Valid() {
		var val types.ResourceMetadata
		k.cdc.MustUnmarshal(headerIterator.Value(), &val)

		if !callback(val) {
			break
		}

		headerIterator.Next()
	}
}

// GetAllResources returns all resources as a list
// Loads everything in memory. Use only for genesis export!
func (k Keeper) GetAllResources(ctx sdk.Context) (list []*types.ResourceWithMetadata, iterErr error) {
	k.IterateAllResourceMetadatas(ctx, func(metadata types.ResourceMetadata) bool {
		resource, err := k.GetResource(ctx, metadata.CollectionId, metadata.Id)
		if err != nil {
			iterErr = err
			return false
		}

		list = append(list, &resource)
		return true
	})

	return
}
