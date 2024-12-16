package keeper

import (
	"bytes"
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iden3/go-merkletree-sql"
)

type TreeStorage struct {
	keeper         *Keeper
	keyPrefix      []byte
	internalPrefix []byte
	currentRoot    *merkletree.Hash
}

func NewTreeStorage(k *Keeper, treeKeyPrefix []byte) TreeStorage {
	return TreeStorage{
		keeper:         k,
		keyPrefix:      treeKeyPrefix,
		internalPrefix: []byte{},
		currentRoot:    nil,
	}
}

// WithPrefix implements the method WithPrefix of the interface db.Storage
func (ts *TreeStorage) WithPrefix(prefix []byte) merkletree.Storage {
	return &TreeStorage{
		ts.keeper,
		ts.keyPrefix,
		merkletree.Concat(ts.internalPrefix, prefix),
		nil,
	}
}

// Get retrieves a value from a key in the db.Storage
func (ts *TreeStorage) Get(ctx context.Context, key []byte) (*merkletree.Node, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)

	res := store.Get(merkletree.Concat(ts.internalPrefix, key))
	if res == nil {
		return nil, merkletree.ErrNotFound
	}

	return merkletree.NewNodeFromBytes(res)
}

// Put inserts new node into Sparse Merkle Tree
func (ts *TreeStorage) Put(ctx context.Context, key []byte, node *merkletree.Node) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	value := node.Value()

	store.Set(merkletree.Concat(ts.internalPrefix, key), value)

	// We return nil error to stay compatible with interface of db.Storage
	return nil
}

// GetRoot returns current Sparse Merkle Tree root
func (ts *TreeStorage) GetRoot(ctx context.Context) (*merkletree.Hash, error) {
	if ts.currentRoot != nil {
		hash := merkletree.Hash{}
		copy(hash[:], ts.currentRoot[:])
		return &hash, nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	value := store.Get(merkletree.Concat(ts.internalPrefix, []byte("root")))
	if value == nil {
		return nil, merkletree.ErrNotFound
	}

	hash, err := merkletree.NewHashFromHex(string(value))
	if err != nil {
		return nil, err
	}
	ts.currentRoot = hash

	return hash, nil
}

// SetRoot updates current Sparse Merkle Tree root
func (ts *TreeStorage) SetRoot(ctx context.Context, hash *merkletree.Hash) error {
	root := &merkletree.Hash{}
	copy(root[:], hash[:])

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	store.Set(merkletree.Concat(ts.internalPrefix, []byte("root")), []byte(root.Hex()))
	ts.currentRoot = root

	// We return nil error to stay compatible with interface of db.Storage
	return nil
}

// Iterate implements the method Iterate of the interface db.Storage
func (ts *TreeStorage) Iterate(ctx context.Context, f func([]byte, *merkletree.Node) (bool, error)) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	iterator := sdk.KVStorePrefixIterator(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	defer closeIteratorOrPanic(iterator)

	for ; iterator.Valid(); iterator.Next() {
		key := merkletree.Clone(bytes.TrimPrefix(iterator.Key(), ts.keyPrefix))
		value := iterator.Value()

		node, err := merkletree.NewNodeFromBytes(value)
		if err != nil {
			return err
		}

		cont, err := f(merkletree.Clone(bytes.TrimPrefix(key, ts.internalPrefix)), node)
		if err != nil {
			return err
		}
		if !cont {
			break
		}
	}

	return nil
}

// List implements the method List of the interface db.Storage
func (ts *TreeStorage) List(ctx context.Context, limit int) ([]merkletree.KV, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	iterator := sdk.KVStorePrefixIterator(sdkCtx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	defer closeIteratorOrPanic(iterator)

	var result []merkletree.KV
	for ; iterator.Valid(); iterator.Next() {
		key := merkletree.Clone(bytes.TrimPrefix(iterator.Key(), ts.keyPrefix))
		value := iterator.Value()

		node, err := merkletree.NewNodeFromBytes(value)
		if err != nil {
			return nil, err
		}

		result = append(result, merkletree.KV{
			K: bytes.TrimPrefix(key, ts.internalPrefix),
			V: *node,
		})

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result, nil
}
