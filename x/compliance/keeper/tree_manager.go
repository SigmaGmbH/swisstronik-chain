package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-merkletree-sql"
	"math/big"
	"swisstronik/x/compliance/types"
)

// GetIssuanceTreeRoot returns root of Sparse Merkle Tree with issued credentials
func (k Keeper) GetIssuanceTreeRoot(ctx sdk.Context) (*big.Int, error) {
	context := sdk.WrapSDKContext(ctx)
	storage := NewTreeStorage(&k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	return tree.Root().BigInt(), nil
}

// GetRevocationTreeRoot returns root of Sparse Merkle Tree with revoked credentials
func (k Keeper) GetRevocationTreeRoot(ctx sdk.Context) (*big.Int, error) {
	context := sdk.WrapSDKContext(ctx)
	storage := NewTreeStorage(&k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	return tree.Root().BigInt(), nil
}

func (k Keeper) AddCredentialHashToIssued(context sdk.Context, credentialHash common.Hash) error {
	// TODO: Implement
	return nil
}

func (k Keeper) MarkCredentialHashAsRevoked(context sdk.Context, credentialHash common.Hash) error {
	// TODO: Implement
	return nil
}

func (k Keeper) GetIssuanceProof(context sdk.Context, credentialHash common.Hash) ([]byte, error) {
	storage := NewTreeStorage(&k, types.KeyPrefixIssuanceTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	credentialHashBig := new(big.Int).SetBytes(credentialHash.Bytes())
	proof, _, err := tree.GenerateProof(sdk.WrapSDKContext(context), credentialHashBig, nil)
	if err != nil {
		return nil, err
	}

	return proof.MarshalJSON()
}

func (k Keeper) GetNonRevocationProof(context sdk.Context, credentialHash common.Hash) ([]byte, error) {
	storage := NewTreeStorage(&k, types.KeyPrefixRevocationTree)
	tree, err := merkletree.NewMerkleTree(context, &storage, 32)
	if err != nil {
		return nil, err
	}

	credentialHashBig := new(big.Int).SetBytes(credentialHash.Bytes())
	proof, _, err := tree.GenerateProof(sdk.WrapSDKContext(context), credentialHashBig, nil)
	if err != nil {
		return nil, err
	}

	return proof.MarshalJSON()
}
