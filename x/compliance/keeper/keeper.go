package keeper

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"slices"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/crypto"

	"swisstronik/x/compliance/types"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetIssuerDetails sets details for provided issuer address
func (k Keeper) SetIssuerDetails(ctx sdk.Context, issuerAddress sdk.AccAddress, details *types.IssuerDetails) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)

	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	store.Set(issuerAddress.Bytes(), detailsBytes)

	return nil
}

// RemoveIssuer removes provided issuer
func (k Keeper) RemoveIssuer(ctx sdk.Context, issuerAddress sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)
	store.Delete(issuerAddress.Bytes())
	// NOTE, all the verification data verified by removed issuer must be deleted from store
	// But for now, let's keep those verifications.
	// They will be filtered out at the time when call `GetAddressDetails` or `GetVerificationDetails`

	// Remove address details for issuer
	k.RemoveAddressDetails(ctx, issuerAddress)
}

// GetIssuerDetails returns details of provided issuer address
func (k Keeper) GetIssuerDetails(ctx sdk.Context, issuerAddress sdk.AccAddress) (*types.IssuerDetails, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)

	detailsBytes := store.Get(issuerAddress.Bytes())
	if detailsBytes == nil {
		return &types.IssuerDetails{}, nil
	}

	var issuerDetails types.IssuerDetails
	if err := proto.Unmarshal(detailsBytes, &issuerDetails); err != nil {
		return nil, err
	}

	return &issuerDetails, nil
}

// IssuerExists checks if issuer exists by checking operator address
func (k Keeper) IssuerExists(ctx sdk.Context, issuerAddress sdk.AccAddress) (bool, error) {
	res, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return false, err
	}
	return len(res.Name) > 0, nil
}

// GetAddressDetails returns actual address details (without non-existent issuers)
func (k Keeper) GetAddressDetails(ctx sdk.Context, address sdk.AccAddress) (*types.AddressDetails, error) {
	addressDetails, err := k.GetFullAddressDetails(ctx, address)
	if err != nil {
		return nil, err
	}

	// Filter verification details by issuer's existance
	var newVerifications []*types.Verification
	for _, verification := range addressDetails.Verifications {
		issuerAddress, err := sdk.AccAddressFromBech32(verification.IssuerAddress)
		if err != nil {
			return nil, err
		}
		exists, err := k.IssuerExists(ctx, issuerAddress)
		if err != nil {
			return nil, err
		}
		if exists {
			newVerifications = append(newVerifications, verification)
		}
	}
	addressDetails.Verifications = newVerifications

	return addressDetails, nil
}

// GetFullAddressDetails returns address details with all verifications
func (k Keeper) GetFullAddressDetails(ctx sdk.Context, address sdk.AccAddress) (*types.AddressDetails, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)

	addressDetailsBytes := store.Get(address.Bytes())
	if addressDetailsBytes == nil {
		return &types.AddressDetails{}, nil
	}

	var addressDetails types.AddressDetails
	if err := proto.Unmarshal(addressDetailsBytes, &addressDetails); err != nil {
		return nil, err
	}

	return &addressDetails, nil
}

// SetAddressDetails writes address details to the storage
func (k Keeper) SetAddressDetails(ctx sdk.Context, address sdk.AccAddress, details *types.AddressDetails) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)
	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}
	store.Set(address.Bytes(), detailsBytes)
	return nil
}

// RemoveAddressDetails deletes address details from store
func (k Keeper) RemoveAddressDetails(ctx sdk.Context, address sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)
	store.Delete(address.Bytes())
}

// IsAddressVerified returns information if address is verified.
func (k Keeper) IsAddressVerified(ctx sdk.Context, address sdk.AccAddress) (bool, error) {
	addressDetails, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return false, err
	}

	// If address is banned, its verification is suspended
	return addressDetails.IsVerified, nil
}

// SetAddressVerificationStatus marks provided address as verified or not verified.
func (k Keeper) SetAddressVerificationStatus(ctx sdk.Context, address sdk.AccAddress, isVerifiedStatus bool) error {
	addressDetails, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return err
	}

	// Skip if address already has provided status
	if addressDetails.IsVerified == isVerifiedStatus {
		return nil
	}

	addressDetails.IsVerified = isVerifiedStatus
	if err := k.SetAddressDetails(ctx, address, addressDetails); err != nil {
		return err
	}

	return nil
}

// AddVerificationDetailsV2 writes details of passed verification by provided address. It writes credential to ZK-SDI
// even if user has no attached public key
func (k Keeper) AddVerificationDetailsV2(ctx sdk.Context, userAddress sdk.AccAddress, verificationType types.VerificationType, details *types.VerificationDetails, userPublicKeyCompressed []byte) ([]byte, error) {
	// Check if issuer is verified and not banned
	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return nil, err
	}

	verificationDetailsID, err := k.addVerificationDetailsInternal(ctx, userAddress, issuerAddress, verificationType, details)
	if err != nil {
		return nil, err
	}

	xPublicKey, err := types.ExtractXCoordinate(userPublicKeyCompressed, false)
	if err != nil {
		return nil, errors.Wrap(types.ErrBadRequest, err.Error())
	}

	if err = k.LinkVerificationIdToPubKey(ctx, xPublicKey.Bytes(), verificationDetailsID); err != nil {
		return nil, errors.Wrap(types.ErrBadRequest, err.Error())
	}
	credentialValue := &types.ZKCredential{
		Type:                verificationType,
		IssuerAddress:       issuerAddress.Bytes(),
		HolderPublicKey:     xPublicKey.Bytes(),
		ExpirationTimestamp: details.ExpirationTimestamp,
		IssuanceTimestamp:   details.IssuanceTimestamp,
	}
	credentialHash, err := credentialValue.Hash()
	if err != nil {
		return nil, errors.Wrap(types.ErrBadRequest, err.Error())
	}

	err = k.AddCredentialHashToIssued(ctx, credentialHash)
	if err != nil {
		return nil, errors.Wrap(types.ErrBadRequest, err.Error())
	}

	return verificationDetailsID, nil
}

// AddVerificationDetails writes details of passed verification by provided address.
// It writes to ZK-SDI only if user has attached public key
func (k Keeper) AddVerificationDetails(ctx sdk.Context, userAddress sdk.AccAddress, verificationType types.VerificationType, details *types.VerificationDetails) ([]byte, error) {
	// Check if issuer is verified and not banned
	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return nil, err
	}

	verificationDetailsID, err := k.addVerificationDetailsInternal(ctx, userAddress, issuerAddress, verificationType, details)
	if err != nil {
		return nil, err
	}

	// If user has attached public key, add to ZK-SDI
	userPublicKey := k.GetHolderPublicKey(ctx, userAddress)
	if userPublicKey != nil {
		if err = k.LinkVerificationIdToPubKey(ctx, userPublicKey, verificationDetailsID); err != nil {
			return nil, err
		}

		credentialValue := &types.ZKCredential{
			Type:                verificationType,
			IssuerAddress:       issuerAddress.Bytes(),
			HolderPublicKey:     userPublicKey,
			ExpirationTimestamp: details.ExpirationTimestamp,
			IssuanceTimestamp:   details.IssuanceTimestamp,
		}
		credentialHash, err := credentialValue.Hash()
		if err != nil {
			return nil, errors.Wrap(types.ErrBadRequest, err.Error())
		}
		err = k.AddCredentialHashToIssued(ctx, credentialHash)
		if err != nil {
			return nil, errors.Wrap(types.ErrBadRequest, err.Error())
		}
	}

	return verificationDetailsID, nil
}

func (k Keeper) addVerificationDetailsInternal(ctx sdk.Context, userAddress sdk.AccAddress, issuerAddress sdk.AccAddress, verificationType types.VerificationType, details *types.VerificationDetails) ([]byte, error) {
	if err := details.ValidateSize(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParam, err.Error())
	}

	isAddressVerified, err := k.IsAddressVerified(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}

	if !isAddressVerified {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer not verified")
	}

	if verificationType <= types.VerificationType_VT_UNSPECIFIED || verificationType > types.VerificationType_VT_BIOMETRIC {
		return nil, errors.Wrap(types.ErrInvalidParam, "invalid verification type")
	}
	details.Type = verificationType
	if details.IssuanceTimestamp < 1 || (details.ExpirationTimestamp > 0 && details.IssuanceTimestamp >= details.ExpirationTimestamp) {
		return nil, errors.Wrap(types.ErrInvalidParam, "invalid issuance timestamp. Should be less than expiration timestamp.")
	}
	if len(details.OriginalData) < 1 {
		return nil, errors.Wrap(types.ErrInvalidParam, "empty proof data")
	}

	detailsBytes, err := details.Marshal()
	if err != nil {
		return nil, err
	}

	// Check if there is no such verification details in storage yet
	verificationDetailsID := crypto.Keccak256(userAddress.Bytes(), verificationType.ToBytes(), detailsBytes)
	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)

	if verificationDetailsStore.Has(verificationDetailsID) {
		return nil, errors.Wrapf(
			types.ErrInvalidParam,
			"provided verification details already in storage. Verification ID: (%s)",
			hexutil.Encode(verificationDetailsID),
		)
	}

	// If there is no such verification details associated with provided address, write them to the table
	verificationDetailsStore.Set(verificationDetailsID, detailsBytes)

	// Associate provided verification details with user address
	verification := &types.Verification{
		Type:           verificationType,
		VerificationId: verificationDetailsID,
		IssuerAddress:  issuerAddress.String(),
	}
	userAddressDetails, err := k.GetAddressDetails(ctx, userAddress)
	if err != nil {
		return nil, err
	}

	if slices.Contains(userAddressDetails.Verifications, verification) {
		return nil, errors.Wrap(types.ErrInvalidParam, "such verification already associated with user address")
	}

	userAddressDetails.Verifications = append(userAddressDetails.Verifications, verification)
	if err := k.SetAddressDetails(ctx, userAddress, userAddressDetails); err != nil {
		return nil, err
	}

	if err = k.LinkVerificationToHolder(ctx, userAddress, verificationDetailsID); err != nil {
		return nil, err
	}

	return verificationDetailsID, nil
}

// SetVerificationDetails writes verification details. Since this function writes directly to the storage,
// it should be used only in genesis.go or in tests
func (k Keeper) SetVerificationDetails(
	ctx sdk.Context,
	userAddress sdk.AccAddress,
	verificationDetailsId []byte,
	details *types.VerificationDetails,
) error {
	if err := details.ValidateSize(); err != nil {
		return errors.Wrap(types.ErrInvalidParam, err.Error())
	}

	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)
	if verificationDetailsStore.Has(verificationDetailsId) {
		return errors.Wrap(types.ErrInvalidParam, "provided verification details already in storage")
	}

	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	// If there is no such verification details associated with provided address, write them to the table
	verificationDetailsStore.Set(verificationDetailsId, detailsBytes)

	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return err
	}

	if err = k.LinkVerificationToHolder(ctx, userAddress, verificationDetailsId); err != nil {
		return err
	}

	// If user has linked public key or self-attached public key, add to ZK-SDI
	var userPublicKey []byte
	userPublicKey = k.GetPubKeyByVerificationId(ctx, verificationDetailsId)
	if userPublicKey == nil {
		userPublicKey = k.GetHolderPublicKey(ctx, userAddress)
	}

	// If there is no public key, skip adding to issuance tree
	if userPublicKey != nil {
		if err = k.LinkVerificationIdToPubKey(ctx, userPublicKey, verificationDetailsId); err != nil {
			return err
		}

		credentialValue := &types.ZKCredential{
			Type:                details.Type,
			IssuerAddress:       issuerAddress.Bytes(),
			HolderPublicKey:     userPublicKey,
			ExpirationTimestamp: details.ExpirationTimestamp,
			IssuanceTimestamp:   details.IssuanceTimestamp,
		}
		credentialHash, err := credentialValue.Hash()
		if err != nil {
			return err
		}
		err = k.AddCredentialHashToIssued(ctx, credentialHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) RevokeVerification(ctx sdk.Context, verificationDetailsId []byte, issuerAddress sdk.AccAddress) error {
	verificationDetails, err := k.GetVerificationDetails(ctx, verificationDetailsId)
	if err != nil {
		return err
	}

	if verificationDetails.IsRevoked {
		return errors.Wrap(types.ErrInvalidParam, "verification was already revoked")
	}

	if verificationDetails.IssuerAddress != issuerAddress.String() {
		return errors.Wrap(types.ErrInvalidParam, "caller is not verification issuer")
	}

	return k.MarkVerificationDetailsAsRevoked(ctx, verificationDetailsId)
}

func (k Keeper) MarkVerificationDetailsAsRevoked(
	ctx sdk.Context,
	verificationDetailsId []byte,
) error {
	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)
	if !verificationDetailsStore.Has(verificationDetailsId) {
		return errors.Wrap(types.ErrInvalidParam, "there is no such verification with provided ID")
	}

	prevVerificationDetailsBytes := verificationDetailsStore.Get(verificationDetailsId)
	if prevVerificationDetailsBytes == nil {
		return errors.Wrap(types.ErrInvalidParam, "verification with provided ID is empty")
	}

	prevVerificationDetails := &types.VerificationDetails{}
	err := proto.Unmarshal(prevVerificationDetailsBytes, prevVerificationDetails)
	if err != nil {
		return err
	}

	prevVerificationDetails.IsRevoked = true

	detailsBytes, err := prevVerificationDetails.Marshal()
	if err != nil {
		return err
	}

	// If there is no such verification details associated with provided address, write them to the table
	verificationDetailsStore.Set(verificationDetailsId, detailsBytes)

	issuerAddress, err := sdk.AccAddressFromBech32(prevVerificationDetails.IssuerAddress)
	if err != nil {
		return err
	}

	userAddress := k.getHolderByVerificationId(ctx, verificationDetailsId)
	if userAddress.Empty() {
		return errors.Wrap(
			types.ErrBadRequest,
			"cannot find associated user address. Please create a ticket if you see this error",
		)
	}

	var attachedPublicKey []byte
	attachedPublicKey = k.GetPubKeyByVerificationId(ctx, verificationDetailsId)
	if attachedPublicKey == nil {
		attachedPublicKey = k.GetHolderPublicKey(ctx, userAddress)
	}
	if attachedPublicKey != nil {
		// Update revocation tree with provided credential
		credential := &types.ZKCredential{
			Type:                prevVerificationDetails.Type,
			IssuerAddress:       issuerAddress.Bytes(),
			HolderPublicKey:     attachedPublicKey,
			ExpirationTimestamp: prevVerificationDetails.ExpirationTimestamp,
			IssuanceTimestamp:   prevVerificationDetails.IssuanceTimestamp,
		}
		credentialHash, err := credential.Hash()
		if err != nil {
			return err
		}

		return k.MarkCredentialHashAsRevoked(ctx, common.BigToHash(credentialHash))
	}

	return nil
}

// GetVerificationDetails returns verification details for provided ID
func (k Keeper) GetVerificationDetails(ctx sdk.Context, verificationId []byte) (*types.VerificationDetails, error) {
	verificationDetails, err := k.GetRawVerificationDetails(ctx, verificationId)
	if err != nil {
		return nil, err
	}

	// Check if verification details exist, return empty struct if not
	if verificationDetails.Type == types.VerificationType_VT_UNSPECIFIED {
		return &types.VerificationDetails{}, nil
	}

	// Check if issuer exists
	issuerAddress, err := sdk.AccAddressFromBech32(verificationDetails.IssuerAddress)
	if err != nil {
		return nil, err
	}
	exists, err := k.IssuerExists(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &types.VerificationDetails{}, nil
	}

	return verificationDetails, nil
}

// GetRawVerificationDetails returns verification details for provided ID
func (k Keeper) GetRawVerificationDetails(ctx sdk.Context, verificationId []byte) (*types.VerificationDetails, error) {
	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)
	verificationDetailsBytes := verificationDetailsStore.Get(verificationId)
	if verificationDetailsBytes == nil {
		return &types.VerificationDetails{}, nil
	}

	var verificationDetails types.VerificationDetails
	if err := proto.Unmarshal(verificationDetailsBytes, &verificationDetails); err != nil {
		return nil, err
	}

	return &verificationDetails, nil
}

func (k Keeper) GetVerificationDetailsByIssuer(ctx sdk.Context, userAddress sdk.AccAddress, issuerAddress sdk.AccAddress) ([]*types.Verification, []*types.VerificationDetails, error) {
	addressDetails, err := k.GetAddressDetails(ctx, userAddress)
	if err != nil {
		return nil, nil, err
	}

	var (
		filteredVerifications       []*types.Verification
		filteredVerificationDetails []*types.VerificationDetails
	)
	for _, verification := range addressDetails.Verifications {
		if verification.IssuerAddress != issuerAddress.String() {
			continue
		}
		verificationDetails, err := k.GetVerificationDetails(ctx, verification.VerificationId)
		if err != nil {
			return nil, nil, err
		}
		filteredVerifications = append(filteredVerifications, verification)
		filteredVerificationDetails = append(filteredVerificationDetails, verificationDetails)
	}
	return filteredVerifications, filteredVerificationDetails, nil
}

func (k Keeper) GetCredentialHashByVerificationId(ctx sdk.Context, verificationId []byte) ([]byte, error) {
	details, err := k.GetVerificationDetails(ctx, verificationId)
	if err != nil {
		return nil, err
	}

	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return nil, err
	}

	holder := k.getHolderByVerificationId(ctx, verificationId)
	var userPublicKey []byte
	userPublicKey = k.GetPubKeyByVerificationId(ctx, verificationId)
	if userPublicKey == nil {
		userPublicKey = k.GetHolderPublicKey(ctx, holder)
	}

	if userPublicKey == nil {
		return nil, errors.Wrap(types.ErrInvalidParam, "verification with provided ID has no public key to attach")
	}

	credentialValue := &types.ZKCredential{
		Type:                details.Type,
		IssuerAddress:       issuerAddress.Bytes(),
		HolderPublicKey:     userPublicKey,
		ExpirationTimestamp: details.ExpirationTimestamp,
		IssuanceTimestamp:   details.IssuanceTimestamp,
	}
	credentialHash, err := credentialValue.Hash()
	if err != nil {
		return nil, err
	}

	return credentialHash.Bytes(), nil
}

// HasVerificationOfType checks if user has verifications of specific type (for example, passed KYC) from provided issuers.
// If there is no provided expected issuers, this function will check if user has any verification of appropriate type.
func (k Keeper) HasVerificationOfType(ctx sdk.Context, userAddress sdk.AccAddress, expectedType types.VerificationType, expirationTimestamp uint32, expectedIssuers []sdk.AccAddress) (bool, error) {
	// Obtain user address details
	userAddressDetails, err := k.GetAddressDetails(ctx, userAddress)
	if err != nil {
		return false, err
	}

	// If expiration is 0, it means infinite period
	if expirationTimestamp < 1 {
		expirationTimestamp = ^uint32(0)
	}

	for _, verification := range userAddressDetails.Verifications {
		if verification.Type == expectedType {
			// If not found matched issuer, do not get details to check expiration
			found := false
			for _, expectedIssuer := range expectedIssuers {
				if verification.IssuerAddress == expectedIssuer.String() {
					found = true
					break
				}
			}
			if len(expectedIssuers) > 0 && !found {
				continue
			}

			verificationDetails, err := k.GetVerificationDetails(ctx, verification.VerificationId)
			if err != nil {
				continue
			}
			// Check if verification is valid by given expiration timestamp
			if verificationDetails.ExpirationTimestamp > 0 && expirationTimestamp > verificationDetails.ExpirationTimestamp {
				continue
			}
			return true, nil
		}
	}

	return false, nil
}

func (k Keeper) GetVerificationsOfType(ctx sdk.Context, userAddress sdk.AccAddress, expectedType types.VerificationType, expectedIssuers ...sdk.AccAddress) ([]*types.VerificationDetails, error) {
	// Obtain user address details
	userAddressDetails, err := k.GetAddressDetails(ctx, userAddress)
	if err != nil {
		return nil, err
	}

	// Filter verifications with expected type
	var appropriateTypeVerifications []*types.Verification
	for _, verification := range userAddressDetails.Verifications {
		if verification.Type == expectedType {
			appropriateTypeVerifications = append(appropriateTypeVerifications, verification)
		}
	}

	if len(appropriateTypeVerifications) == 0 {
		return nil, nil
	}

	// Extract verification data
	var verifications []*types.VerificationDetails
	for _, verification := range appropriateTypeVerifications {
		// Filter verifications by expected issuer
		if expectedIssuers != nil && slices.ContainsFunc(expectedIssuers, func(expectedIssuer sdk.AccAddress) bool {
			if expectedIssuer.String() == verification.IssuerAddress {
				return true
			}
			return false
		}) == false {
			continue
		}

		verificationDetails, err := k.GetVerificationDetails(ctx, verification.VerificationId)
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, verificationDetails)
	}

	return verifications, nil
}

// GetOperatorDetails returns the operator details
func (k Keeper) GetOperatorDetails(ctx sdk.Context, operator sdk.AccAddress) (*types.OperatorDetails, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOperatorDetails)

	detailsBytes := store.Get(operator.Bytes())
	if detailsBytes == nil {
		return &types.OperatorDetails{}, nil
	}

	var operatorDetails types.OperatorDetails
	if err := proto.Unmarshal(detailsBytes, &operatorDetails); err != nil {
		return nil, err
	}

	return &operatorDetails, nil
}

// AddOperator adds initial/regular operator.
// Initial operator can not be removed
func (k Keeper) AddOperator(ctx sdk.Context, operator sdk.AccAddress, operatorType types.OperatorType) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOperatorDetails)

	if operatorType <= types.OperatorType_OT_UNSPECIFIED || operatorType > types.OperatorType_OT_REGULAR {
		return errors.Wrap(types.ErrInvalidParam, "invalid operator type")
	}

	details := &types.OperatorDetails{
		Operator:     operator.String(),
		OperatorType: operatorType,
	}
	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	store.Set(operator.Bytes(), detailsBytes)
	return nil
}

// RemoveRegularOperator removes regular operator
func (k Keeper) RemoveRegularOperator(ctx sdk.Context, operator sdk.AccAddress) error {
	operatorDetails, err := k.GetOperatorDetails(ctx, operator)
	if err != nil || operatorDetails == nil {
		return errors.Wrapf(types.ErrInvalidOperator, "operator not exists")
	}

	if operatorDetails.OperatorType != types.OperatorType_OT_REGULAR {
		return errors.Wrapf(types.ErrNotAuthorized, "operator not a regular type")
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOperatorDetails)
	store.Delete(operator.Bytes())
	return nil
}

// OperatorExists checks if operator exists
func (k Keeper) OperatorExists(ctx sdk.Context, operator sdk.AccAddress) (bool, error) {
	res, err := k.GetOperatorDetails(ctx, operator)
	if err != nil || res == nil {
		return false, err
	}
	return len(res.Operator) > 0, nil
}

// GetHolderPublicKey returns the compressed holder public key
func (k Keeper) GetHolderPublicKey(ctx sdk.Context, user sdk.AccAddress) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixHolderPublicKeys)

	publicKeyBytes := store.Get(user.Bytes())
	if publicKeyBytes == nil {
		return nil
	}

	return publicKeyBytes
}

// SetHolderPublicKey returns the compressed holder public key
func (k Keeper) SetHolderPublicKey(ctx sdk.Context, user sdk.AccAddress, publicKey []byte) error {
	// Check if there is no public key
	currentPublicKeyBytes := k.GetHolderPublicKey(ctx, user)
	if currentPublicKeyBytes != nil {
		return errors.Wrap(types.ErrInvalidParam, "public key already set")
	}

	xCoordPublicKey, err := types.ExtractXCoordinate(publicKey, false)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidParam, "cannot parse provided public key: (%s)", err)
	}

	k.SetHolderPublicKeyBytes(ctx, user, xCoordPublicKey.Bytes())

	return nil
}

func (k Keeper) SetHolderPublicKeyBytes(ctx sdk.Context, user sdk.AccAddress, xCoordinateBytes []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixHolderPublicKeys)
	store.Set(user.Bytes(), xCoordinateBytes)
}

func (k Keeper) LinkVerificationToHolder(ctx sdk.Context, userAddress sdk.AccAddress, verificationId []byte) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToHolder)

	// Check if there is already linked user
	linkedUserBytes := store.Get(verificationId)
	if linkedUserBytes != nil {
		if bytes.Equal(linkedUserBytes, userAddress.Bytes()) {
			// This user is already linked
			return nil
		}

		return errors.Wrap(types.ErrBadRequest, "provided verification id is already linked to another user")
	}

	store.Set(verificationId, userAddress.Bytes())
	return nil
}

func (k Keeper) getHolderByVerificationId(ctx sdk.Context, verificationId []byte) sdk.AccAddress {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToHolder)
	return store.Get(verificationId)
}

func (k Keeper) LinkVerificationIdToPubKey(ctx sdk.Context, publicKey []byte, verificationId []byte) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToPubKey)

	// Check if there is already linked user
	currentPublicKey := store.Get(verificationId)
	if currentPublicKey != nil {
		if bytes.Equal(currentPublicKey, publicKey) {
			// This user is already linked
			return nil
		}

		return errors.Wrap(types.ErrBadRequest, "provided verification id is already linked to another public key")
	}

	store.Set(verificationId, publicKey)
	return nil
}

func (k Keeper) GetPubKeyByVerificationId(ctx sdk.Context, verificationId []byte) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToPubKey)
	return store.Get(verificationId)
}

// IsVerificationRevoked checks if verification with provided verification id is revoked or its issuer
// was not verified or was removed.
func (k Keeper) IsVerificationRevoked(ctx sdk.Context, verificationId []byte) (bool, error) {
	verificationDetails, err := k.GetVerificationDetails(ctx, verificationId)
	if err != nil {
		return false, err
	}

	if verificationDetails.IsRevoked {
		return true, nil
	}

	issuerAddress, err := sdk.AccAddressFromBech32(verificationDetails.IssuerAddress)
	addressDetails, err := k.GetAddressDetails(ctx, issuerAddress)
	if err != nil {
		return false, err
	}

	if addressDetails.IsRevoked || !addressDetails.IsVerified {
		return true, nil
	}

	return false, nil
}

func (k Keeper) ConvertCredential(ctx sdk.Context, verificationId []byte, publicKeyToSet []byte, caller sdk.AccAddress) error {
	// Check if signer is owner of credential
	credentialOwner := k.getHolderByVerificationId(ctx, verificationId)
	if !credentialOwner.Equals(caller) {
		return errors.Wrap(types.ErrBadRequest, "signer is not credential holder")
	}

	var holderPublicKey []byte
	holderPublicKey = k.GetHolderPublicKey(ctx, caller)
	if holderPublicKey == nil {
		// validate provided public key
		xCoordPublicKey, err := types.ExtractXCoordinate(publicKeyToSet, false)
		if err != nil {
			return errors.Wrapf(types.ErrInvalidParam, "cannot parse provided public key: (%s)", err)
		}
		holderPublicKey = xCoordPublicKey.Bytes()
	}

	err := k.LinkVerificationIdToPubKey(ctx, holderPublicKey, verificationId)
	if err != nil {
		return err
	}

	isVerificationRevoked, err := k.IsVerificationRevoked(ctx, verificationId)
	if err != nil {
		return err
	}
	if isVerificationRevoked {
		return errors.Wrap(types.ErrBadRequest, "credential was revoked")
	}

	details, err := k.GetVerificationDetails(ctx, verificationId)
	if err != nil {
		return err
	}

	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return err
	}

	credentialValue := &types.ZKCredential{
		Type:                details.Type,
		IssuerAddress:       issuerAddress.Bytes(),
		HolderPublicKey:     holderPublicKey,
		ExpirationTimestamp: details.ExpirationTimestamp,
		IssuanceTimestamp:   details.IssuanceTimestamp,
	}
	credentialHash, err := credentialValue.Hash()
	if err != nil {
		return err
	}

	isIncluded, err := k.IsIncludedInIssuanceTree(ctx, credentialHash)
	if err != nil {
		return err
	}

	if isIncluded {
		return errors.Wrap(types.ErrBadRequest, "credential is already included in issuance tree")
	}

	return k.AddCredentialHashToIssued(ctx, credentialHash)
}

// TODO: Create fn to obtain all verified issuers with their aliases
