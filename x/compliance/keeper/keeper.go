package keeper

import (
	"fmt"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/ethereum/go-ethereum/common"
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

func (k *Keeper) AddVerificationEntry(ctx sdk.Context, subjectAddress, issuerAddress common.Address, originChain string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	adapterData := types.IssuerAdapterContractDetail{
		IssuerAlias:     issuerAddress.String(),
		ContractAddress: issuerAddress.Bytes(),
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
	verificationDataBytes, err := verificationData.Marshal()
	if err != nil {
		return err
	}

	store.Set(subjectAddress.Bytes(), verificationDataBytes)
	return nil
}

func (k *Keeper) GetVerificationData(ctx sdk.Context, subjectAddress common.Address) (*types.VerificationData, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerification)

	verificationDataBytes := store.Get(subjectAddress.Bytes())
	if verificationDataBytes == nil {
		return &types.VerificationData{}, nil
	}

	var verificationData types.VerificationData
	if err := proto.Unmarshal(verificationDataBytes, &verificationData); err != nil {
		return nil, err
	}

	return &verificationData, nil
}
