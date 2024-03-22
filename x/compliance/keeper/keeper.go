package keeper

import (
	"cosmossdk.io/errors"
	"fmt"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/crypto"
	"slices"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

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

func (k Keeper) SetAddressDetailsRaw(ctx sdk.Context, subjectAddress sdk.Address, data *types.AddressDetails) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	dataBytes, err := data.Marshal()
	if err != nil {
		return err
	}

	store.Set(subjectAddress.Bytes(), dataBytes)
	return nil
}

// SetIssuerDetails sets details for provided issuer address
func (k Keeper) SetIssuerDetails(ctx sdk.Context, issuerAddress sdk.Address, details *types.IssuerDetails) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)

	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	store.Set(issuerAddress.Bytes(), detailsBytes)
	return nil
}

// RemoveIssuer removes provided issuer
func (k Keeper) RemoveIssuer(ctx sdk.Context, issuerAddress sdk.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)
	store.Delete(issuerAddress.Bytes())
}

// GetIssuerDetails returns details of provided issuer address
func (k Keeper) GetIssuerDetails(ctx sdk.Context, issuerAddress sdk.Address) (*types.IssuerDetails, error) {
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

// GetAddressDetails returns address details
func (k Keeper) GetAddressDetails(ctx sdk.Context, address sdk.Address) (*types.AddressDetails, error) {
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
func (k Keeper) SetAddressDetails(ctx sdk.Context, address sdk.Address, details *types.AddressDetails) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)
	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}
	store.Set(address.Bytes(), detailsBytes)
	return nil
}

// IsAddressVerified returns information if address is verified.
func (k Keeper) IsAddressVerified(ctx sdk.Context, address sdk.Address) (bool, error) {
	addressDetails, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return false, err
	}

	// If address is banned, its verification is suspended
	return addressDetails.IsVerified, nil
}

// SetAddressVerificationStatus marks provided address as verified or not verified.
func (k Keeper) SetAddressVerificationStatus(ctx sdk.Context, address sdk.Address, isVerifiedStatus bool) error {
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

// AddVerificationDetails writes details of passed verification by provided address.
func (k Keeper) AddVerificationDetails(ctx sdk.Context, userAddress sdk.Address, verificationType types.VerificationType, details *types.VerificationDetails) error {
	// Check if issuer is verified and not banned
	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return err
	}

	// TODO: Uncomment once verification mechanism is done
	//isAddressVerified, err := k.IsAddressVerified(ctx, issuerAddress)
	//if err != nil {
	//	return err
	//}
	//
	//if !isAddressVerified {
	//	return errors.Wrap(types.ErrInvalidParam, "issuer is not verified")
	//}

	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	// Check if there is no such verification details in storage yet
	verificationDetailsID := crypto.Keccak256(userAddress.Bytes(), verificationType.ToBytes(), detailsBytes)
	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)

	if verificationDetailsStore.Has(verificationDetailsID) {
		return errors.Wrap(types.ErrInvalidParam, "provided verification details already in storage")
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
		return err
	}

	if slices.Contains(userAddressDetails.Verifications, verification) {
		return errors.Wrap(types.ErrInvalidParam, "such verification already associated with user address")
	}

	userAddressDetails.Verifications = append(userAddressDetails.Verifications, verification)
	if err := k.SetAddressDetails(ctx, userAddress, userAddressDetails); err != nil {
		return err
	}

	return nil
}

// GetVerificationDetails returns verification details for provided ID
func (k Keeper) GetVerificationDetails(ctx sdk.Context, verificationDetailsId []byte) (*types.VerificationDetails, error) {
	verificationDetailsStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)
	verificationDetailsBytes := verificationDetailsStore.Get(verificationDetailsId)
	if verificationDetailsBytes == nil {
		return nil, nil
	}

	var verificationDetails types.VerificationDetails
	if err := proto.Unmarshal(verificationDetailsBytes, &verificationDetails); err != nil {
		return nil, err
	}

	return &verificationDetails, nil
}

// HasVerificationOfType checks if user has verifications of specific type (for example, passed KYC) from provided issuers.
// If there is no provided expected issuers, this function will check if user has any verification of appropriate type.
func (k Keeper) HasVerificationOfType(ctx sdk.Context, userAddress sdk.Address, expectedType types.VerificationType, expectedIssuers []sdk.Address) (bool, error) {
	// Obtain user address details
	userAddressDetails, err := k.GetAddressDetails(ctx, userAddress)
	if err != nil {
		return false, err
	}

	// Filter verifications with expected type
	var appropriateTypeVerifications []*types.Verification
	for _, verification := range userAddressDetails.Verifications {
		if verification.Type == expectedType {
			appropriateTypeVerifications = append(appropriateTypeVerifications, verification)
		}
	}

	// If there is no provided issuers, check if there are any appropriate verification
	if len(expectedIssuers) == 0 && len(appropriateTypeVerifications) != 0 {
		return true, nil
	}

	// Filter verifications with expected issuers
	for _, verification := range appropriateTypeVerifications {
		for _, expectedIssuer := range expectedIssuers {
			if verification.IssuerAddress == expectedIssuer.String() {
				return true, nil
			}
		}
	}
	return false, nil
}

func (k Keeper) GetVerificationsOfType(ctx sdk.Context, userAddress sdk.Address, expectedType types.VerificationType, expectedIssuers ...sdk.Address) ([]*types.VerificationDetails, error) {
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
		verificationDetails, err := k.GetVerificationDetails(ctx, verification.VerificationId)
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, verificationDetails)
	}

	return verifications, nil
}

// IssuerExists checks if issuer exists by checking operator address
func (k Keeper) IssuerExists(ctx sdk.Context, issuerAddress sdk.Address) (bool, error) {
	res, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return false, err
	}

	exists := len(res.Operator) != 0
	return exists, nil
}

// TODO: Create fn to obtain all verified issuers with their aliases
