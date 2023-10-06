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
func (k Keeper) GetDIDDocumentCount(ctx *sdk.Context) uint64 {
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
func (k Keeper) SetDIDDocumentCount(ctx *sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)

	key := types.StrBytes(types.DocumentCountKey)
	valueBytes := []byte(strconv.FormatUint(count, 10))

	store.Set(key, valueBytes)
}

func (k Keeper) AddNewDIDDocumentVersion(ctx *sdk.Context, didDoc *types.DIDDocumentWithMetadata) error {
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

func (k Keeper) GetLatestDIDDocument(ctx *sdk.Context, did string) (types.DIDDocumentWithMetadata, error) {
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
func (k Keeper) SetDIDDocumentVersion(ctx *sdk.Context, value *types.DIDDocumentWithMetadata, override bool) error {
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
func (k Keeper) GetDIDDocumentVersion(ctx *sdk.Context, id, version string) (types.DIDDocumentWithMetadata, error) {
	store := ctx.KVStore(k.storeKey)

	if !k.HasDIDDocumentVersion(ctx, id, version) {
		return types.DIDDocumentWithMetadata{}, sdkerrors.ErrNotFound.Wrap("diddoc version not found")
	}

	var value types.DIDDocumentWithMetadata
	valueBytes := store.Get(types.GetDocumentVersionKey(id, version))
	k.cdc.MustUnmarshal(valueBytes, &value)

	return value, nil
}

func (k Keeper) GetAllDIDDocumentVersions(ctx *sdk.Context, did string) ([]*types.Metadata, error) {
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
func (k Keeper) SetLatestDIDDocumentVersion(ctx *sdk.Context, did, version string) error {
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
func (k Keeper) GetLatestDIDDocumentVersion(ctx *sdk.Context, id string) (string, error) {
	store := ctx.KVStore(k.storeKey)

	if !k.HasLatestDIDDocumentVersion(ctx, id) {
		return "", sdkerrors.ErrNotFound.Wrap(id)
	}

	return string(store.Get(types.GetLatestDocumentVersionKey(id))), nil
}

func (k Keeper) HasDIDDocument(ctx *sdk.Context, id string) bool {
	return k.HasLatestDIDDocumentVersion(ctx, id)
}

func (k Keeper) HasLatestDIDDocumentVersion(ctx *sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetLatestDocumentVersionKey(id))
}

func (k Keeper) HasDIDDocumentVersion(ctx *sdk.Context, id, version string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetDocumentVersionKey(id, version))
}

func (k Keeper) IterateDIDs(ctx *sdk.Context, callback func(did string) (continue_ bool)) {
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

func (k Keeper) IterateDIDDocumentVersions(ctx *sdk.Context, did string, callback func(version types.DIDDocumentWithMetadata) (continue_ bool)) {
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

func (k Keeper) IterateAllDIDDocumentVersions(ctx *sdk.Context, callback func(version types.DIDDocumentWithMetadata) (continue_ bool)) {
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
func (k Keeper) GetAllDIDDocuments(ctx *sdk.Context) ([]*types.DIDDocumentVersionSet, error) {
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

func FindDIDDocument(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, did string) (res types.DIDDocumentWithMetadata, found bool, err error) {
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

func MustFindDIDDocument(k *Keeper, ctx *sdk.Context, inMemoryDIDDocs map[string]types.DIDDocumentWithMetadata, did string) (res types.DIDDocumentWithMetadata, err error) {
	res, found, err := FindDIDDocument(k, ctx, inMemoryDIDDocs, did)
	if err != nil {
		return types.DIDDocumentWithMetadata{}, err
	}

	if !found {
		return types.DIDDocumentWithMetadata{}, types.ErrDIDDocumentNotFound.Wrap(did)
	}

	return res, nil
}

func FindVerificationMethod(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, didURL string) (res types.VerificationMethod, found bool, err error) {
	did, _, _, _ := types.MustSplitDIDUrl(didURL)

	didDoc, found, err := FindDIDDocument(k, ctx, inMemoryDIDs, did)
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

func MustFindVerificationMethod(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, didURL string) (res types.VerificationMethod, err error) {
	res, found, err := FindVerificationMethod(k, ctx, inMemoryDIDs, didURL)
	if err != nil {
		return types.VerificationMethod{}, err
	}

	if !found {
		return types.VerificationMethod{}, types.ErrVerificationMethodNotFound.Wrap(didURL)
	}

	return res, nil
}

func VerifySignature(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, message []byte, signature types.SignInfo) error {
	verificationMethod, err := MustFindVerificationMethod(k, ctx, inMemoryDIDs, signature.VerificationMethodId)
	if err != nil {
		return err
	}

	err = types.VerifySignature(verificationMethod, message, signature.Signature)
	if err != nil {
		return types.ErrInvalidSignature.Wrapf("method id: %s", signature.VerificationMethodId)
	}

	return nil
}

func VerifyAllSignersHaveAllValidSignatures(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata, message []byte, signers []string, signatures []*types.SignInfo) error {
	for _, signer := range signers {
		signatures := types.FindSignInfosBySigner(signatures, signer)

		if len(signatures) == 0 {
			return types.ErrSignatureNotFound.Wrapf("signer: %s", signer)
		}

		for _, signature := range signatures {
			err := VerifySignature(k, ctx, inMemoryDIDs, message, signature)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// VerifyAllSignersHaveAtLeastOneValidSignature verifies that all signers have at least one valid signature.
// Omit didToBeUpdated and updatedDID if not updating a DID. Otherwise those values will be used to better format error messages.
func VerifyAllSignersHaveAtLeastOneValidSignature(k *Keeper, ctx *sdk.Context, inMemoryDIDs map[string]types.DIDDocumentWithMetadata,
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
			err := VerifySignature(k, ctx, inMemoryDIDs, message, signature)
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
