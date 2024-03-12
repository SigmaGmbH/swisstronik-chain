package keeper

import (
	"cosmossdk.io/errors"
	"fmt"
	"github.com/cosmos/gogoproto/proto"

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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerAlias)

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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerAlias)

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
