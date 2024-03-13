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

func (k Keeper) SetAddressInfoRaw(ctx sdk.Context, subjectAddress sdk.Address, data *types.AddressInfo) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	dataBytes, err := data.Marshal()
	if err != nil {
		return err
	}

	store.Set(subjectAddress.Bytes(), dataBytes)
	return nil
}

// TODO: methods for ban / unban
func (k Keeper) AddVerificationEntry(ctx sdk.Context, subjectAddress, issuerAddress sdk.Address, originChain string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	// TODO: Check if entry for address already exists
	// If entry exists, add verificationData to existing entry
	// Otherwise, create a new one
	adapterData := types.IssuerAdapterContractDetail{
		IssuerAlias:     issuerAddress.String(),
		ContractAddress: issuerAddress.String(),
	}

	verificationEntry := &types.VerificationEntry{
		AdapterData:         &adapterData,
		OriginChain:         originChain,
		IssuanceTimestamp:   uint32(ctx.BlockHeader().Time.Unix()),
		ExpirationTimestamp: 0,
		OriginalData:        nil,
	}

	verificationData := types.VerificationData{
		VerificationType: types.VerificationType_VT_KYC,
		Entries:          []*types.VerificationEntry{verificationEntry},
	}

	addrInfo := types.AddressInfo{
		Address:       subjectAddress.String(),
		IsVerified:    true,
		BanData:       nil,
		Verifications: []*types.VerificationData{&verificationData},
	}

	addrInfoBytes, err := addrInfo.Marshal()
	if err != nil {
		return err
	}

	store.Set(subjectAddress.Bytes(), addrInfoBytes)
	return nil
}

// GetAddressInfo returns `AddressInfo` associated with provided address.
func (k Keeper) GetAddressInfo(ctx sdk.Context, address sdk.Address) (*types.AddressInfo, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	addrInfoBytes := store.Get(address.Bytes())
	if addrInfoBytes == nil {
		return &types.AddressInfo{}, nil
	}

	var addrInfo types.AddressInfo
	if err := proto.Unmarshal(addrInfoBytes, &addrInfo); err != nil {
		return nil, err
	}

	return &addrInfo, nil
}

// SetIssuerDetailsInner sets description about provided issuer address
func (k Keeper) SetIssuerDetailsInner(ctx sdk.Context, issuerAddress sdk.Address, alias string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)

	if len(alias) == 0 {
		return errors.Wrap(types.ErrInvalidParam, "invalid issuer alias")
	}

	details := &types.IssuerDetails{IssuerAlias: alias}
	detailsBytes, err := details.Marshal()
	if err != nil {
		return err
	}

	store.Set(issuerAddress.Bytes(), detailsBytes)
	return nil
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

// GetIssuerAlias returns human-readable alias of provided issuer address
func (k Keeper) GetIssuerAlias(ctx sdk.Context, issuerAddress sdk.Address) (string, error) {
	issuerDetails, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return "", err
	}

	return issuerDetails.IssuerAlias, nil
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

// IsAddressVerified returns information if address is verified and not banned.
// If address is banned, this function will return `false` to prevent issuer from writing new verification data
func (k Keeper) IsAddressVerified(ctx sdk.Context, address sdk.Address) (bool, error) {
	addressDetails, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return false, err
	}

	// If address is banned, its verification is suspended
	return addressDetails.IsVerified && !addressDetails.IsBanned, nil
}

// MarkAddressAsVerified marks provided address as verified. This function should be called
// as a result of accepted governance proposal.
func (k Keeper) MarkAddressAsVerified(ctx sdk.Context, address sdk.Address) error {
	// TODO: Add call to `x/evm` to check if this address is contract address
	addressDetails, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return err
	}

	// If address is banned, return error
	if addressDetails.IsBanned {
		return errors.Wrap(types.ErrInvalidParam, "address is banned")
	}

	// Skip if address is already verified
	if addressDetails.IsVerified {
		return nil
	}

	addressDetails.IsVerified = true
	if err := k.SetAddressDetails(ctx, address, addressDetails); err != nil {
		return err
	}

	return nil
}

// AddVerificationDetails writes details of passed verification by provided address.
func (k Keeper) AddVerificationDetails(ctx sdk.Context, userAddress sdk.Address, verificationType types.VerificationType, details types.VerificationDetails) error {
	// Check if issuer is verified and not banned
	issuerAddress, err := sdk.AccAddressFromBech32(details.Issuer)
	if err != nil {
		return err
	}

	isAddressVerified, err := k.IsAddressVerified(ctx, issuerAddress)
	if err != nil {
		return err
	}

	if !isAddressVerified {
		return errors.Wrap(types.ErrInvalidParam, "issuer is not verified")
	}

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
