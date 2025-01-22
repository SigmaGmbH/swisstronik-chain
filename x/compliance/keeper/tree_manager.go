package keeper

import (
	"encoding/json"
	merkletree "github.com/SigmaGmbH/go-merkletree-sql/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/mimc7"
	"math/big"
	"swisstronik/x/compliance/types"
)

// GetIssuanceTreeRoot returns root of Sparse Merkle Tree with issued credentials
func (k Keeper) GetIssuanceTreeRoot(ctx sdk.Context) (*big.Int, error) {
	context := sdk.WrapSDKContext(ctx)
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	return tree.Root().BigInt(), nil
}

// GetRevocationTreeRoot returns root of Sparse Merkle Tree with revoked credentials
func (k Keeper) GetRevocationTreeRoot(ctx sdk.Context) (*big.Int, error) {
	context := sdk.WrapSDKContext(ctx)
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	return tree.Root().BigInt(), nil
}

func (k Keeper) AddCredentialHashToIssued(ctx sdk.Context, credentialHash *big.Int) error {
	context := sdk.WrapSDKContext(ctx)
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return err
	}

	key, err := mimc7.Hash([]*big.Int{credentialHash}, big.NewInt(0))
	if err != nil {
		return err
	}

	// Add zero element to revocation tree
	if err = k.addZeroElementToRevocationTree(ctx); err != nil {
		return err
	}

	return tree.Add(context, key, credentialHash)
}

func (k Keeper) MarkCredentialHashAsRevoked(ctx sdk.Context, credentialHash common.Hash) error {
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return err
	}

	value := credentialHash.Big()
	key, err := mimc7.Hash([]*big.Int{value}, big.NewInt(0))
	if err != nil {
		return err
	}

	return tree.Add(sdk.WrapSDKContext(ctx), key, value)
}

func (k Keeper) addZeroElementToRevocationTree(ctx sdk.Context) error {
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return err
	}

	// We add zero element to revocation tree to make it work correctly
	zero := big.NewInt(0)
	if _, _, _, err = tree.Get(sdk.WrapSDKContext(ctx), zero); err != nil {
		if err = tree.Add(sdk.WrapSDKContext(ctx), zero, zero); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) GetIssuanceProof(ctx sdk.Context, credentialHash common.Hash) ([]byte, error) {
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return nil, err
	}

	credentialHashBig := new(big.Int).SetBytes(credentialHash.Bytes())
	credentialKey, err := mimc7.Hash([]*big.Int{credentialHashBig}, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	proof, err := tree.GenerateCircomVerifierProof(sdk.WrapSDKContext(ctx), credentialKey, nil)
	if err != nil {
		return nil, err
	}

	return json.Marshal(proof)
}

func (k Keeper) GetNonRevocationProof(ctx sdk.Context, credentialHash common.Hash) ([]byte, error) {
	storage := NewTreeStorage(ctx, &k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return nil, err
	}

	credentialHashBig := new(big.Int).SetBytes(credentialHash.Bytes())
	credentialKey, err := mimc7.Hash([]*big.Int{credentialHashBig}, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	proof, err := tree.GenerateCircomVerifierProof(sdk.WrapSDKContext(ctx), credentialKey, nil)
	if err != nil {
		return nil, err
	}

	return json.Marshal(proof)
}

// SetTreeRoot is used only for testing
func (k Keeper) SetTreeRoot(context sdk.Context, treeKey []byte, root *merkletree.Hash) error {
	ctx := sdk.WrapSDKContext(context)
	storage := NewTreeStorage(context, &k, treeKey)
	return storage.SetRoot(ctx, root)
}

func (k Keeper) IsIncludedInIssuanceTree(context sdk.Context, credentialHash *big.Int) (bool, error) {
	ctx := sdk.WrapSDKContext(context)
	storage := NewTreeStorage(context, &k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(ctx, &storage, 32)
	if err != nil {
		return false, err
	}

	_, _, _, err = tree.Get(context, credentialHash)

	if err != nil {
		if err == merkletree.ErrKeyNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
