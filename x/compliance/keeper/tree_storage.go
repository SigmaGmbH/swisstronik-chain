package keeper

import (
	"bytes"
	"context"
	"github.com/SigmaGmbH/go-merkletree-sql/v2"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TreeStorage struct {
	keeper         *Keeper
	keyPrefix      []byte
	internalPrefix []byte
	currentRoot    *merkletree.Hash
	ctx            sdk.Context
}

func NewTreeStorage(ctx sdk.Context, k *Keeper, treeKeyPrefix []byte) TreeStorage {
	return TreeStorage{
		keeper:         k,
		keyPrefix:      treeKeyPrefix,
		internalPrefix: []byte{},
		currentRoot:    nil,
		ctx:            ctx,
	}
}

// WithPrefix implements the method WithPrefix of the interface db.Storage
func (ts *TreeStorage) WithPrefix(prefix []byte) merkletree.Storage {
	return &TreeStorage{
		ts.keeper,
		ts.keyPrefix,
		merkletree.Concat(ts.internalPrefix, prefix),
		nil,
		ts.ctx,
	}
}

// Get retrieves a value from a key in the db.Storage
func (ts *TreeStorage) Get(_ context.Context, key []byte) (*merkletree.Node, error) {
	store := prefix.NewStore(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)

	res := store.Get(merkletree.Concat(ts.internalPrefix, key))
	if res == nil {
		return nil, merkletree.ErrNotFound
	}

	return merkletree.NewNodeFromBytes(res)
}

// Put inserts new node into Sparse Merkle Tree
func (ts *TreeStorage) Put(_ context.Context, key []byte, node *merkletree.Node) error {
	store := prefix.NewStore(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	value := node.Value()

	store.Set(merkletree.Concat(ts.internalPrefix, key), value)

	// We return nil error to stay compatible with interface of db.Storage
	return nil
}

// GetRoot returns current Sparse Merkle Tree root
func (ts *TreeStorage) GetRoot(_ context.Context) (*merkletree.Hash, error) {
	if ts.currentRoot != nil {
		hash := merkletree.Hash{}
		copy(hash[:], ts.currentRoot[:])
		return &hash, nil
	}

	store := prefix.NewStore(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
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
func (ts *TreeStorage) SetRoot(_ context.Context, hash *merkletree.Hash) error {
	root := &merkletree.Hash{}
	copy(root[:], hash[:])

	store := prefix.NewStore(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
	store.Set(merkletree.Concat(ts.internalPrefix, []byte("root")), []byte(root.Hex()))
	ts.currentRoot = root

	// We return nil error to stay compatible with interface of db.Storage
	return nil
}

// Iterate implements the method Iterate of the interface db.Storage
func (ts *TreeStorage) Iterate(_ context.Context, f func([]byte, *merkletree.Node) (bool, error)) error {
	iterator := sdk.KVStorePrefixIterator(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
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
func (ts *TreeStorage) List(_ context.Context, limit int) ([]merkletree.KV, error) {
	iterator := sdk.KVStorePrefixIterator(ts.ctx.KVStore(ts.keeper.storeKey), ts.keyPrefix)
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
